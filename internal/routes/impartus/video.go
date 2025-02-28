package impartus_routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/crux-bphc/lex/internal/impartus"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

// Returns video info based on videoId
func getVideoInfo(ctx *gin.Context) {
	videoId := ctx.Param("videoId")
	token := impartus.GetImpartusJwtForUser(ctx)

	data, err := impartus.Client.GetVideoInfo(token, videoId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-video-info",
			"cause":   "impartus",
		})
		return
	}

	ctx.Data(http.StatusOK, "application/json", data)
}

// Returns video info based on ttid
func getTTIDInfo(ctx *gin.Context) {
	ttid := ctx.Param("ttid")
	token := impartus.GetImpartusJwtForUser(ctx)

	data, err := impartus.Client.GetTTIDInfo(token, ttid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-video-info",
			"cause":   "impartus",
		})
		return
	}

	ctx.Data(http.StatusOK, "application/json", data)
}

// Returns list of slide image urls for the given ttid
func getSlides(ctx *gin.Context) {
	ttid := ctx.Param("ttid")
	token := impartus.GetImpartusJwtForUser(ctx)

	data, err := impartus.Client.GetTTIDInfo(token, ttid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-ttid-info",
			"cause":   "impartus",
		})
		return
	}

	var rawData struct {
		SessionId int `json:"sessionId"`
		SubjectId int `json:"subjectId"`
		VideoId   int `json:"videoId"`
	}
	if err := json.Unmarshal(data, &rawData); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "unmarshal-ttid-data",
		})
		return
	}

	impartusToken, err := impartus.Repository.GetLectureToken(rawData.SessionId, rawData.SubjectId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-lecture-token",
		})
		return
	}

	slides, err := impartus.Client.GetSlides(impartusToken, strconv.Itoa(rawData.VideoId))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-slides",
			"cause":   "impartus",
		})
		return
	}

	ctx.JSON(http.StatusOK, slides)
}

// Returns the auto-generated PDF of the slides for the given ttid
func downloadSlides(ctx *gin.Context) {
	ttid := ctx.Param("ttid")
	token := impartus.GetImpartusJwtForUser(ctx)

	data, err := impartus.Client.GetTTIDInfo(token, ttid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-ttid-info",
			"cause":   "impartus",
		})
		return
	}

	var rawData struct {
		SubjectName string `json:"subjectName"`
		SessionId   int    `json:"sessionId"`
		SubjectId   int    `json:"subjectId"`
		VideoId     int    `json:"videoId"`
	}
	if err := json.Unmarshal(data, &rawData); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "unmarshal-ttid-data",
		})
		return
	}

	impartusToken, err := impartus.Repository.GetLectureToken(rawData.SessionId, rawData.SubjectId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-lecture-token",
		})
		return
	}

	pdf, err := impartus.Client.GetSlidesPDF(impartusToken, strconv.Itoa(rawData.VideoId))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-slides-pdf",
			"cause":   "impartus",
		})
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", rawData.SubjectName))
	ctx.Data(http.StatusOK, "application/pdf", pdf)
}

// Returns the decryption key for the particular video without an Authorization header
func getTTIDdecryptionKey(ctx *gin.Context) {
	ttid := ctx.Param("ttid")
	token := impartus.GetImpartusJwtForUser(ctx)

	data, err := impartus.Client.GetDecryptionKey(token, ttid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-decryption-key",
			"cause":   "impartus",
		})
		return
	}

	data = impartus.Client.NormalizeDecryptionKey(data)

	ctx.Data(http.StatusOK, "application/pgp-keys", data)
}

var m3u8Regex = regexp.MustCompile("http.*inm3u8=(.*)")

// The index m3u8 file containing streams of different video resolutions
func getIndexM3U8(ctx *gin.Context) {
	ttid := ctx.Param("ttid")
	token := impartus.GetImpartusJwtForUser(ctx)

	hostUrl := location.Get(ctx).String()

	data, err := impartus.Client.GetIndexM3U8(token, ttid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-index-m3u8",
			"cause":   "impartus",
		})
		return
	}
	data = m3u8Regex.ReplaceAll(data, []byte(fmt.Sprintf("%s/impartus/chunk/m3u8?m3u8=$1", hostUrl)))

	ctx.Data(http.StatusOK, "application/x-mpegurl", data)
}

