package keycloak

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
	// TODO: add more user attributes
	BITS_ID string `json:"bits_id"`
}

func init() {
	// The JWKS for keycloak
	URL := "https://auth.crux-bphc.com/realms/CRUx/protocol/openid-connect/certs"

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

	// Type cast to add custom claims
	if claims, ok := token.Claims.(*UserClaims); ok {
		return claims, nil
	}

	return nil, errors.New("cannot process claims")
}

func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := getBearerToken(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}

		claims, err := parseToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
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
