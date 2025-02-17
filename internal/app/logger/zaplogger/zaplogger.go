package zaplogger

import (
	"net/http"
	"sync"
	"time"

	"github.com/d5kx/shorturl/internal/app/conf"

	"github.com/d5kx/shorturl/internal/util/e"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	ZapLogger struct {
		zap *zap.Logger
	}
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

var (
	log  = ZapLogger{zap: zap.NewNop()}
	once sync.Once
)

func GetInstance() (ZapLogger, error) {
	var err error = nil
	once.Do(func() {
		err = zapLoggerInit(conf.GetLoggerLevel())
	})

	return log, err
}

func (z ZapLogger) Zap() *zap.Logger { return z.zap }

func zapLoggerInit(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.Encoding = "console" //json or console
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	zl, err := cfg.Build()
	if err != nil {
		return e.WrapError("can't create logger configuration", err)
	}
	defer zl.Sync()
	log.zap = zl
	return nil
}

func (z ZapLogger) RequestLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := responseData{status: 0, size: 0}
		lw := logResponseWriter{ResponseWriter: w, responseData: &responseData}
		h(&lw, r)

		duration := time.Since(start)
		z.zap.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
		)
	}
}
func (z ZapLogger) Info(msg string, fields ...any) {
	var f []zapcore.Field

	for _, v := range fields {
		if field, ok := v.(zapcore.Field); ok {
			f = append(f, field)
		}
	}
	z.zap.Info(msg, f...)
}
func (z ZapLogger) Fatal(msg string, fields ...any) {
	var f []zapcore.Field

	for _, v := range fields {
		if field, ok := v.(zapcore.Field); ok {
			f = append(f, field)
		}
	}
	z.zap.Fatal(msg, f...)
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
