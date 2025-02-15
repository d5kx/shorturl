package logger

import "net/http"

type Logger interface {
	Init(string) error
	RequestLogging(http.HandlerFunc) http.HandlerFunc
}
