package impartus

import (
	"fmt"
	"io"
	"net/http"
)

type ImpartusClient struct {
	baseUrl string
}

func (client *ImpartusClient) GetSubjects(token string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/subjects", client.baseUrl), nil)
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (client *ImpartusClient) GetDecryptionKey(token string, ttid string) ([]byte, error) {
	decryptionKeyEndpoint := fmt.Sprintf("%s/fetchvideo/getVideoKey?ttid=%s&keyid=0", client.baseUrl, ttid)
	req, err := http.NewRequest(http.MethodGet, decryptionKeyEndpoint, nil)
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

	// Get the data using the Bearer token
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (client *ImpartusClient) NormalizeDecryptionKey(key []byte) []byte {
	// Do this for some reason, by god thank github for existing impartus downloaders
	data := key[2:]
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

func (client *ImpartusClient) GetIndexM3U8(token string, ttid string) ([]byte, error) {
	lectureUrl := fmt.Sprintf("%s/fetchvideo?type=index.m3u8&ttid=%s&token=%s", client.baseUrl, ttid, token)

	resp, err := http.Get(lectureUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (client *ImpartusClient) GetM3U8Chunk(token string, m3u8 string) ([]byte, error) {
	chunkUrl := fmt.Sprintf("%s/fetchvideo?tag=LC&inm3u8=%s", client.baseUrl, m3u8)
	resp, err := http.Get(chunkUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// the raw m3u8 from impartus
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

var Client = ImpartusClient{baseUrl: "https://bitshyd.impartus.com/api"}
