package impartus

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var m3u8Regex = regexp.MustCompile("http.*inm3u8=(.*)")

func GetLecture(ttid string, token string, hostUrl string) ([]byte, error) {
	lectureUrl := fmt.Sprintf("%s/fetchvideo?type=index.m3u8&ttid=%s&token=%s", baseImpartusUrl, ttid, token)

	resp, err := http.Get(lectureUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data = m3u8Regex.ReplaceAll(data, []byte(fmt.Sprintf("%s/impartus/chunk/m3u8?m3u8=$1&token=%s", hostUrl, token)))

	return data, nil
}
