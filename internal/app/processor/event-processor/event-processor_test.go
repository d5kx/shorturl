package eventprocessor

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/d5kx/shorturl/internal/app/storage/mock"
)

func Test_Process(t *testing.T) {
	p := New(mockstorage.New())
	p.AddAddress("localhost:8080")
	testCases := []struct {
		name         string
		path         string
		method       string
		contentType  string
		body         string
		expectedCode int
	}{
		{
			name:         "CONNECT request",
			path:         "/",
			method:       http.MethodConnect,
			contentType:  "text/plain",
			body:         "http://ya.ru",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "POST request",
			path:         "/",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "http://ya.ru",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "GET request",
			path:         "/AbCdEf",
			method:       http.MethodGet,
			contentType:  "text/plain",
			body:         "",
			expectedCode: http.StatusTemporaryRedirect,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			body := bytes.NewBuffer([]byte(tc.body))
			r := httptest.NewRequest(tc.method, tc.path, body)
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			p.Process(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}

}
func Test_methodPostHandleFunc(t *testing.T) {
	p := New(mockstorage.New())
	p.AddAddress("localhost:8080")

	testCases := []struct {
		name                string
		method              string
		contentType         string
		body                string
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "POST: valid request",
			method:              http.MethodPost,
			contentType:         "text/plain",
			body:                "http://ya.ru",
			expectedCode:        http.StatusCreated,
			expectedContentType: "text/plain",
			expectedBody:        "http://localhost:8080/AbCdEf",
		},
		{
			name:                "POST: wrong Content-Type",
			method:              http.MethodPost,
			contentType:         "text/json",
			body:                "http://ya.ru",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
		{
			name:                "POST: no link in the request body",
			method:              http.MethodPost,
			contentType:         "text/plain",
			body:                "",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			body := bytes.NewBuffer([]byte(tc.body))
			r := httptest.NewRequest(tc.method, "/", body)
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			p.methodPostHandleFunc(w, r)

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

}
