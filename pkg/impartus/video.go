package pkg

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/iunary/fakeuseragent"
)

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
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Get the video info based on videoId from impartus
func (client *ImpartusClient) GetVideoInfo(token, videoId string) ([]byte, error) {
	videoUrl := fmt.Sprintf("%s/videos/params/id/%s", client.BaseUrl, videoId)
	req, err := http.NewRequest(http.MethodGet, videoUrl, nil)
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
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Get the video info based on ttid from impartus
func (client *ImpartusClient) GetTTIDInfo(token, ttid string) ([]byte, error) {
	videoUrl := fmt.Sprintf("%s/lectures/%s/info", client.BaseUrl, ttid)
	req, err := http.NewRequest(http.MethodGet, videoUrl, nil)
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
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Get the main m3u8 file containing videos with different resolutions
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
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Get m3u8 file from impartus
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
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code: " + resp.Status)
	}

	// the raw m3u8 from impartus
	return io.ReadAll(resp.Body)
}
