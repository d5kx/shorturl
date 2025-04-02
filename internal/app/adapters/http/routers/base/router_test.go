package baserouter

import (
	"bytes"
	"compress/gzip"
	baseauth "github.com/d5kx/shorturl/internal/app/adapters/auth/base"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/zap"
	usedb "github.com/d5kx/shorturl/internal/app/usecases/db"
	"github.com/d5kx/shorturl/internal/util/generators/basegen"
	"github.com/d5kx/shorturl/internal/util/generators/mockgen"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/d5kx/shorturl/internal/app/adapters/compress/gzip"
	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/gomock"
	//"github.com/d5kx/shorturl/internal/app/adapters/storages/mock"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
)

func TestRouter(t *testing.T) {
	conf.ParseFlags()

	// создадим конроллер моков и экземпляр мок-хранилища
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := gomockstor.NewMockLinkStorage(ctrl)
	ping := gomockstor.NewMockDB(ctrl)

	s.EXPECT().Get(gomock.Any(), "AbCdEf").Return("00000-0000", "http://ya.ru", nil)
	s.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", "", nil)

	s.EXPECT().IsExist(gomock.Any(), gomock.Any()).Return(false, nil)
	s.EXPECT().IsExist(gomock.Any(), gomock.Any()).AnyTimes()

	s.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	s.EXPECT().Save(gomock.Any(), gomock.Any()).AnyTimes()

	s.EXPECT().SaveTx(gomock.Any(), gomock.Any()).Return(nil)
	s.EXPECT().SaveTx(gomock.Any(), gomock.Any()).AnyTimes()

	slice := make([][]string, 0)
	slice = append(slice, []string{"asdfgh", "http://aa.ru"})
	slice = append(slice, []string{"zxcvbn", "http://bb.ru"})

	s.EXPECT().GetUserUrls(gomock.Any(), "74a2c362-a6a4-4bd0-9529-8a28f81db1fa").Return(slice, nil)

	s.EXPECT().GetUserUrls(gomock.Any(), gomock.Any()).Return(make([][]string, 0), nil)
	s.EXPECT().GetUserUrls(gomock.Any(), gomock.Any()).AnyTimes()

	ping.EXPECT().Ping(gomock.Any()).Return(true)
	ping.EXPECT().Ping(gomock.Any()).AnyTimes()

	//logger := mocklogger.New()
	logger, err := zaplogger.New()
	assert.NoError(t, err, "Ошибка создания логгера: %s", err)

	gen := basegen.New()

	dbUse := usedb.New(ping)
	linkUse := uselink.New(s, mockgen.New(), logger)

	compressor := gzipc.New(logger)
	handler := basehandler.New(linkUse, dbUse, logger)

	auth := baseauth.New(gen, logger)
	f := New(handler, compressor, auth, logger)

	ts := httptest.NewServer(f.rout)
	defer ts.Close()

	var testTable = []struct {
		name                string
		method              string
		path                string
		contentType         string
		body                string
		cookieName          string
		cookieValue         string
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
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Bad Request\n",
		},
		{
			name:                "POST: no link in the request body",
			method:              http.MethodPost,
			path:                "/",
			contentType:         "text/plain",
			body:                "",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Bad Request\n",
		},
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
		{
			name:                "POST: api/json valid compressed request",
			path:                "/api/shorten",
			method:              http.MethodPost,
			contentType:         "application/json",
			body:                `{"url":"https://practicum.yandex.ru"}`,
			expectedCode:        http.StatusCreated,
			expectedContentType: "application/json",
			expectedBody:        `{"result":"` + conf.GetResURLAdr() + `/AbCdEf"` + `}`,
		},
		{
			name:                "POST: api/json/batch valid compressed request",
			path:                "/api/shorten/batch",
			method:              http.MethodPost,
			contentType:         "application/json",
			body:                `[{"correlation_id":"id=1","original_url":"https://ya1.ru"},{"correlation_id":"id=2","original_url":"https://ya2.ru"}]`,
			expectedCode:        http.StatusCreated,
			expectedContentType: "application/json",
			expectedBody:        `[{"correlation_id":"id=1","short_url":"AbCdEf"},{"correlation_id":"id=2","short_url":"AbCdEf"}]`,
		},
		{
			name:                "GET: valid request",
			method:              http.MethodGet,
			path:                "/AbCdEf",
			contentType:         "text/plain",
			body:                "",
			expectedCode:        http.StatusTemporaryRedirect,
			expectedContentType: "",
			expectedLocation:    "http://ya.ru",
			expectedBody:        "",
		},
		{
			name:                "GET: /api/user/urls valid request",
			method:              http.MethodGet,
			path:                "/api/user/urls",
			contentType:         "application/json",
			body:                "",
			cookieName:          "user_id",
			cookieValue:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQ0NDY3MDAsIlVzZXJJRCI6Ijc0YTJjMzYyLWE2YTQtNGJkMC05NTI5LThhMjhmODFkYjFmYSJ9.bRsY2SC6jDVgHEIfhvMuYtFuwXxpVOoylzMjAN8g9ok",
			expectedCode:        http.StatusOK,
			expectedContentType: "",
			expectedLocation:    "",
			expectedBody:        `[{"short_url":"http://localhost:8080/asdfgh","original_url":"http://aa.ru"},{"short_url":"http://localhost:8080/zxcvbn","original_url":"http://bb.ru"}]`,
		},
		//set user_id	{"user_id": "74a2c362-a6a4-4bd0-9529-8a28f81db1fa"}
		//set cookie	{"cookie": "user_id=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQ0NDY3MDAsIlVzZXJJRCI6Ijc0YTJjMzYyLWE2YTQtNGJkMC05NTI5LThhMjhmODFkYjFmYSJ9.bRsY2SC6jDVgHEIfhvMuYtFuwXxpVOoylzMjAN8g9ok; Path=/; HttpOnly; Secure; SameSite=Strict"}
		{
			name:                "GET: /api/user/urls no links",
			method:              http.MethodGet,
			path:                "/api/user/urls",
			contentType:         "application/json",
			body:                "",
			expectedCode:        http.StatusNoContent,
			expectedContentType: "",
			expectedLocation:    "",
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
		{
			name:             "GET: ping",
			path:             "/ping",
			method:           http.MethodGet,
			contentType:      "",
			body:             "",
			expectedCode:     http.StatusOK,
			expectedLocation: "",
			expectedBody:     "",
		},
	}
	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			var body *bytes.Buffer

			switch tc.name {
			case "POST: api/json valid compressed request", "POST: api/json/batch valid compressed request":
				body = bytes.NewBuffer(nil)
				zb := gzip.NewWriter(body)
				_, err := zb.Write([]byte(tc.body))
				require.NoError(t, err)
				err = zb.Close()
				require.NoError(t, err)
			default:
				body = bytes.NewBuffer([]byte(tc.body))
			}

			req, err := http.NewRequest(tc.method, ts.URL+tc.path, body)
			require.NoError(t, err)

			req.Header.Set("Content-Type", tc.contentType)

			switch tc.name {
			case "POST: api/json valid compressed request", "POST: api/json/batch valid compressed request":
				req.Header.Set("Content-Encoding", "gzip")
				req.Header.Set("Accept-Encoding", "gzip")
			case "GET: /api/user/urls valid request":
				cookie := &http.Cookie{Name: tc.cookieName, Value: tc.cookieValue, Path: "/"}
				jar, _ := cookiejar.New(nil)
				jar.SetCookies(req.URL, []*http.Cookie{cookie})
				ts.Client().Jar = jar
			}

			ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			var respBody []byte

			switch tc.name {
			case "POST: api/json valid compressed request", "POST: api/json/batch valid compressed request":
				zr, err := gzip.NewReader(resp.Body)
				require.NoError(t, err)

				respBody, err = io.ReadAll(zr)
				require.NoError(t, err)
			default:
				respBody, err = io.ReadAll(resp.Body)
				require.NoError(t, err)
			}

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

			switch tc.name {
			case "GET: /api/user/urls valid request":
				assert.Equal(t, tc.expectedBody, sb.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}