func getM3U8ChunkInfo(ctx *gin.Context) {
	ttid := ctx.Param("ttid")
	token := impartus.GetImpartusJwtForUser(ctx)

	data, err := impartus.Client.GetIndexM3U8(token, ttid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-index-m3u8",
			"cause":   "impartus",
		})
		return
	}
	indexM3U8 := string(data)
	tracks := m3u8Regex.FindAllStringSubmatch(indexM3U8, -1)

	// get data from the first track
	data, err = impartus.Client.GetM3U8Chunk(token, tracks[0][1])
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-m3u8-chunk",
			"cause":   "impartus",
		})
		return
	}

	hasRightView := bytes.Contains(data, []byte("#EXT-X-DISCONTINUITY"))

	ctx.JSON(http.StatusOK, gin.H{
		"tracks": tracks,
		"views": gin.H{
			"left":  true,
			"right": hasRightView,
		},
	})
}

var cipherUriRegex = regexp.MustCompile(`URI=".*ttid=(\d*)&.*"`)

// Replace decryption key URI in the chunk with the server implementation
func transformChunk(hostUrl string, chunk []byte) []byte {
	decryptionKeyUrl := fmt.Sprintf(`URI="%s/impartus/ttid/$1/key"`, hostUrl)
	return cipherUriRegex.ReplaceAll(chunk, []byte(decryptionKeyUrl))
}

// Direct link to the m3u8 file with the uri of the decryption key for the AES-128 cipher
// replaced by the server implementation
func getM3U8Chunk(ctx *gin.Context) {
	m3u8 := ctx.Query("m3u8")
	token := impartus.GetImpartusJwtForUser(ctx)
	hostUrl := location.Get(ctx).String()

	data, err := impartus.Client.GetM3U8Chunk(token, m3u8)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-m3u8-chunk",
			"cause":   "impartus",
		})
		return
	}

	data = transformChunk(hostUrl, data)
	ctx.Data(http.StatusOK, "application/x-mpegurl", data)
}

// Returns the left view of the video
func getLeftView(ctx *gin.Context) {
	m3u8 := ctx.Query("m3u8")
	token := impartus.GetImpartusJwtForUser(ctx)
	hostUrl := location.Get(ctx).String()

	data, err := impartus.Client.GetM3U8Chunk(token, m3u8)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-m3u8-chunk",
			"cause":   "impartus",
		})
		return
	}

	views := impartus.SplitViews(data)
	data = transformChunk(hostUrl, views.Left)
	ctx.Data(http.StatusOK, "application/x-mpegurl", data)
}

// Returns the right view of the video
func getRightView(ctx *gin.Context) {
	m3u8 := ctx.Query("m3u8")
	token := impartus.GetImpartusJwtForUser(ctx)
	hostUrl := location.Get(ctx).String()

	data, err := impartus.Client.GetM3U8Chunk(token, m3u8)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-m3u8-chunk",
			"cause":   "impartus",
		})
		return
	}

	views := impartus.SplitViews(data)
	data = transformChunk(hostUrl, views.Right)
	ctx.Data(http.StatusOK, "application/x-mpegurl", data)
}

func RegisterVideoRoutes(r *gin.RouterGroup) {
	authorized := r.Group("/")
	authorized.Use(impartus.ValidJwtMiddleware)

	authorized.GET("/video/:videoId/info", getVideoInfo)
	authorized.GET("/ttid/:ttid/info", getTTIDInfo)

	authorized.GET("/ttid/:ttid/slides", getSlides)
	authorized.GET("/ttid/:ttid/slides/download", downloadSlides)

	authorized.GET("/ttid/:ttid/key", getTTIDdecryptionKey)
	authorized.GET("/ttid/:ttid/m3u8", getIndexM3U8)
	authorized.GET("/ttid/:ttid/m3u8/info", getM3U8ChunkInfo)

	authorized.GET("/chunk/m3u8", getM3U8Chunk)
	authorized.GET("/chunk/m3u8/left", getLeftView)
	authorized.GET("/chunk/m3u8/right", getRightView)
}
