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
	Router chi.Router
	proc   handlers.Handler
	comp   compress.Compressor
	log    loggers.Logger
}

func New(processor handlers.Handler, compressor compress.Compressor, logger loggers.Logger) *BaseRouter {
	var r BaseRouter
	r.log = logger
	r.proc = processor
	r.comp = compressor

	r.Router = chi.NewRouter()
	r.Router.Post(`/`, r.log.RequestLogging(r.comp.RequestCompress(r.proc.Post)))
	r.Router.Post(`/api/shorten`, r.log.RequestLogging(r.comp.RequestCompress(r.proc.PostAPIShorten)))
	r.Router.Get(`/ping`, r.log.RequestLogging(r.proc.PingDB))
	r.Router.Get(`/{id}`, r.log.RequestLogging(r.comp.RequestCompress(r.proc.Get)))
	r.Router.NotFound(r.log.RequestLogging(r.comp.RequestCompress(r.proc.BadRequest)))
	r.Router.MethodNotAllowed(r.log.RequestLogging(r.comp.RequestCompress(r.proc.BadRequest)))

	return &r
}

func (r *BaseRouter) Run() error {
	err := http.ListenAndServe(conf.GetServAdr(), r.Router)
	if err != nil {
		return e.WrapError("can't start http servers", err)
	}

	return nil
}
