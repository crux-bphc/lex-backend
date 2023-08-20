package main

import (
	"fmt"
	"log"

	"github.com/crux-bphc/lex/functions"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	// Returns the decryption key without the need for a Authorization header
	r.GET("/impartus/key/:ttid", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		data, err := functions.GetDecryptionKey(ttid, token)
		if err != nil {
			log.Println(err)
		}

		ctx.Data(200, "application/pgp-keys", data)
	})

	// Direct link to the m3u8 file with the uri of the decryption key for the AES-128 cipher replaced by the server implementation
	r.GET("/impartus/video", func(ctx *gin.Context) {
		inm3u8 := ctx.Query("inm3u8")
		// token := ctx.Query("token")

		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}
		data, err := functions.GetM3U8(inm3u8, fmt.Sprintf("%s://%s/key", scheme, ctx.Request.Host))
		if err != nil {
			log.Println(err)
		}

		ctx.Data(200, "application/x-mpegurl", data)
	})

	r.Run(":3000")
}
