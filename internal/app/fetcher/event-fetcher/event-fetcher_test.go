package eventfetcher

import (
	"bytes"
	"github.com/d5kx/shorturl/internal/app/log/simple"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/stor/mock"
)

func TestRouter(t *testing.T) {
	sl := simplelogger.GetInstance()
	p := eventprocessor.New(mockstor.New(), sl)
	f := New(&p, sl)
	conf.ParseFlags()
	ts := httptest.NewServer(f.Router)
	defer ts.Close()

	var testTable = []struct {
		name                string
		method              string
		path                string
		contentType         string
		body                string
		expectedCode        int
		expectedContentType string
		expectedBody        string
		expectedLocation    string
	}{
		{
			name:                "CONNECT request",
			method:              http.MethodConnect,
			path:                "/",
			contentType:         "text/plain",
			body:                "http://ya.ru",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
		{
			name:                "POST: valid request",
			method:              http.MethodPost,
			path:                "/",
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
			method:              http.MethodPost,
			path:                "/",
			contentType:         "text/plain",
			body:                "",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
		{
			name:                "POST: db error emulation",
			method:              http.MethodPost,
			path:                "/",
			contentType:         "text/plain",
			body:                "db_error",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},
		{
			name:                "GET: valid request",
			method:              http.MethodGet,
			path:                "/AbCdEf",
			contentType:         "text/plain",
			body:                "",
			expectedCode:        http.StatusTemporaryRedirect,
			expectedContentType: "text/plain",
			expectedLocation:    "http://ya.ru",
			expectedBody:        "",
		},
		{
			name:             "GET: non-existent short link",
			method:           http.MethodGet,
			path:             "/ZbCdEf",
			contentType:      "text/plain",
			body:             "",
			expectedCode:     http.StatusBadRequest,
			expectedLocation: "",
			expectedBody:     "",
		},
		{
			name:             "GET: short link missing in request",
			path:             "/",
			method:           http.MethodGet,
			contentType:      "text/plain",
			body:             "",
			expectedCode:     http.StatusBadRequest,
			expectedLocation: "",
			expectedBody:     "",
		},
	}
	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			body := bytes.NewBuffer([]byte(tc.body))
			req, err := http.NewRequest(tc.method, ts.URL+tc.path, body)
			require.NoError(t, err)

			req.Header.Set("Content-Type", tc.contentType)

			ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var sb strings.Builder
			sb.Write(respBody)

			switch tc.method {
			case http.MethodPost:
				assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
				assert.Equal(t, tc.expectedContentType, resp.Header.Get("Content-Type"), "ContentType не совпадает с ожидаемым")
				assert.Equal(t, tc.expectedBody, sb.String(), "Тело ответа не совпадает с ожидаемым")
			case http.MethodGet:
				assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
				assert.Equal(t, tc.expectedLocation, resp.Header.Get("Location"), "Адрес переадресации не совпадает с ожидаемым")
			default:
				assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			}
		})
	}
}
