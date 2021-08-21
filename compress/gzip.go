package compress

import (
	"compress/gzip"
	"net/http"
	"netpollmux/header"
)

func NewGzipWriter(w http.ResponseWriter, r *http.Request) *Compressor {
	return newGzipWriter(w, r, false)
}

func newGzipWriter(w http.ResponseWriter, r *http.Request, useBuffer bool) *Compressor {
	c := newCompressor(useBuffer)
	c.compressType = header.GZIP
	c.w = w
	if useBuffer {
		c.writer = gzip.NewWriter(c.buf)
	} else {
		c.writer = gzip.NewWriter(w)
	}
	c.ready(w, r)
	return c
}

func Gzip(w http.ResponseWriter, r *http.Request, body []byte, code int) (int, error) {
	gz := newGzipWriter(w, r, true)
	n, err := gz.Write(body)
	w.WriteHeader(code)
	gz.Close()
	return n, err
}
