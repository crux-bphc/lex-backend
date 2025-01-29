package impartus

import (
	"io"
	"net/http"

	"github.com/iunary/fakeuseragent"
	"github.com/signintech/gopdf"
)

func WriteImagesToPDF(imageUrls []string, writer io.Writer) (int64, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4Landscape})
	httpClient := &http.Client{}
	for _, imageUrl := range imageUrls {
		req, err := http.NewRequest(http.MethodGet, imageUrl, nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("user-agent", fakeuseragent.RandomUserAgent())

		resp, err := httpClient.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		pdf.AddPage()
		imageHolder, err := gopdf.ImageHolderByReader(resp.Body)
		if err != nil {
			return 0, err
		}
		pdf.ImageByHolder(imageHolder, 0, 0, nil)
	}

	return pdf.WriteTo(writer)
}
