package baserouter

import (
	"bytes"
	"compress/gzip"
	usedb "github.com/d5kx/shorturl/internal/app/usecases/db"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d5kx/shorturl/internal/util/generators/mockgen"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/d5kx/shorturl/internal/app/adapters/compress/gzip"
	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/mock"

	"github.com/d5kx/shorturl/internal/app/adapters/storages/gomock"
	//"github.com/d5kx/shorturl/internal/app/adapters/storages/mock"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
)

func TestRouter(t *testing.T) {
	conf.ParseFlags()
	ml := mocklogger.New()

	// создадим конроллер моков и экземпляр мок-хранилища
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := gomockstor.NewMockLinkStorage(ctrl)
	ping := gomockstor.NewMockDB(ctrl)

	s.EXPECT().Get(gomock.Any(), "AbCdEf").Return("http://ya.ru", nil)
	s.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", nil)

	s.EXPECT().IsExist(gomock.Any(), gomock.Any()).Return(false, nil)
	s.EXPECT().IsExist(gomock.Any(), gomock.Any()).AnyTimes()

	s.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	s.EXPECT().Save(gomock.Any(), gomock.Any()).AnyTimes()

	ping.EXPECT().Ping(gomock.Any()).Return(true)
	ping.EXPECT().Ping(gomock.Any()).AnyTimes()

	postUse := usedb.New(ping)

	u := uselink.New(s /*mockstor.New()*/, mockgen.New(), ml)
	c := gzipc.New(ml)
	p := basehandler.New(u, postUse, ml)
	f := New(p, c, ml)

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
		/*{
			name:                "POST: db error emulation",
			method:              http.MethodPost,
			path:                "/",
			contentType:         "text/plain",
			body:                "db_error",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "",
		},*/
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
			case "POST: api/json valid compressed request":
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
			if tc.name == "POST: api/json valid compressed request" {
				req.Header.Set("Content-Encoding", "gzip")
				req.Header.Set("Accept-Encoding", "gzip")
			}

			ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			var respBody []byte

			switch tc.name {
			case "POST: api/json valid compressed request":
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
		})
	}
}
