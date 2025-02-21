package baserouter

import (
	"net/http"

	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/handlers"
	"github.com/d5kx/shorturl/internal/util/e"

	"github.com/go-chi/chi/v5"
)

type BaseRouter struct {
	Router chi.Router
	proc   handlers.Handler
	log    loggers.Logger
}

func New(processor handlers.Handler, logger loggers.Logger) *BaseRouter {
	var r BaseRouter
	r.log = logger
	r.proc = processor

	r.Router = chi.NewRouter()
	r.Router.Post(`/`, r.log.RequestLogging(r.proc.Post))
	r.Router.Post(`/api/shorten`, r.log.RequestLogging(r.proc.PostAPIShorten))
	r.Router.Get(`/{id}`, r.log.RequestLogging(r.proc.Get))
	r.Router.NotFound(r.log.RequestLogging(r.proc.BadRequest))
	r.Router.MethodNotAllowed(r.log.RequestLogging(r.proc.BadRequest))

	return &r
}

func (r *BaseRouter) Run() error {
	err := http.ListenAndServe(conf.GetServAdr(), r.Router)
	if err != nil {
		return e.WrapError("can't start http servers", err)
	}

	return nil
}
