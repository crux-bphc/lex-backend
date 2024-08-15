package main

import (
	"net/http"

	"github.com/crux-bphc/lex/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	stats "github.com/semihalev/gin-stats"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(location.Default())
	router.Use(stats.RequestStats())

	router.GET("/stats", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, stats.Report())
	})

	router.SetTrustedProxies(nil)

	routes.RegisterImpartusRoutes(router)
	routes.RegisterUserRoutes(router)

	router.Run(":3000")
}
