package routes

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/crux-bphc/lex/internal/auth"
	"github.com/crux-bphc/lex/internal/impartus"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/surrealdb/surrealdb.go"
)

// ensures that the user accessing multipartus is still using the same password
// which means that this user's courses are accessible to other users.
func impartusValidJwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)
		impartusJwt, err := surrealdb.SmartUnmarshal[string](
			impartus.Repository.DB.Query(
				"SELECT VALUE fn::get_token(id) FROM ONLY user WHERE email = $email LIMIT 1",
				map[string]interface{}{
					"email": claims.EMail,
				},
			),
		)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(impartusJwt) == 0 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": errors.New("enter correct impartus password to access resource"),
			})
			return
		}

		ctx.Set("IMPARTUS_JWT", impartusJwt)

		ctx.Next()
	}
}

// returns the already fetched impartus jwt token of the user from the database
func getImpartusJwtForUser(ctx *gin.Context) string {
	token, _ := ctx.Get("IMPARTUS_JWT")
	return token.(string)
}

// Map of impartus session id as key and a tuple of [year, sem] as value
var impartusSessionMap = map[int][2]int{
	1426: {2024, 1},
	1369: {2023, 2},
	1339: {2023, 1},
	1275: {2022, 2},
	1249: {2022, 1},
}

func RegisterImpartusRoutes(router *gin.Engine) {
	r := router.Group("/impartus")
	r.Use(auth.Middleware())

	authorized := r.Group("/")
	authorized.Use(impartusValidJwtMiddleware())

	r.GET("/user", func(ctx *gin.Context) {
		// TODO: return a bunch of user info such as number of pinned subjects etc
		// also return if user currently has a valid impartus jwt

		claims := auth.GetClaims(ctx)
		registered, err := surrealdb.SmartUnmarshal[bool](
			impartus.Repository.DB.Query(
				"count(SELECT id FROM user WHERE email = $email) == 1",
				map[string]interface{}{
					"email": claims.EMail,
				},
			),
		)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"registered": registered,
			// TODO: add number of subjects/lectures added to database to response
		})
	})

	// Creates a new entry for the user in the database.
	r.POST("/user", func(ctx *gin.Context) {
		body := struct {
			Password string `json:"password" binding:"required"`
		}{}

		if err := ctx.BindJSON(&body); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		claims := auth.GetClaims(ctx)

		impartusToken, err := impartus.Client.GetToken(claims.EMail, body.Password)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		user := impartus.User{
			EMail:     claims.EMail,
			Password:  body.Password,
			Jwt:       impartusToken,
			UpdatedAt: time.Now(),
		}

		// create user in database
		_, err = impartus.Repository.DB.Create("user", user)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Registered",
			// TODO: add number of subjects/lectures added to database to response
		})
	})

	// Returns a map of available session ids as [year, sem]
	r.GET("/session", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, impartusSessionMap)
	})

	// Returns a tuple of a specific session id as [year, sem]
	r.GET("/session/:id", func(ctx *gin.Context) {
		sessionId, _ := strconv.Atoi(ctx.Param("id"))
		ctx.JSON(http.StatusOK, impartusSessionMap[sessionId])
	})

	r.GET("/subject", func(ctx *gin.Context) {
		// TODO: subject search endpoint
		// for now it just returns all the subjects
		subjects, err := impartus.Repository.GetSubjects()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, subjects)
	})

	// Returns a list of all the valid lecture sections for the particular subject
	r.GET("/subject/:department/:code", func(ctx *gin.Context) {
		// CS/ECE/EEE/INSTR becomes CS,ECE,EEE,INSTR in the URL
		department := strings.ReplaceAll(ctx.Param("department"), ",", "/")

		subjectCode := ctx.Param("code")

		subjectId := fmt.Sprintf("subject:['%s','%s']", department, subjectCode)

		lectures, err := impartus.Repository.GetLectures(subjectId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, lectures)
	})

	// Returns the list of subjects the user has pinned
	r.GET("/user/subjects", func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)
		subjects, err := impartus.Repository.GetPinnedSubjects(claims.EMail)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, subjects)
	})

	modifyPinnedSubjects := func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)

		subjectId := ctx.Query("id")
		if len(subjectId) == 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "please provide a valid 'id' parameter for your subject",
			})
			return
		}

		var err error

		switch ctx.Request.Method {
		case http.MethodPost:
			_, err = impartus.Repository.DB.Query(`LET $user = (SELECT VALUE id FROM user WHERE email = $email);RELATE ONLY $user->pinned->$subjectId`, map[string]interface{}{
				"email":     claims.EMail,
				"subjectId": subjectId,
			})
		case http.MethodDelete:
			_, err = impartus.Repository.DB.Query(`LET $user = (SELECT VALUE id FROM user WHERE email = $email);DELETE $user->bought WHERE out=$subjectId`, map[string]interface{}{
				"email":     claims.EMail,
				"subjectId": subjectId,
			})
		}

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("action on %s successful", subjectId),
		})
	}

	// Add a subject to the user's pinned section
	r.POST("/user/subjects", modifyPinnedSubjects)

	// Remove a subject from the user's pinned section
	r.DELETE("/user/subjects", modifyPinnedSubjects)

	// Returns a list of videos from the lecture using a registered user's impartus jwt token
	r.GET("/lecture/:sessionId/:subjectId", func(ctx *gin.Context) {
		sessionId := ctx.Param("sessionId")
		subjectId := ctx.Param("subjectId")
		lectureId := fmt.Sprintf("lecture:[%s,%s]", sessionId, subjectId)

		impartusToken, err := impartus.Repository.GetLectureToken(lectureId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		data, err := impartus.Client.GetVideos(impartusToken, subjectId, sessionId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", data)
	})

	// Returns the decryption key for the particular video without an Authorization header
	authorized.GET("/video/:ttid/key", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := getImpartusJwtForUser(ctx)

		data, err := impartus.Client.GetDecryptionKey(token, ttid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		data = impartus.Client.NormalizeDecryptionKey(data)

		ctx.Data(http.StatusOK, "application/pgp-keys", data)
	})

	m3u8Regex := regexp.MustCompile("http.*inm3u8=(.*)")

	// Gets a video stream
	authorized.GET("/video/:ttid/m3u8", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := getImpartusJwtForUser(ctx)

		hostUrl := location.Get(ctx).String()

		data, err := impartus.Client.GetIndexM3U8(token, ttid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		data = m3u8Regex.ReplaceAll(data, []byte(fmt.Sprintf("%s/impartus/chunk/m3u8?m3u8=$1", hostUrl)))

		ctx.Data(http.StatusOK, "application/x-mpegurl", data)
	})

	cipherUriRegex := regexp.MustCompile(`URI=".*ttid=(\d*)&.*"`)

	// Direct link to the m3u8 file with the uri of the decryption key for the AES-128 cipher
	// replaced by the server implementation
	authorized.GET("/chunk/m3u8", func(ctx *gin.Context) {
		m3u8 := ctx.Query("m3u8")
		token := getImpartusJwtForUser(ctx)

		hostUrl := location.Get(ctx).String()

		data, err := impartus.Client.GetM3U8Chunk(token, m3u8)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		decryptionKeyUrl := fmt.Sprintf(`URI="%s/impartus/video/$1/key"`, hostUrl)
		data = cipherUriRegex.ReplaceAll(data, []byte(decryptionKeyUrl))
		ctx.Data(http.StatusOK, "application/x-mpegurl", data)
	})
}
