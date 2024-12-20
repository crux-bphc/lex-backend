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

type ImpartusClient struct {
	BaseUrl string
}

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

func (client *ImpartusClient) GetSubjects(token string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/subjects", client.BaseUrl), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Gets a list of video from impartus for that specific subject and session
func (client *ImpartusClient) GetVideos(token string, subjectId, sessionId int) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/subjects/%d/lectures/%d", client.BaseUrl, subjectId, sessionId), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("user-agent", fakeuseragent.RandomUserAgent())

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (client *ImpartusClient) GetDecryptionKey(token, ttid string) ([]byte, error) {
	decryptionKeyEndpoint := fmt.Sprintf("%s/fetchvideo/getVideoKey?ttid=%s&keyid=0", client.BaseUrl, ttid)
	req, err := http.NewRequest(http.MethodGet, decryptionKeyEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("user-agent", fakeuseragent.RandomUserAgent())

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Get the data using the Bearer token
	return io.ReadAll(resp.Body)
}

func (client *ImpartusClient) NormalizeDecryptionKey(key []byte) []byte {
	// Do this for some reason, by god thank github for existing impartus downloaders
	data := key[2:]
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

func (client *ImpartusClient) GetIndexM3U8(token, ttid string) ([]byte, error) {
	lectureUrl := fmt.Sprintf("%s/fetchvideo?type=index.m3u8&ttid=%s&token=%s", client.BaseUrl, ttid, token)
	req, err := http.NewRequest(http.MethodGet, lectureUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", fakeuseragent.RandomUserAgent())

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (client *ImpartusClient) GetM3U8Chunk(token, m3u8 string) ([]byte, error) {
	chunkUrl := fmt.Sprintf("%s/fetchvideo?tag=LC&inm3u8=%s", client.BaseUrl, m3u8)
	req, err := http.NewRequest(http.MethodGet, chunkUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", fakeuseragent.RandomUserAgent())

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// the raw m3u8 from impartus
	return io.ReadAll(resp.Body)
}
