package impartus_routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/crux-bphc/lex/internal/impartus"
	"github.com/gin-gonic/gin"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// Returns a list of subjects based on the query
func searchSubjects(ctx *gin.Context) {
	query := ctx.Query("q")

	subjects, err := impartus.Repository.GetSubjects(query)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-subjects",
		})
		return
	}

	ctx.JSON(http.StatusOK, subjects)
}

// Returns information about a particular subject
func getSubjectInfo(ctx *gin.Context) {
	// CS/ECE/EEE/INSTR becomes CS,ECE,EEE,INSTR in the URL
	department := strings.ReplaceAll(ctx.Param("department"), ",", "/")
	subjectCode := ctx.Param("code")

	subject, err := surrealdb.Select[impartus.Subject](
		impartus.Repository.DB,
		models.RecordID{Table: "subject", ID: []string{department, subjectCode}},
	)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "select-subject",
		})
		return
	}

	ctx.JSON(http.StatusOK, subject)
}

// Returns a list of all the valid lecture sections for the particular subject
func getSubjectLectures(ctx *gin.Context) {
	// CS/ECE/EEE/INSTR becomes CS,ECE,EEE,INSTR in the URL
	department := strings.ReplaceAll(ctx.Param("department"), ",", "/")
	subjectCode := ctx.Param("code")

	lectures, err := surrealdb.Query[[]impartus.Lecture](
		impartus.Repository.DB,
		"SELECT * OMIT users FROM lecture WHERE subject=$subject ORDER BY impartus_session DESC",
		map[string]interface{}{
			"subject": models.RecordID{Table: "subject", ID: []string{department, subjectCode}},
		},
	)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-lectures",
		})
		return
	}

	ctx.JSON(http.StatusOK, (*lectures)[0].Result)
}

// Returns a list of videos from the lecture using a registered user's impartus jwt token
func getLectureVideos(ctx *gin.Context) {
	sessionId, _ := strconv.Atoi(ctx.Param("sessionId"))
	subjectId, _ := strconv.Atoi(ctx.Param("subjectId"))

	impartusToken, err := impartus.Repository.GetLectureToken(sessionId, subjectId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-lecture-token",
		})
		return
	}

	data, err := impartus.Client.GetVideos(impartusToken, subjectId, sessionId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-videos",
			"cause":   "impartus",
		})
		return
	}

	ctx.Data(http.StatusOK, "application/json", data)
}

func RegisterSubjectRoutes(r *gin.RouterGroup) {
	r.GET("/subject/search", searchSubjects)
	r.GET("/subject/:department/:code", getSubjectInfo)
	r.GET("/subject/:department/:code/lectures", getSubjectLectures)
	r.GET("/lecture/:sessionId/:subjectId", getLectureVideos)
}
