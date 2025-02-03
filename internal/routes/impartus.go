package routes

import (
	"net/http"
	"strconv"

	"github.com/crux-bphc/lex/internal/auth"
	"github.com/crux-bphc/lex/internal/impartus"
	impartus_routes "github.com/crux-bphc/lex/internal/routes/impartus"
	"github.com/gin-gonic/gin"
)

func RegisterImpartusRoutes(router *gin.Engine) {
	r := router.Group("/impartus")
	r.Use(auth.Middleware())

	// Returns a map of available session ids as [year, sem]
	r.GET("/session", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, impartus.SessionMap)
	})

	// Returns a tuple of a specific session id as [year, sem]
	r.GET("/session/:id", func(ctx *gin.Context) {
		sessionId, _ := strconv.Atoi(ctx.Param("id"))
		ctx.JSON(http.StatusOK, impartus.SessionMap[sessionId])
	})

	impartus_routes.RegisterUserRoutes(r)
	impartus_routes.RegisterSubjectRoutes(r)

	impartus_routes.RegisterVideoRoutes(r)

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
				"cause":   "impartus",
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", data)
	})
}
