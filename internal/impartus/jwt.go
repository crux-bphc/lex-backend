package impartus

import (
	"errors"
	"net/http"

	"github.com/crux-bphc/lex/internal/auth"
	"github.com/gin-gonic/gin"
	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// ensures that the user accessing multipartus is still using the same password
// which means that this user's courses are accessible to other users.
func ValidJwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := auth.GetClaims(ctx)
		impartusJwtResult, err := surrealdb.Query[string](
			Repository.DB,
			"RETURN fn::get_token($user)",
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
			})
			return
		}

		impartusJwt := (*impartusJwtResult)[0].Result

		if len(impartusJwt) == 0 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": errors.New("enter correct impartus password to access resource"),
			})
			return
		}

		ctx.Set("IMPARTUS_JWT", impartusJwt)

		ctx.Next()
	}
}

// returns the already fetched impartus jwt token of the user from the database
func GetImpartusJwtForUser(ctx *gin.Context) string {
	token, _ := ctx.Get("IMPARTUS_JWT")
	return token.(string)
}
