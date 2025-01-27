package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/iunary/fakeuseragent"
)

// Returns the impartus jwt token which is active for 7 days
func (client *ImpartusClient) GetToken(username, password string) (string, error) {
	postBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/auth/signin", client.BaseUrl), bytes.NewBuffer(postBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	// need to set user-agent otherwise impartus throws a 403 forbidden
	req.Header.Set("user-agent", fakeuseragent.RandomUserAgent())

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("invalid status code: " + resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	responseBody := struct {
		Token string `json:"token"`
	}{}

	if err := json.Unmarshal(data, &responseBody); err != nil {
		return "", err
	}

	if len(responseBody.Token) == 0 {
		return "", errors.New("failed to retrieve a token, probably because of wrong password")
	}

	return responseBody.Token, nil
}

// Check whether the impartus JWT token is currently valid or not
func (client *ImpartusClient) VerifyToken(token string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/profile", client.BaseUrl), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == 200, nil
}
