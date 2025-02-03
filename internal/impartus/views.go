package impartus

import (
	"bufio"
	"bytes"
)

type views struct {
	Left  []byte
	Right []byte
}

func SplitViews(chunk []byte) views {
	scanner := bufio.NewScanner(bytes.NewReader(chunk))
	scanner.Split(bufio.ScanLines)
	var left bytes.Buffer
	var right bytes.Buffer

	type WriteMode int
	const (
		Header WriteMode = iota
		Left
		Right
	)

	var mode WriteMode = Header

	for scanner.Scan() {
		line := scanner.Bytes()

		if mode == Header && bytes.HasPrefix(line, []byte("#EXTINF")) {
			mode = Left
		} else if bytes.HasPrefix(line, []byte("#EXT-X-DISCONTINUITY")) {
			mode = Right
			continue
		}

		switch mode {
		case Header:
			left.Write(line)
			left.WriteByte('\n')
			right.Write(line)
			right.WriteByte('\n')
		case Left:
			left.Write(line)
			left.WriteByte('\n')
		case Right:
			right.Write(line)
			right.WriteByte('\n')
		}
	}

	if mode == Right {
		left.WriteString("#EXT-X-ENDLIST\n")
	}

	return views{Left: left.Bytes(), Right: right.Bytes()}
}
