package functions

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// regex for finding the AES cipher key uri
// TODO: maybe find a better regex
var keyUriRegex = regexp.MustCompile("#EXT-X-KEY:METHOD=AES-128,URI=\".*ttid=(\\d*)&.*\"")

// Gets the bytes of the m3u8 file with the decryption key replaced
func GetM3U8(inm3u8 string, repl string) ([]byte, error) {
	url := fmt.Sprintf("https://bitshyd.impartus.com/api/fetchvideo?tag=LC&inm3u8=%s", inm3u8)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// the raw m3u8 from impartus
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Find the lecture ttid and replace it with a custom url
	data = keyUriRegex.ReplaceAll(data, []byte("#EXT-X-KEY:METHOD=AES-128,URI=\""+repl+"\""))

	return data, nil
}
