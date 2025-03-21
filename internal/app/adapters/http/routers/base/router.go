package baserouter

import (
	"net/http"

	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers"

	"github.com/d5kx/shorturl/internal/app/adapters/compress"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/util/e"

	"github.com/go-chi/chi/v5"
)

type BaseRouter struct {
	Router  chi.Router
	handler handlers.Handler
	comp    compress.Compressor
	log     loggers.Logger
}

func New(handler handlers.Handler, compressor compress.Compressor, logger loggers.Logger) *BaseRouter {
	var r BaseRouter
	r.log = logger
	r.handler = handler
	r.comp = compressor

	r.Router = chi.NewRouter()
	r.Router.Post(`/`, r.log.RequestLogging(r.comp.RequestCompress(r.handler.Post)))
	r.Router.Post(`/api/shorten`, r.log.RequestLogging(r.comp.RequestCompress(r.handler.PostAPIShorten)))
	r.Router.Post(`/api/shorten/batch`, r.log.RequestLogging(r.comp.RequestCompress(r.handler.PostAPIShortenBatch)))
	r.Router.Get(`/ping`, r.log.RequestLogging(r.handler.PingDB))
	r.Router.Get(`/{id}`, r.log.RequestLogging(r.comp.RequestCompress(r.handler.Get)))
	r.Router.NotFound(r.log.RequestLogging(r.comp.RequestCompress(r.handler.BadRequest)))
	r.Router.MethodNotAllowed(r.log.RequestLogging(r.comp.RequestCompress(r.handler.BadRequest)))

	return &r
}

func (r *BaseRouter) Run() error {
	err := http.ListenAndServe(conf.GetServAdr(), r.Router)
	if err != nil {
		return e.WrapError("can't start http servers", err)
	}

	return nil
}
