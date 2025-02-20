package mocklogger

import (
	"net/http"
)

type (
	MockLogger struct {
	}
)

var (
	loggerInstance MockLogger
)

func GetInstance() MockLogger {
	return loggerInstance
}

func (m MockLogger) RequestLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r)
	}
}

func (m MockLogger) Info(msg string, fields ...any) {

}

func (m MockLogger) Fatal(msg string, fields ...any) {

}

func (m MockLogger) Debug(msg string, fields ...any) {

}
