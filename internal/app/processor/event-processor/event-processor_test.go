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

func TestMethodPostHandleFunc(t *testing.T) {

	p := New(mockstorage.New())
	p.AddAddress("localhost:8080")

	testCases := []struct {
		method              string
		contentType         string
		body                string
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			method:              http.MethodPost,
			contentType:         "text/plain",
			body:                "http://ya.ru",
			expectedCode:        http.StatusCreated,
			expectedContentType: "text/plain",
			expectedBody:        "http://localhost:8080/AbCdEf",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

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

func TestmethodGetHandleFunc(t *testing.T) {

}
