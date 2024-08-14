package routes

import (
	"github.com/crux-bphc/lex/internal/auth"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(base *gin.RouterGroup) {
	r := base.Group("/user")

	r.Use(auth.Middleware())

	// TODO: rudimentary endpoint for testing
	r.GET("/info", func(ctx *gin.Context) {

		ctx.JSON(200, gin.H{
			"claims": auth.GetClaims(ctx),
		})
	})
}
