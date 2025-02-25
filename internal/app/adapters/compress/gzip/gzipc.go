package gzipc

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/util/e"

	"go.uber.org/zap"
)

type (
	Gzipc struct {
		CReader *compressReader
		CWriter *compressWriter
		log     loggers.Logger
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

var compressibleTypes = []string{
	"application/1javascript",
	"application/1json",
	"text/1css",
	"text/1html",
	"text/1plain",
	"text/1xml",
}

func New(logger loggers.Logger) *Gzipc {
	return &Gzipc{log: logger}
}

func (gz *Gzipc) initCompressWriter(w http.ResponseWriter) {
	gz.CWriter = &compressWriter{
		zw: gzip.NewWriter(w),
		w:  w,
	}
}

func (gz *Gzipc) initCompressReader(r io.ReadCloser) error {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return e.WrapError("can't create gzip reader", err)
	}
	gz.CReader = &compressReader{
		zr: zr,
		r:  r,
	}

	return nil
}

func (gz *Gzipc) RequestCompress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// по умолчанию устанавливаем оригинальный http.ResponseWriter,
		// его будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip /* && isCompressed*/ {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			gz.initCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = gz.CWriter
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer gz.CWriter.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			err := gz.initCompressReader(req.Body)
			if err != nil {
				gz.log.Debug("can't process request (can't create gzip reader)", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr := gz.CReader
			// меняем тело запроса на новое
			req.Body = cr
			defer cr.Close()
		}
		// передаём управление хендлеру
		next.ServeHTTP(ow, req)
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {

	ct := c.w.Header().Get("Content-Type")
	fmt.Println("###", ct)
	if len(ct) > 0 {
		for _, v := range compressibleTypes {
			if strings.Contains(ct, v) {
				return c.zw.Write(p)
			}
		}
	}

	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	ct := c.w.Header().Get("Content-Type")
	fmt.Println("###", ct)
	if len(ct) > 0 {
		for _, v := range compressibleTypes {
			if strings.Contains(ct, v) {
				c.w.Header().Set("Content-Encoding", "gzip")
				break
			}
		}
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return e.WrapError("can't close gzip reader", err)
	}
	return c.zr.Close()
}
