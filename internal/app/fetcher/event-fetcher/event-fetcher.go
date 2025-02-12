package eventfetcher

import (
	"github.com/d5kx/shorturl/internal/app/logger"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Fetcher struct {
	Router chi.Router
	proc   *eventprocessor.Processor
}

func New(processor *eventprocessor.Processor) Fetcher {
	var f Fetcher

	f.Router = chi.NewRouter()
	f.proc = processor

	f.Router.Post(`/`, logger.RequestLogger(f.proc.Post))
	f.Router.Get(`/{id}`, logger.RequestLogger(f.proc.Get))
	f.Router.NotFound(logger.RequestLogger(f.proc.BadRequest))
	f.Router.MethodNotAllowed(logger.RequestLogger(f.proc.BadRequest))

	return f
}

func (f *Fetcher) Fetch() error {
	err := http.ListenAndServe(conf.GetServAdr(), f.Router)
	if err != nil {
		return e.WrapError("can't start http server", err)
	}

	return nil
}
