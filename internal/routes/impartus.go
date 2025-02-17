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
}
