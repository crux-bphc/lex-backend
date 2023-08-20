package functions

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var keyUriRegex = regexp.MustCompile("#EXT-X-KEY:METHOD=AES-128,URI=\".*?ttid=(\\d*)&.*\"")

// TODO
// Need to test if any auth token works, even if the user is not registered to the course
func GetM3U8(inm3u8 string, uri string) ([]byte, error) {
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := keyUriRegex.FindAllSubmatch(data, 1)
	fmt.Printf("%q\n", matches)

	return data, nil
}
