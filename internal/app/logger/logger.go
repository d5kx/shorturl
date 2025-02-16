package logger

import "net/http"

type Logger interface {
	//Init(string) error
	RequestLogging(http.HandlerFunc) http.HandlerFunc
	Info(string, ...any)
	Fatal(string, ...any)
}
