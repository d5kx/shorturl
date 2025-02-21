package basehandler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d5kx/shorturl/internal/app/loggers/mock"

	"github.com/d5kx/shorturl/internal/app/adapters/storages/mock"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/stretchr/testify/assert"
)

// go test -v -count 1 -coverprofile="cover.txt"
// go tool cover -html=cover.txt
func Test_methodPost(t *testing.T) {
	conf.ParseFlags()
	ml := mocklogger.New()
	p := New(mockstor.New(), ml)

	testCases := []struct {
		name                string
		path                string
		method              string
		contentType         string
		body                string
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "POST: valid request",
			path:                "/",
			method:              http.MethodPost,
			contentType:         "text/plain",
			body:                "http://ya.ru",
			expectedCode:        http.StatusCreated,
			expectedContentType: "text/plain",
			expectedBody:        conf.GetResURLAdr() + "/AbCdEf",
		},
		{
			name:                "POST: wrong Content-Type",
			path:                "/",
			method:              http.MethodPost,
			contentType:         "text/json",
			body:                "http://ya.ru",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
		{
			name:                "POST: no link in the request body",
			path:                "/",
			method:              http.MethodPost,
			contentType:         "text/plain",
			body:                "",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
		{
			name:                "POST: db error emulation",
			path:                "/",
			method:              http.MethodPost,
			contentType:         "text/plain",
			body:                "db_error",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			body := bytes.NewBuffer([]byte(tc.body))
			r := httptest.NewRequest(tc.method, tc.path, body)
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			p.Post(w, r)

			b := make([]byte, w.Body.Len())
			w.Body.Read(b)
			var sb strings.Builder
			sb.Write(b)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedContentType, w.Header().Get("Content-Type"), "ContentType не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedBody, sb.String(), "Тело ответа не совпадает с ожидаемым")
		})
	}
}

func Test_methodPostApiShorten(t *testing.T) {
	ml := mocklogger.New()
	p := New(mockstor.New(), ml)
	testCases := []struct {
		name                string
		path                string
		method              string
		contentType         string
		body                string
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "POST: api/json valid request",
			path:                "/api/shorten",
			method:              http.MethodPost,
			contentType:         "application/json",
			body:                `{"url":"https://practicum.yandex.ru"}`,
			expectedCode:        http.StatusCreated,
			expectedContentType: "application/json",
			expectedBody:        `{"result":"` + conf.GetResURLAdr() + `/AbCdEf"` + `}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			body := bytes.NewBuffer([]byte(tc.body))
			r := httptest.NewRequest(tc.method, tc.path, body)
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			p.PostAPIShorten(w, r)

			b := make([]byte, w.Body.Len())
			w.Body.Read(b)
			var sb strings.Builder
			sb.Write(b)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedContentType, w.Header().Get("Content-Type"), "ContentType не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedBody, sb.String(), "Тело ответа не совпадает с ожидаемым")
		})
	}
}
func Test_methodGetHandleFunc(t *testing.T) {
	ml := mocklogger.New()
	p := New(mockstor.New(), ml)
	//conf.ParseFlags()
	testCases := []struct {
		name             string
		path             string
		method           string
		contentType      string
		body             string
		expectedCode     int
		expectedLocation string
	}{
		{
			name:             "GET: valid request",
			path:             "/AbCdEf",
			method:           http.MethodGet,
			contentType:      "text/plain",
			body:             "",
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "http://ya.ru",
		},
		{
			name:             "GET: non-existent short link",
			path:             "/ZbCdEf",
			method:           http.MethodGet,
			contentType:      "text/plain",
			body:             "",
			expectedCode:     http.StatusBadRequest,
			expectedLocation: "",
		},
		{
			name:             "GET: short link missing in request",
			path:             "/",
			method:           http.MethodGet,
			contentType:      "text/plain",
			body:             "",
			expectedCode:     http.StatusBadRequest,
			expectedLocation: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			body := bytes.NewBuffer([]byte(tc.body))
			r := httptest.NewRequest(tc.method, tc.path, body)
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			p.Get(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedLocation, w.Header().Get("Location"), "Адрес переадресации не совпадает с ожидаемым")

		})
	}
}
