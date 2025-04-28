package baserouter

import (
	"github.com/d5kx/shorturl/internal/app/adapters/auth"
	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers"

	"github.com/d5kx/shorturl/internal/app/adapters/compress"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/go-chi/chi/v5"
)

type BaseRouter struct {
	rout    chi.Router
	handler handlers.Handler
	comp    compress.Compressor
	log     loggers.Logger
	auth    auth.Authorizer
}

func New(handler handlers.Handler, compressor compress.Compressor, authorizer auth.Authorizer, logger loggers.Logger) *BaseRouter {
	var r BaseRouter
	r.log = logger
	r.handler = handler
	r.comp = compressor
	r.auth = authorizer

	r.rout = chi.NewRouter()
	// сохраняет оригинальную ссылку для данного пользователя, принимает в теле запроса, выдает куку авторизации если ее нет
	// curl -v -X POST -H "Content-Type:text/plain" -d "http://ya.ru" "http://localhost:8080"
	// curl -v -X POST -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" -d "http://ya.ru" "http://localhost:8080"
	// curl -v -X POST -H "Content-Type:text/plain" -d "http://ya.ru" --cookie "user_id=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MzA3Njk2MjUsIlVzZXJJRCI6IjdjYjMyZGM1LWQ3M2YtNDBmZi1iNmRmLTI0NjdlMDg3MzYwMyJ9.du4NRzw32M5X_OCu3UOoF4MV59ZFdVsFZlsUyZMZ1hQ" "http://localhost:8080"
	r.rout.Post(`/`, r.log.RequestLogging(r.comp.Do(r.auth.Do(r.handler.Post))))

	// сохраняет оригинальную ссылку для данного пользователя, принимает в json формате, выдает куку авторизации если ее нет
	// curl -v -X POST -H "Content-Type:application/json"  -H "Accept-Encoding:gzip" --output "-" -d "{\"url\": \"https://practicum.yandex.ru\"}" "http://localhost:8080/api/shorten"
	r.rout.Post(`/api/shorten`, r.log.RequestLogging(r.comp.Do(r.auth.Do(r.handler.PostAPIShorten))))

	// сохраняет список оригинальных ссылок для данного пользователя, принимает в json формате, выдает куку авторизации если ее нет
	// curl -v -X POST -H "Content-Type:application/json" -d "[{\"correlation_id\":\"id=1\",\"original_url\":\"https://ya1.ru\"},{\"correlation_id\":\"id=2\",\"original_url\":\"https://ya2.ru\"}]", "http://localhost:8080/api/shorten/batch"
	r.rout.Post(`/api/shorten/batch`, r.log.RequestLogging(r.comp.Do(r.auth.Do(r.handler.PostAPIShortenBatch))))

	// регистрирует пользователя по логину/паролю в json формате, выдает подписанную куку
	// curl -v -X POST -H "Content-Type:application/json" -d "{\"login\": \"striped\",\"password\": \"820610\"}" "http://localhost:8080/api/user/register"
	r.rout.Post(`/api/user/register`, r.log.RequestLogging(r.comp.Do(r.handler.PostAPIUserRegister(r.auth.SendUserAuthCookie(nil)))))

	// авторизирует пользователя по логину/паролю в json формате, выдает подписанную куку
	// curl -v -X POST -H "Content-Type:application/json" -d "{\"login\": \"striped\",\"password\": \"820610\"}" "http://localhost:8080/api/user/login"
	r.rout.Post(`/api/user/login`, r.log.RequestLogging(r.comp.Do(r.handler.PostAPIUserLogin(r.auth.SendUserAuthCookie(nil)))))

	// обработчик для тестирования https сервера
	// curl -Lv  --cacert "C:\go\shorturl\cmd\shortener\sec\cert.pem" https://localhost:4040
	r.rout.Get(`/`, r.log.RequestLogging(r.handler.GetHTTPS))

	// выдает информации о результатах пинга БД
	// curl -v -X GET "http://localhost:8080/ping"
	r.rout.Get(`/ping`, r.log.RequestLogging(r.handler.PingDB))

	// выдает оригинальную ссылку по короткой, отправляет статус 307, редирект по оригинальной ссылке
	// curl -v -X GET -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" "http://localhost:8080/EeZjtNwXX"
	r.rout.Get(`/{id}`, r.log.RequestLogging(r.comp.Do(r.handler.Get)))

	// выдает в json список всех ссылок пользователя
	// curl -v -X GET -H "Content-Type:text/plain" --cookie "user_id=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MzA3Njk2MjUsIlVzZXJJRCI6IjdjYjMyZGM1LWQ3M2YtNDBmZi1iNmRmLTI0NjdlMDg3MzYwMyJ9.du4NRzw32M5X_OCu3UOoF4MV59ZFdVsFZlsUyZMZ1hQ" "http://localhost:8080/api/user/urls"
	r.rout.Get(`/api/user/urls`, r.log.RequestLogging(r.comp.Do(r.auth.Do(r.handler.GetUserUrls))))

	// помечает ссылки в БД как удаленные для данного пользователя
	// curl -v -X DELETE -H "Content-Type:application/json" -d "[\"6qxTVvsy\", \"RTfd56hn\"]" --cookie "user_id=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MzA3Njk2MjUsIlVzZXJJRCI6IjdjYjMyZGM1LWQ3M2YtNDBmZi1iNmRmLTI0NjdlMDg3MzYwMyJ9.du4NRzw32M5X_OCu3UOoF4MV59ZFdVsFZlsUyZMZ1hQ" "http://localhost:8080/api/user/urls"
	r.rout.Delete(`/api/user/urls`, r.log.RequestLogging(r.comp.Do(r.auth.Do(r.handler.DeleteUserUrls))))

	r.rout.NotFound(r.log.RequestLogging(r.comp.Do(r.handler.BadRequest)))
	r.rout.MethodNotAllowed(r.log.RequestLogging(r.comp.Do(r.handler.BadRequest)))

	return &r
}

func (r *BaseRouter) Mux() chi.Router {
	return r.rout
}

func (r *BaseRouter) Run() error {

	return nil
}
