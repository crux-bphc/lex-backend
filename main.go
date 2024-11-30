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
	router.Use(cors.New(cors.Config{
		// TODO(release): need to properly configure these before release
		AllowAllOrigins: true,
		AllowHeaders:    []string{"*"},
	}))

	var locationMiddleware gin.HandlerFunc

	if baseUri := os.Getenv("BASE_URI"); len(baseUri) > 0 {
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
