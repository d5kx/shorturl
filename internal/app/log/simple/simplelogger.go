package simplelogger

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/d5kx/shorturl/internal/app/conf"
)

type (
	SimpleLogger struct {
		simple *log.Logger
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
	loggerInstance SimpleLogger
	once           sync.Once
)

func GetInstance() SimpleLogger {
	once.Do(func() {
		loggerInstance.simple = log.New(os.Stdout, strings.ToUpper(conf.GetLoggerLevel())+" ", log.LstdFlags|log.Lmicroseconds)
	})
	return loggerInstance
}

func (s SimpleLogger) RequestLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := responseData{status: 0, size: 0}
		lw := logResponseWriter{ResponseWriter: w, responseData: &responseData}
		h(&lw, r)

		duration := time.Since(start)
		s.Info("got incoming HTTP request",
			"uri:", r.RequestURI,
			"method:", r.Method,
			"status:", responseData.status,
			"size:", responseData.size,
			"duration", duration.String(),
		)
	}
}

func (s SimpleLogger) Info(msg string, fields ...any) {
	s.simple.Println(msg, fields)
}

func (s SimpleLogger) Fatal(msg string, fields ...any) {
	s.simple.Fatal(msg, fields)
}

func (s SimpleLogger) Debug(msg string, fields ...any) {
	s.simple.Println(msg, fields)
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
