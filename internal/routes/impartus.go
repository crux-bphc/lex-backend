package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/crux-bphc/lex/internal/auth"
	"github.com/crux-bphc/lex/internal/impartus"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

func RegisterImpartusRoutes(router *gin.Engine) {
	r := router.Group("/impartus")

	authorized := r.Group("/")
	authorized.Use(auth.Middleware())

	authorized.GET("/user", func(ctx *gin.Context) {
		// TODO: return a bunch of user info such as number of pinned subjects etc
		// also return if user currently has a valid impartus jwt
	})

	// Creates a new entry for the user in the database.
	authorized.POST("/user", func(ctx *gin.Context) {
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

		ctx.JSON(200, gin.H{
			"message": "Registered",
			// TODO: add number of subjects/lectures added to database to response
		})
	})

	// Returns a list of all the valid lecture sections for the particular subject
	r.GET("/subject", func(ctx *gin.Context) {
		// this is the id of the subject stored in the database
		subjectId := ctx.Query("id")
		if len(subjectId) == 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "please provide a valid 'id' parameter for your subject",
			})
			return
		}

		lectures, err := impartus.Repository.GetLectures(subjectId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(200, gin.H{
			"count":    len(lectures),
			"sections": lectures,
		})
	})

	r.GET("/subject/search", func(ctx *gin.Context) {
		// TODO: subject search endpoint
	})

	// Returns the list of subjects the user has pinned
	authorized.GET("/user/subjects", func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)
		subjects, err := impartus.Repository.GetPinnedSubjects(claims.EMail)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(200, gin.H{
			"count":    len(subjects),
			"subjects": subjects,
		})
	})

	// Add a subject to the user's pinned section
	authorized.POST("/user/subjects", func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)

		subjectId := ctx.Query("id")
		if len(subjectId) == 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "please provide a valid 'id' parameter for your subject",
			})
			return
		}

		_, err := impartus.Repository.DB.Query(`LET $user = (SELECT VALUE id FROM user WHERE email = $email);RELATE ONLY $user->pinned->$subjectId`, map[string]interface{}{
			"email":     claims.EMail,
			"subjectId": subjectId,
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": fmt.Sprintf("pinned %s", subjectId),
		})
	})

	// Remove a subject from the user's pinned section
	authorized.DELETE("/user/subjects", func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)

		subjectId := ctx.Query("id")
		if len(subjectId) == 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "please provide a valid 'id' parameter for your subject",
			})
			return
		}

		_, err := impartus.Repository.DB.Query(`LET $user = (SELECT VALUE id FROM user WHERE email = $email);DELETE $user->bought WHERE out=$subjectId`, map[string]interface{}{
			"email":     claims.EMail,
			"subjectId": subjectId,
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": fmt.Sprintf("unpinned %s", subjectId),
		})
	})

	r.GET("/session/:sessionId/:subjectId", func(ctx *gin.Context) {
		// TODO: return list of lectures from the the specific lecture section using
		// the registered user's impartus jwt token
	})

	// Returns the decryption key for the particular video without an Authorization header
	r.GET("/video/:ttid/key", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		data, err := impartus.Client.GetDecryptionKey(token, ttid)
		data = impartus.Client.NormalizeDecryptionKey(data)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.Data(200, "application/pgp-keys", []byte(data))
	})

	m3u8Regex := regexp.MustCompile("http.*inm3u8=(.*)")

	// Gets a video stream
	r.GET("/video/:ttid/m3u8", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		hostUrl := location.Get(ctx).String()

		data, err := impartus.Client.GetIndexM3U8(token, ttid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		data = m3u8Regex.ReplaceAll(data, []byte(fmt.Sprintf("%s/impartus/chunk/m3u8?m3u8=$1&token=%s", hostUrl, token)))

		ctx.Data(200, "application/x-mpegurl", data)
	})

	cipherUriRegex := regexp.MustCompile(`URI=".*ttid=(\d*)&.*"`)

	/*
		Direct link to the m3u8 file with the uri of the decryption key for the AES-128 cipher
		replaced by the server implementation
	*/
	r.GET("/chunk/m3u8", func(ctx *gin.Context) {
		m3u8 := ctx.Query("m3u8")
		token := ctx.Query("token")

		hostUrl := location.Get(ctx).String()

		data, err := impartus.Client.GetM3U8Chunk(token, m3u8)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		decryptionKeyUrl := fmt.Sprintf(`URI="%s/impartus/video/$1/key?token=%s"`, hostUrl, token)
		data = cipherUriRegex.ReplaceAll(data, []byte(decryptionKeyUrl))
		ctx.Data(200, "application/x-mpegurl", data)
	})
}
