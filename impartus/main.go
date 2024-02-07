package impartus

import (
	"log"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

var baseImpartusUrl = "https://bitshyd.impartus.com/api"

func HandleImpartus(router *gin.Engine) {
	r := router.Group("/impartus")

	// Returns the decryption key without the need for a Authorization header
	r.GET("/lecture/:ttid/key", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		data, err := GetDecryptionKey(ttid, token)
		if err != nil {
			log.Println(err)
		}

		ctx.Data(200, "application/pgp-keys", []byte(data))
	})

	// Gets a video stream based on the internet connection and bandwidth of the user.
	r.GET("/lecture/:ttid/m3u8", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		hostUrl := location.Get(ctx).String()

		data, err := GetLecture(ttid, token, hostUrl)
		if err != nil {
			log.Println(err)
		}

		ctx.Data(200, "application/x-mpegurl", data)
	})

	// Direct link to the m3u8 file with the uri of the decryption key for the AES-128 cipher replaced by the server implementation
	r.GET("/chunk/m3u8", func(ctx *gin.Context) {
		m3u8 := ctx.Query("m3u8")
		token := ctx.Query("token")

		hostUrl := location.Get(ctx).String()

		data, err := GetM3U8Chunk(m3u8, token, hostUrl)
		if err != nil {
			log.Println(err)
		}
		ctx.Data(200, "application/x-mpegurl", data)
	})
}
