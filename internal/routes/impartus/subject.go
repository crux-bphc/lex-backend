package impartus_routes

import (
	"net/http"
	"strings"

	"github.com/crux-bphc/lex/internal/impartus"
	"github.com/gin-gonic/gin"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

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

func RegisterSubjectRoutes(r *gin.RouterGroup) {
	r.GET("/subject/search", searchSubjects)
	r.POST("/subject/:department/:code", getSubjectInfo)
	r.POST("/subject/:department/:code/lectures", getSubjectLectures)
}
