package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/crux-bphc/lex/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	stats "github.com/semihalev/gin-stats"
)

func main() {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// CORS setup
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	if allowedOrigins := os.Getenv("CORS_ORIGINS"); len(allowedOrigins) > 0 {
		corsConfig.AllowOrigins = strings.Split(allowedOrigins, ",")
	} else {
		corsConfig.AllowAllOrigins = true
	}

	router.Use(cors.New(corsConfig))

	// Location middleware
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

	// Stats middleware
	router.Use(stats.RequestStats())
	router.GET("/stats", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, stats.Report())
	})

	router.SetTrustedProxies(nil)

	// Multipartus
	routes.RegisterImpartusRoutes(router)

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Lex backend is online!")
	})

	router.Run(":3000")
}
