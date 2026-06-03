package compress

import (
	"compress/gzip"
	"io"
)

func NewGzipReader(r io.Reader) io.ReadCloser {
	pr, pw := io.Pipe()

	go func() {
		gw := gzip.NewWriter(pw)
		_, err := io.Copy(gw, r)
		if err != nil {
			gw.Close()
			pw.CloseWithError(err)
			return
		}
		if err := gw.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()

	return pr
}
