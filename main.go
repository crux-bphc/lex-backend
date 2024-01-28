package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/crux-bphc/lex/functions"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	r := gin.Default()
	// r.SetTrustedProxies(nil)
	r.Use(cors.Default())

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

		// Any auth token works, even if the user is not registered to the course
		token := ctx.Query("token")

		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}
		replacement := fmt.Sprintf("%s://%s/impartus/key/$1?token=%s", scheme, ctx.Request.Host, token)
		data, err := functions.GetM3U8(inm3u8, replacement)
		if err != nil {
			log.Println(err)
		}

		ctx.Data(200, "application/x-mpegurl", data)
	})

	// Gets a video stream based on the internet connection and bandwidth of the user.
	r.GET("/impartus/lecture/:ttid", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}
		baseUrl := fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)
		data, err := functions.GetLecture(ttid, token, baseUrl)
		if err != nil {
			log.Println(err)
		}

		ctx.Data(200, "application/x-mpegurl", data)
	})

	r.Run(":3000")
}
