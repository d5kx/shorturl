package gzipc

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/d5kx/shorturl/internal/util/e"
)

type (
	Gzipc struct {
		CReader *compressReader
		CWriter *compressWriter
		a       int
	}

	compressReader struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}

	compressWriter struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}
)

func New() *Gzipc {
	Gzipc
	return &Gzipc{}
}

func (gz *Gzipc) initCompressWriter(w http.ResponseWriter) {
	gz.CWriter.zw = gzip.NewWriter(w)
	gz.CWriter.w = w
}

func (gz *Gzipc) initCompressReader(r io.ReadCloser) error {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return e.WrapError("can't create gzip reader", err)
	}
	gz.CReader.r = r
	gz.CReader.zr = zr

	return nil
}

func (gz *Gzipc) RequestCompress(http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		gz.initCompressWriter(w)

		err := gz.initCompressReader(req.Body)
		if err != nil {
			//добавить лог
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}
