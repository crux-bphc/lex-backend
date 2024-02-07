package impartus

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// regex for finding the AES cipher key uri
var cipherUriRegex = regexp.MustCompile(`URI=".*ttid=(\d*)&.*"`)

// Gets the bytes of the m3u8 file with the decryption key replaced
func GetM3U8Chunk(m3u8 string, token string, hostUrl string) ([]byte, error) {
	url := fmt.Sprintf("%s/fetchvideo?tag=LC&inm3u8=%s", baseImpartusUrl, m3u8)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// the raw m3u8 from impartus
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	decryptionKeyUrl := fmt.Sprintf(`URI="%s/impartus/lecture/$1/key?token=%s"`, hostUrl, token)
	data = cipherUriRegex.ReplaceAll(data, []byte(decryptionKeyUrl))

	// Find the lecture ttid and replace it with a custom url
	return data, nil
}
