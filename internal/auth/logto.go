package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwks keyfunc.Keyfunc

type UserClaims struct {
	jwt.RegisteredClaims
	EMail string `json:"email"`
}

func init() {
	// The JWKS for logto
	URL := "https://logto.local.crux-bphc.com/oidc/jwks"

	var err error
	jwks, err = keyfunc.NewDefault([]string{URL})

	if err != nil {
		log.Fatalf("Failed to create JWK Set from resource at the given URL.\nError: %s", err)
	}
}

func getBearerToken(ctx *gin.Context) (string, error) {
	header := ctx.Request.Header.Get("authorization")
	if header == "" {
		return "", errors.New("authorization header is empty")
	}

	// split the authorization into Bearer + token
	authHeaderArray := strings.Split(header, " ")
	if len(authHeaderArray) != 2 {
		return "", errors.New("incorrectly formatted authorization header")
	}

	return authHeaderArray[1], nil
}

// checks if the jwt token is obtained from the correct logto client
func checkAudience(claims jwt.ClaimStrings) bool {
	validAudiences := []string{
		"n3h4pbp70440nj5h9ofph", // Lex
		"yjmouftg5ba37lf70ooas", // Multipartus Downloader
	}

	for _, aud := range claims {
		for _, valid := range validAudiences {
			if aud == valid {
				return true
			}
		}
	}

	return false
}

func parseToken(jwtToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &UserClaims{}, jwks.Keyfunc)

	// Parse the JWT.
	if err != nil {
		return nil, err
	}

	// Check if the token is valid.
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	if aud, err := token.Claims.GetAudience(); err != nil {
		return nil, err
	} else if !checkAudience(aud) {
		return nil, errors.New("invalid audience")
	}

	// Type cast to add custom claims
	if claims, ok := token.Claims.(*UserClaims); ok {
		return claims, nil
	}

	return nil, errors.New("cannot process claims")
}

// Middleware for checking Logto JWT tokens
func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var token string

		// Use the url query if available, otherwise use the auth header
		if token = ctx.Query("token"); len(token) == 0 {
			var err error
			token, err = getBearerToken(ctx)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
					"code":    "get-bearer-token",
				})
				return
			}
		}

		claims, err := parseToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
				"code":    "parse-token",
			})
			return
		}

		ctx.Set("CLAIMS", claims)

		ctx.Next()
	}
}

func GetClaims(ctx *gin.Context) *UserClaims {
	claims, exists := ctx.Get("CLAIMS")
	if !exists {
		log.Fatalln("'CLAIMS' is not present in Gin context, use the middleware to set it")
		return nil
	}

	return claims.(*UserClaims)
}
