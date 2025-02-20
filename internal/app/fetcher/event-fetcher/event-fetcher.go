package eventfetcher

import (
	"net/http"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/log"
	"github.com/d5kx/shorturl/internal/app/processor"
	"github.com/d5kx/shorturl/internal/util/e"

	"github.com/go-chi/chi/v5"
)

type Fetcher struct {
	Router chi.Router
	proc   processor.Processor
	log    logger.Logger
}

func New(processor processor.Processor, logger logger.Logger) *Fetcher {
	var f Fetcher
	f.log = logger
	f.proc = processor

	f.Router = chi.NewRouter()
	f.Router.Post(`/`, f.log.RequestLogging(f.proc.Post))
	f.Router.Post(`/api/shorten`, f.log.RequestLogging(f.proc.PostAPIShorten))
	f.Router.Get(`/{id}`, f.log.RequestLogging(f.proc.Get))
	f.Router.NotFound(f.log.RequestLogging(f.proc.BadRequest))
	f.Router.MethodNotAllowed(f.log.RequestLogging(f.proc.BadRequest))

	return &f
}

func (f *Fetcher) Fetch() error {
	err := http.ListenAndServe(conf.GetServAdr(), f.Router)
	if err != nil {
		return e.WrapError("can't start http server", err)
	}

	return nil
}
