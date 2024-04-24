package routes

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/crux-bphc/lex/internal/impartus"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

func registerUser(ctx *gin.Context) {
	type requestBody struct {
		Password string `json:"password" binding:"required"`
		Token    string `json:"token" binding:"required"`
	}
	var body requestBody
	ctx.BindJSON(&body)

	// TODO: validate received impartus JWT

	user := impartus.User{
		// TODO: get email from keycloak
		EMail:    "f2022xxxx@hyderabad.bits-pilani.ac.in",
		Password: body.Password,
		Jwt:      body.Token,
	}

	// create user in database
	_, err := impartus.Repository.DB.Create("user", &user)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Registered",
		// TODO: add number of subjects/lectures added to database to response
	})
}

func RegisterImpartusRoutes(router *gin.Engine) {
	r := router.Group("/impartus")

	r.POST("/register", registerUser)

	// Returns the decryption key without the need for a Authorization header
	r.GET("/lecture/:ttid/key", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		data, err := impartus.Client.GetDecryptionKey(token, ttid)
		data = impartus.Client.NormalizeDecryptionKey(data)
		if err != nil {
			log.Println(err)
		}

		ctx.Data(200, "application/pgp-keys", []byte(data))
	})

	m3u8Regex := regexp.MustCompile("http.*inm3u8=(.*)")

	// Gets a video stream
	r.GET("/lecture/:ttid/m3u8", func(ctx *gin.Context) {
		ttid := ctx.Param("ttid")
		token := ctx.Query("token")

		hostUrl := location.Get(ctx).String()

		data, err := impartus.Client.GetIndexM3U8(token, ttid)
		if err != nil {
			log.Println(err)
		}
		data = m3u8Regex.ReplaceAll(data, []byte(fmt.Sprintf("%s/impartus/chunk/m3u8?m3u8=$1&token=%s", hostUrl, token)))

		ctx.Data(200, "application/x-mpegurl", data)
	})

	cipherUriRegex := regexp.MustCompile(`URI=".*ttid=(\d*)&.*"`)

	// Direct link to the m3u8 file with the uri of the decryption key for the AES-128 cipher replaced by the server implementation
	r.GET("/chunk/m3u8", func(ctx *gin.Context) {
		m3u8 := ctx.Query("m3u8")
		token := ctx.Query("token")

		hostUrl := location.Get(ctx).String()

		data, err := impartus.Client.GetM3U8Chunk(token, m3u8)
		if err != nil {
			log.Println(err)
		}

		decryptionKeyUrl := fmt.Sprintf(`URI="%s/impartus/lecture/$1/key?token=%s"`, hostUrl, token)
		data = cipherUriRegex.ReplaceAll(data, []byte(decryptionKeyUrl))
		ctx.Data(200, "application/x-mpegurl", data)
	})
}
