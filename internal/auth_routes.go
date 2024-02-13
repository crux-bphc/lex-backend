package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func RegisterAuthRoutes(router *gin.Engine) {
	r := router.Group("auth")

	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = os.Getenv("GIN_MODE") == "release"

	gothic.Store = store

	googleProvider := google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), os.Getenv("ORIGIN")+"/auth/callback/google", "email", "profile")
	goth.UseProviders(googleProvider)

	r.GET("/google", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", "google")
		c.Request.URL.RawQuery = q.Encode()
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	r.GET("/callback/google", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", "google")
		c.Request.URL.RawQuery = q.Encode()
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		res, err := json.Marshal(user)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		jsonString := string(res)
		c.Data(http.StatusOK, "application/json", []byte(jsonString))
	})

	r.GET("/exchange", func(c *gin.Context) {
		url_str := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", os.Getenv("KEYCLOAK_URL"), os.Getenv("KEYCLOAK_REALM"))

		form_data := url.Values{}
		form_data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
		form_data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
		form_data.Set("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
		form_data.Set("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))
		form_data.Set("subject_token", c.Query("token"))
		form_data.Set("subject_issuer", "google")

		req, err := http.NewRequest(http.MethodPost, url_str, strings.NewReader(form_data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Data(resp.StatusCode, "application/json", []byte(data))
	})

}
