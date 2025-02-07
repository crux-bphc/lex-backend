package impartus_routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/crux-bphc/lex/internal/auth"
	"github.com/crux-bphc/lex/internal/impartus"
	"github.com/gin-gonic/gin"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

func getUserInfo(ctx *gin.Context) {
	// TODO: return a bunch of user info such as number of pinned subjects etc
	// also return if user currently has a valid impartus jwt

	claims := auth.GetClaims(ctx)
	registered, err := surrealdb.Query[struct {
		Registered bool `json:"registered"`
		Valid      bool `json:"valid"`
	}](
		impartus.Repository.DB,
		"RETURN {registered: record::exists($user), valid: type::is::string(fn::get_token($user))}",
		map[string]interface{}{
			"user": models.RecordID{
				Table: "user",
				ID:    claims.EMail,
			},
		},
	)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "db-query",
		})
		return
	}

	ctx.JSON(http.StatusOK, (*registered)[0].Result)
}

// Creates a new entry for the user in the database.
func registerUser(ctx *gin.Context) {
	body := struct {
		Password string `json:"password" binding:"required"`
	}{}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    "invalid-body",
		})
		return
	}

	claims := auth.GetClaims(ctx)

	impartusToken, err := impartus.Client.GetToken(claims.EMail, body.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    "get-token",
			"cause":   "impartus",
		})
		return
	}

	user := impartus.User{
		ID: &models.RecordID{
			Table: "user",
			ID:    claims.EMail,
		},
		EMail:     claims.EMail,
		Password:  body.Password,
		Jwt:       impartusToken,
		UpdatedAt: time.Now(),
	}

	// create user in database
	_, err = surrealdb.Create[any](impartus.Repository.DB, models.Table("user"), user)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "db-create",
		})
		return
	}

	// no of total lectures the user is registered to
	lectures, err := surrealdb.Query[int](impartus.Repository.DB, "fn::extract_lectures($user)", map[string]interface{}{
		"user": (*user.ID),
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "extract-lectures",
		})
		return
	}

	// no of pinned subjects of the user
	pinned, err := surrealdb.Query[int](impartus.Repository.DB, "fn::pin_registered($user)", map[string]interface{}{
		"user": (*user.ID),
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "pin-lectures",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Registered",
		"lectures": (*lectures)[0].Result,
		"pinned":   (*pinned)[0].Result,
	})
}

// Returns the list of subjects the user has pinned
func getPinnedSubjects(ctx *gin.Context) {
	claims := auth.GetClaims(ctx)
	subjects, err := impartus.Repository.GetPinnedSubjects(claims.EMail)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "get-pinned-subjects",
		})
		return
	}

	ctx.JSON(http.StatusOK, subjects)
}

// Adds or removes a subject from the user's pinned subjects
func modifyPinnedSubjects(ctx *gin.Context) {
	claims := auth.GetClaims(ctx)

	body := struct {
		Department string `json:"department" binding:"required"`
		Code       string `json:"code" binding:"required"`
	}{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    "invalid-body",
		})
		return
	}

	var err error
	vars := map[string]interface{}{
		"user":    models.RecordID{Table: "user", ID: claims.EMail},
		"subject": models.RecordID{Table: "subject", ID: []string{body.Department, body.Code}},
	}

	switch ctx.Request.Method {
	case http.MethodPost:
		_, err = surrealdb.Query[any](impartus.Repository.DB, "RELATE ONLY $user->pinned->$subject", vars)
	case http.MethodDelete:
		_, err = surrealdb.Query[any](impartus.Repository.DB, "DELETE $user->pinned WHERE out=$subject", vars)
	}

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"code":    "db-query",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("action on %s %s successful", body.Department, body.Code),
	})
}

func RegisterUserRoutes(r *gin.RouterGroup) {
	r.GET("/user", getUserInfo)
	r.POST("/user", registerUser)
	r.GET("/user/subjects", getPinnedSubjects)
	r.POST("/user/subjects", modifyPinnedSubjects)
	r.DELETE("/user/subjects", modifyPinnedSubjects)
}
