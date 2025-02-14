package logger

import (
	"net/http"
	"time"

	"github.com/d5kx/shorturl/internal/util/e"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// структура для хранения данных об ответе
	responseData struct {
		status int
		size   int
	}
	// собственная реализацию http.ResponseWriter
	logResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

var Log *zap.Logger

func Init(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	ecfg := zap.NewProductionEncoderConfig()
	ecfg.EncodeTime = zapcore.ISO8601TimeEncoder
	ecfg.EncodeDuration = zapcore.StringDurationEncoder

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.Encoding = "console" //json or console
	cfg.DisableCaller = true
	cfg.EncoderConfig = ecfg

	zl, err := cfg.Build()
	if err != nil {
		return e.WrapError("can't create logger configuration", err)
	}
	defer zl.Sync()

	Log = zl
	return nil
}

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := responseData{status: 0, size: 0}
		lw := logResponseWriter{ResponseWriter: w, responseData: &responseData}
		h(&lw, r)

		duration := time.Since(start)

		Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
		)
	}
}

func (r *logResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *logResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
