package pkg

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/iunary/fakeuseragent"
)

// Gets the AES decryption key for the video using the given token
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
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code: " + resp.Status)
	}

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
