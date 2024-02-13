package main

import (
	"log"

	"github.com/crux-bphc/lex/internal"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(location.Default())

	internal.RegisterImpartusRoutes(router)
	internal.RegisterAuthRoutes(router)

	router.Run(":3000")
}
