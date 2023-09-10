package functions

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// need to find a better regex for url
var urlRegex = regexp.MustCompile("((http|https)://)(www.)?[a-zA-Z0-9@:%._\\+~#?&//=]{2,256}\\.[a-z]{2,6}\\b([-a-zA-Z0-9@:%._\\+~#?&//=]*)")

func GetLecture(ttid string, token string, baseUrl string) ([]byte, error) {
	lectureUrl := fmt.Sprintf("https://bitshyd.impartus.com/api/fetchvideo?type=index.m3u8&ttid=%s&token=%s", ttid, token)

	resp, err := http.Get(lectureUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data = urlRegex.ReplaceAllFunc(data, func(b []byte) []byte {
		videoUrl, _ := url.ParseQuery(string(b))
		newVideoUrl := fmt.Sprintf("%s/impartus/video?inm3u8=%s&token=%s", baseUrl, videoUrl.Get("inm3u8"), token)
		return []byte(newVideoUrl)
	})

	return data, nil
}
