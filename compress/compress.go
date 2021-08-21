package compress

import (
	"bytes"
	"net/http"
	"netpollmux/header"
	"sync"
)

var bufPool *sync.Pool

var compressorPool *sync.Pool

func init() {
	bufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
	compressorPool = &sync.Pool{
		New: func() interface{} {
			return &Compressor{}
		},
	}
}

func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufPool.Put(buf)
}

func putCompressor(compressor *Compressor) {
	if compressor.buf != nil {
		putBuffer(compressor.buf)
	}
	compressorPool.Put(compressor)
}

type CompressWriter interface {
	Write(p []byte) (int, error)
	Flush() error
	Close() error
}

type Compressor struct {
	writer       CompressWriter
	w            http.ResponseWriter
	compress     bool
	compressType string
	buf          *bytes.Buffer
	useBuffer    bool
}

func newCompressor(useBuffer bool) *Compressor {
	c := compressorPool.Get().(*Compressor)
	c.useBuffer = useBuffer
	if useBuffer {
		buf := bufPool.Get().(*bytes.Buffer)
		c.buf = buf
	}
	return c
}

func (c *Compressor) ready(w http.ResponseWriter, r *http.Request) {
	if !header.CheckAcceptEncoding(r, c.compressType) {
		c.compress = false
		return
	}
	header.SetContentEncoding(w, c.compressType)
	c.compress = true
}

func (c *Compressor) Write(b []byte) (int, error) {
	if c.compress {
		header.SetHeader(c.w, header.ContentType, http.DetectContentType(b))
		return c.writer.Write(b)
	} else {
		return c.w.Write(b)
	}
}

func (c *Compressor) Close() error {
	defer putCompressor(c)
	if c.compress {
		c.writer.Flush()
		if c.useBuffer {
			c.w.Write(c.buf.Bytes())
		}
		return c.writer.Close()
	}
	return nil
}
