package eventfetcher

import (
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

	f.Router.Post(`/`, f.proc.Post)
	f.Router.Get(`/{id}`, f.proc.Get)
	f.Router.NotFound(f.proc.BadRequest)
	f.Router.MethodNotAllowed(f.proc.BadRequest)

	return f
}

func (f *Fetcher) Fetch() error {
	err := http.ListenAndServe(conf.GetServAdr(), f.Router)
	if err != nil {
		return e.WrapError("can't start http server", err)
	}

	return nil
}
