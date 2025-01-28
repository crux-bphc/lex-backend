package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/crux-bphc/lex/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	stats "github.com/semihalev/gin-stats"
)

func main() {
	router := gin.Default()

	// TODO(release): need to properly configure these before release
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"*"}

	router.Use(cors.New(corsConfig))

	var locationMiddleware gin.HandlerFunc

	if baseUri := os.Getenv("PUBLIC_URI"); len(baseUri) > 0 {
		u, err := url.Parse(baseUri)
		if err != nil {
			log.Fatalln(err)
		}

		locationMiddleware = location.New(location.Config{
			Scheme: u.Scheme,
			Host:   u.Host,
			Base:   u.Path,
		})
	} else {
		locationMiddleware = location.Default()
	}

	router.Use(locationMiddleware)

	router.Use(stats.RequestStats())

	router.GET("/stats", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, stats.Report())
	})

	router.SetTrustedProxies(nil)

	routes.RegisterImpartusRoutes(router)
	routes.RegisterUserRoutes(router)

	router.Run(":3000")
}
