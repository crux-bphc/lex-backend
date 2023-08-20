package functions

import (
	"fmt"
	"io"
	"net/http"
)

// Gets the decryption key for the AES-128 cipher used for encryption by impartus
func GetDecryptionKey(ttid string, token string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://bitshyd.impartus.com/api/fetchvideo/getVideoKey?ttid=%s&keyid=0", ttid), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Get the data using the Bearer token
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Do this for some reason, by god thank github for existing impartus downloaders
	data = data[2:]
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return data, nil
}
