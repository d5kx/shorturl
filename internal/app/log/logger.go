package logger

import "net/http"

type Logger interface {
	RequestLogging(http.HandlerFunc) http.HandlerFunc
	Info(string, ...any)
	Fatal(string, ...any)
}
