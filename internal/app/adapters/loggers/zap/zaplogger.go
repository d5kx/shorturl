package zaplogger

import (
	"net/http"
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

func New() (*ZapLogger, error) {
	log := ZapLogger{zap: zap.NewNop()}
	err := log.init(conf.GetLoggerLevel())
	return &log, err
}

func (z *ZapLogger) init(level string) error {
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
		return e.WrapError("can't create loggers configuration", err)
	}
	defer zl.Sync()
	z.zap = zl
	return nil
}

func (z *ZapLogger) RequestLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := responseData{status: 0, size: 0}
		lw := logResponseWriter{ResponseWriter: w, responseData: &responseData}
		next.ServeHTTP(&lw, r)

		duration := time.Since(start)
		z.zap.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("Content-type", r.Header.Get("Content-type")),
			zap.String("Accept-Encoding", r.Header.Get("Accept-Encoding")),
			zap.String("Content-Encoding", r.Header.Get("Content-Encoding")),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
		)
	}
}

func (z *ZapLogger) Info(msg string, fields ...any) {
	z.zap.Info(msg, z.anyToZapFields(fields)...)
}

func (z *ZapLogger) Fatal(msg string, fields ...any) {
	z.zap.Fatal(msg, z.anyToZapFields(fields)...)
}

func (z *ZapLogger) Debug(msg string, fields ...any) {
	z.zap.Debug(msg, z.anyToZapFields(fields)...)
}

func (z *ZapLogger) anyToZapFields(fields []any) []zapcore.Field {
	var f []zapcore.Field
	for _, v := range fields {
		if field, ok := v.(zapcore.Field); ok {
			f = append(f, field)
		}
	}
	return f
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
