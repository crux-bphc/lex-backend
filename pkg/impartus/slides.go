package pkg

import (
	"encoding/json"
	"fmt"
	"strings"
)

type slide struct {
	ID    int    `json:"id"`
	Url   string `json:"url"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Fetches slides based on videoId and returns them in a nicer format
func (client *ImpartusClient) GetSlides(token, videoId string) ([]slide, error) {
	data, err := client.GetVideoInfo(token, videoId)
	if err != nil {
		return nil, err
	}

	var rawSlidesData struct {
		Slides []struct {
			StartPoint int    `json:"timepoint"`
			EndPoint   int    `json:"end_point"`
			FileID     string `json:"fileid"`
			EmbedID    int    `json:"embed_id"`
			SlideID    int    `json:"slideId"`
		} `json:"slides"`
	}

	if err := json.Unmarshal(data, &rawSlidesData); err != nil {
		return nil, err
	}

	slides := make([]slide, len(rawSlidesData.Slides))

	for i, rawSlide := range rawSlidesData.Slides {
		slides[i] = slide{
			ID:    rawSlide.SlideID,
			Start: rawSlide.StartPoint,
			End:   rawSlide.EndPoint,
			Url:   fmt.Sprintf("%s/download1/embedded/%s/img_%d.png", strings.TrimSuffix(client.BaseUrl, "/api"), rawSlide.FileID, rawSlide.EmbedID),
		}
	}

	return slides, nil
}
