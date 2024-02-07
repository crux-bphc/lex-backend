package main

import (
	"github.com/crux-bphc/lex/impartus"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(location.Default())

	impartus.HandleImpartus(router)

	router.Run(":3000")
}
