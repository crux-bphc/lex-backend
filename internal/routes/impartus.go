package routes

import (
	"encoding/json"
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
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// ensures that the user accessing multipartus is still using the same password
// which means that this user's courses are accessible to other users.
func impartusValidJwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)
		impartusJwtResult, err := surrealdb.Query[string](
			impartus.Repository.DB,
			"RETURN fn::get_token($user)",
			map[string]interface{}{
				"user": models.RecordID{
					Table: "user",
					ID:    claims.EMail,
				},
			},
		)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		impartusJwt := (*impartusJwtResult)[0].Result

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
		registered, err := surrealdb.Query[struct {
			Registered bool `json:"registered"`
			Valid      bool `json:"valid"`
		}](
			impartus.Repository.DB,
			"RETURN {registered: record::exists($user), valid: type::is::string(fn::get_token($user))}",
			map[string]interface{}{
				"user": models.RecordID{
					Table: "user",
					ID:    claims.EMail,
				},
			},
		)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "db-query",
			})
			return
		}

		ctx.JSON(http.StatusOK, (*registered)[0].Result)
	})

	// Creates a new entry for the user in the database.
	r.POST("/user", func(ctx *gin.Context) {
		body := struct {
			Password string `json:"password" binding:"required"`
		}{}

		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
				"code":    "invalid-body",
			})
			return
		}

		claims := auth.GetClaims(ctx)

		impartusToken, err := impartus.Client.GetToken(claims.EMail, body.Password)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
				"code":    "get-token",
			})
			return
		}

		user := impartus.User{
			ID: &models.RecordID{
				Table: "user",
				ID:    claims.EMail,
			},
			EMail:     claims.EMail,
			Password:  body.Password,
			Jwt:       impartusToken,
			UpdatedAt: time.Now(),
		}

		// create user in database
		_, err = surrealdb.Create[any](impartus.Repository.DB, models.Table("user"), user)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "db-create",
			})
			return
		}

		_, err = surrealdb.Query[any](impartus.Repository.DB, "fn::extract_lectures($user)", map[string]interface{}{
			"user": (*user.ID),
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "extract-lectures",
			})
			return
		}

		_, err = surrealdb.Query[any](impartus.Repository.DB, "fn::pin_registered($user)", map[string]interface{}{
			"user": (*user.ID),
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "pin-lectures",
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

	r.GET("/subject/search", func(ctx *gin.Context) {
		query := ctx.Query("q")

		subjects, err := impartus.Repository.GetSubjects(query)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-subjects",
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

		data, err := surrealdb.Query[struct {
			Subject  impartus.Subject   `json:"subject,omitempty"`
			Lectures []impartus.Lecture `json:"lectures,omitempty"`
		}](
			impartus.Repository.DB,
			"SELECT (SELECT * FROM ONLY $parent) as subject, (SELECT * OMIT users FROM lecture WHERE subject = $parent.id) as lectures FROM ONLY $subject",
			map[string]interface{}{
				"subject": models.RecordID{Table: "subject", ID: []string{department, subjectCode}},
			},
		)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-lectures",
			})
			return
		}

		ctx.JSON(http.StatusOK, (*data)[0].Result)
	})

	// Returns the list of subjects the user has pinned
	r.GET("/user/subjects", func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)
		subjects, err := impartus.Repository.GetPinnedSubjects(claims.EMail)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-pinned-subjects",
			})
			return
		}

		ctx.JSON(http.StatusOK, subjects)
	})

	modifyPinnedSubjects := func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)

		body := struct {
			Department string `json:"department" binding:"required"`
			Code       string `json:"code" binding:"required"`
		}{}
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
				"code":    "invalid-body",
			})
			return
		}

		var err error
		vars := map[string]interface{}{
			"user":    models.RecordID{Table: "user", ID: claims.EMail},
			"subject": models.RecordID{Table: "subject", ID: []string{body.Department, body.Code}},
		}

		switch ctx.Request.Method {
		case http.MethodPost:
			_, err = surrealdb.Query[any](impartus.Repository.DB, "RELATE ONLY $user->pinned->$subject", vars)
		case http.MethodDelete:
			_, err = surrealdb.Query[any](impartus.Repository.DB, "DELETE $user->pinned WHERE out=$subject", vars)
		}

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "db-query",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("action on %s %s successful", body.Department, body.Code),
		})
	}

	// Add a subject to the user's pinned section
	r.POST("/user/subjects", modifyPinnedSubjects)

	// Remove a subject from the user's pinned section
	r.DELETE("/user/subjects", modifyPinnedSubjects)

	// Returns a list of videos from the lecture using a registered user's impartus jwt token
	r.GET("/lecture/:sessionId/:subjectId", func(ctx *gin.Context) {
		sessionId, _ := strconv.Atoi(ctx.Param("sessionId"))
		subjectId, _ := strconv.Atoi(ctx.Param("subjectId"))

		impartusToken, err := impartus.Repository.GetLectureToken(sessionId, subjectId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-lecture-token",
			})
			return
		}

		data, err := impartus.Client.GetVideos(impartusToken, subjectId, sessionId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-videos",
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", data)
	})

	// Returns video info based on videoId
	authorized.GET("/video/:videoId/info", func(ctx *gin.Context) {
		videoId := ctx.Param("videoId")
		token := getImpartusJwtForUser(ctx)

		data, err := impartus.Client.GetVideoInfo(token, videoId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-video-info",
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", data)
	})

	// Returns list of slide image urls for the given ttid
	authorized.GET("/ttid/:ttid/slides", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := getImpartusJwtForUser(ctx)

		data, err := impartus.Client.GetTTIDInfo(token, ttid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-ttid-info",
			})
			return
		}

		var rawData struct {
			SessionId int `json:"sessionId"`
			SubjectId int `json:"subjectId"`
			VideoId   int `json:"videoId"`
		}
		if err := json.Unmarshal(data, &rawData); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "unmarshal-ttid-data",
			})
			return
		}

		impartusToken, err := impartus.Repository.GetLectureToken(rawData.SessionId, rawData.SubjectId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-lecture-token",
			})
			return
		}

		data, err = impartus.Client.GetVideoInfo(impartusToken, strconv.Itoa(rawData.VideoId))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-video-info",
			})
			return
		}

		var rawSlidesData struct {
			Slides []struct {
				StartPoint int    `json:"timepoint"`
				EndPoint   int    `json:"end_point"`
				FileID     string `json:"fileid"`
				EmbedID    int    `json:"embed_id"`
				SlideID    int    `json:"slideId"`
			} `json:"slides"`
		}

		if err := json.Unmarshal(data, &rawSlidesData); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "unmarshal-slides-data",
			})
			return
		}

		slides := make([]struct {
			ID    int    `json:"id"`
			Url   string `json:"url"`
			Start int    `json:"start"`
			End   int    `json:"end"`
		}, len(rawSlidesData.Slides))

		for i, slide := range rawSlidesData.Slides {
			slides[i].ID = slide.SlideID
			slides[i].Start = slide.StartPoint
			slides[i].End = slide.EndPoint
			slides[i].Url = fmt.Sprintf("%s/download1/embedded/%s/img_%d.png", strings.TrimSuffix(impartus.Client.BaseUrl, "/api"), slide.FileID, slide.EmbedID)
		}

		ctx.JSON(http.StatusOK, slides)
	})

	// Returns video info based on ttid
	authorized.GET("/ttid/:ttid/info", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := getImpartusJwtForUser(ctx)

		data, err := impartus.Client.GetTTIDInfo(token, ttid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-video-info",
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", data)
	})

	// Returns the decryption key for the particular video without an Authorization header
	authorized.GET("/ttid/:ttid/key", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := getImpartusJwtForUser(ctx)

		data, err := impartus.Client.GetDecryptionKey(token, ttid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-decryption-key",
			})
			return
		}

		data = impartus.Client.NormalizeDecryptionKey(data)

		ctx.Data(http.StatusOK, "application/pgp-keys", data)
	})

	m3u8Regex := regexp.MustCompile("http.*inm3u8=(.*)")

	// Gets a video stream
	authorized.GET("/ttid/:ttid/m3u8", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := getImpartusJwtForUser(ctx)

		hostUrl := location.Get(ctx).String()

		data, err := impartus.Client.GetIndexM3U8(token, ttid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
				"code":    "get-index-m3u8",
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
				"code":    "get-m3u8-chunk",
			})
			return
		}

		decryptionKeyUrl := fmt.Sprintf(`URI="%s/impartus/ttid/$1/key"`, hostUrl)
		data = cipherUriRegex.ReplaceAll(data, []byte(decryptionKeyUrl))
		ctx.Data(http.StatusOK, "application/x-mpegurl", data)
	})
}
