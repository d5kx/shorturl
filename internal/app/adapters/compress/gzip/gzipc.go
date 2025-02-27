package gzipc

import (
	"compress/gzip"
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
		w               http.ResponseWriter
		zw              *gzip.Writer
		needCompression bool
	}
)

var compressibleTypes = []string{
	"application/json",
	"text/html",
}

func New(logger loggers.Logger) *Gzipc {
	return &Gzipc{
		log:     logger,
		CReader: nil,
		CWriter: nil,
	}
}

func (gz *Gzipc) initCompressWriter(w http.ResponseWriter) {

	if gz.CWriter != nil {
		gz.CWriter.zw.Reset(w)
	} else {
		gz.CWriter = &compressWriter{zw: gzip.NewWriter(w)}
	}
	gz.CWriter.w = w
	gz.CWriter.needCompression = false
}

func (gz *Gzipc) initCompressReader(r io.ReadCloser) error {
	if gz.CReader != nil {
		err := gz.CReader.zr.Reset(r)
		if err != nil {
			return e.WrapError("can't reset gzip reader", err)
		}
	} else {
		zr, err := gzip.NewReader(r)
		if err != nil {
			return e.WrapError("can't create gzip reader", err)
		}
		gz.CReader = &compressReader{zr: zr}
	}
	gz.CReader.r = r

	return nil
}

func (gz *Gzipc) RequestCompress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter,
		// его будем передавать следующей функции
		ow := w
		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		supportsGzip := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			gz.initCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = gz.CWriter
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer gz.CWriter.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		sendsGzip := strings.Contains(req.Header.Get("Content-Encoding"), "gzip")
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
		// передаём управление следующему хендлеру
		next.ServeHTTP(ow, req)
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.needCompression || c.isСompressibleType() {
		c.needCompression = true
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.needCompression || c.isСompressibleType() {
		c.needCompression = true
		c.w.Header().Set("Content-Encoding", "gzip")
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if c.needCompression {
		return c.zw.Close()
	}
	return nil
}

func (c *compressWriter) isСompressibleType() bool {
	ct := c.w.Header().Get("Content-Type")
	if len(ct) > 0 {
		for _, v := range compressibleTypes {
			if strings.Contains(ct, v) {
				return true
			}
		}
	}
	return false
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
