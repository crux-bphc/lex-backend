package impartus_test

import (
	"bytes"
	"testing"

	"github.com/crux-bphc/lex/internal/impartus"
)

func TestMultipleViews(t *testing.T) {
	// Test for multiple views
	chunk := []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=AES-128,URI="http://bitshyd.impartus.com/api/fetchvideo/getVideoKey?ttid=9072014&keyid=0"
#EXTINF:10.436278,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0000_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0001_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0002_hls_0.ts
#EXT-X-DISCONTINUITY
#EXTINF:10.436278,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v3_0000_hls_0.ts
#EXTINF:10.440000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v3_0001_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v3_0002_hls_0.ts
#EXT-X-ENDLIST
`)

	views := impartus.SplitViews(chunk)

	leftView := []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=AES-128,URI="http://bitshyd.impartus.com/api/fetchvideo/getVideoKey?ttid=9072014&keyid=0"
#EXTINF:10.436278,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0000_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0001_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0002_hls_0.ts
#EXT-X-ENDLIST
`)

	rightView := []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=AES-128,URI="http://bitshyd.impartus.com/api/fetchvideo/getVideoKey?ttid=9072014&keyid=0"
#EXTINF:10.436278,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v3_0000_hls_0.ts
#EXTINF:10.440000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v3_0001_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v3_0002_hls_0.ts
#EXT-X-ENDLIST
`)

	if !bytes.Equal(views.Left, leftView) {
		t.Errorf("Expected left view to be %s, got %s", leftView, views.Left)
	}

	if !bytes.Equal(views.Right, rightView) {
		t.Errorf("Expected right view to be %s, got %s", rightView, views.Right)
	}
}

func TestSingleView(t *testing.T) {
	// Test for single view
	chunk := []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=AES-128,URI="http://bitshyd.impartus.com/api/fetchvideo/getVideoKey?ttid=9072014&keyid=0"
#EXTINF:10.436278,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0000_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0001_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0002_hls_0.ts
#EXT-X-ENDLIST
`)

	views := impartus.SplitViews(chunk)

	leftView := []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=AES-128,URI="http://bitshyd.impartus.com/api/fetchvideo/getVideoKey?ttid=9072014&keyid=0"
#EXTINF:10.436278,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0000_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0001_hls_0.ts
#EXTINF:10.400000,
http://bitshyd.impartus.com/api/fetchvideo?ts=http%3A%2F%2F172.16.3.45%2F%2Fdownload1%2F9072014_hls%2F854x480_30%2F854x480_30v1_0002_hls_0.ts
#EXT-X-ENDLIST
`)

	if !bytes.Equal(views.Left, leftView) {
		t.Errorf("Expected left view to be %s, got %s", leftView, views.Left)
	}

}
