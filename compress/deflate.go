package compress

import (
	"compress/zlib"
	"net/http"
	"netpollmux/header"
)

func NewDeflateWriter(w http.ResponseWriter, r *http.Request) *Compressor {
	return NewZlibWriter(w, r)
}

func NewZlibWriter(w http.ResponseWriter, r *http.Request) *Compressor {
	return newZlibWriter(w, r, false)
}

func newZlibWriter(w http.ResponseWriter, r *http.Request, useBuffer bool) *Compressor {
	c := newCompressor(useBuffer)
	c.w = w
	c.compressType = header.DEFLATE
	if useBuffer {
		c.writer = zlib.NewWriter(c.buf)
	} else {
		c.writer = zlib.NewWriter(w)
	}
	c.ready(w, r)
	return c
}

func Zlib(w http.ResponseWriter, r *http.Request, body []byte, code int) (int, error) {
	gz := newZlibWriter(w, r, true)
	n, err := gz.Write(body)
	w.WriteHeader(code)
	gz.Close()
	return n, err
}

func Deflate(w http.ResponseWriter, r *http.Request, body []byte, code int) (int, error) {
	return Zlib(w, r, body, code)
}
