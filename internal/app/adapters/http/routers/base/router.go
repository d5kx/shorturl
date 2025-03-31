package baserouter

import (
	"github.com/d5kx/shorturl/internal/app/adapters/auth"
	"net/http"

	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers"

	"github.com/d5kx/shorturl/internal/app/adapters/compress"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/util/e"

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
	r.rout.Post(`/`, r.log.RequestLogging(r.comp.Do(r.auth.Do(r.handler.Post))))
	r.rout.Post(`/api/shorten`, r.log.RequestLogging(r.comp.Do(r.handler.PostAPIShorten)))
	r.rout.Post(`/api/shorten/batch`, r.log.RequestLogging(r.comp.Do(r.handler.PostAPIShortenBatch)))
	r.rout.Get(`/ping`, r.log.RequestLogging(r.handler.PingDB))
	r.rout.Get(`/{id}`, r.log.RequestLogging(r.comp.Do(r.handler.Get)))
	r.rout.Get(`/api/user/urls`, r.log.RequestLogging(r.comp.Do(r.auth.Do(r.handler.GetUserUrls))))
	r.rout.NotFound(r.log.RequestLogging(r.comp.Do(r.handler.BadRequest)))
	r.rout.MethodNotAllowed(r.log.RequestLogging(r.comp.Do(r.handler.BadRequest)))

	return &r
}

func (r *BaseRouter) Run() error {
	err := http.ListenAndServe(conf.GetServAdr(), r.rout)
	if err != nil {
		return e.WrapError("can't start http servers", err)
	}

	return nil
}
