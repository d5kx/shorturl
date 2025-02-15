package eventfetcher

import (
	"github.com/d5kx/shorturl/internal/app/logger/zaplogger"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Fetcher struct {
	Router chi.Router
	proc   *eventprocessor.Processor
	log    *zaplogger.ZapLogger
}

func New(processor *eventprocessor.Processor) Fetcher {
	var f Fetcher
	f.log = zaplogger.GetInstance()

	f.proc = processor

	f.Router = chi.NewRouter()
	f.Router.Post(`/`, f.log.RequestLogging(f.proc.Post))
	f.Router.Get(`/{id}`, f.log.RequestLogging(f.proc.Get))
	f.Router.NotFound(f.log.RequestLogging(f.proc.BadRequest))
	f.Router.MethodNotAllowed(f.log.RequestLogging(f.proc.BadRequest))

	return f
}

func (f *Fetcher) Fetch() error {
	err := http.ListenAndServe(conf.GetServAdr(), f.Router)
	if err != nil {
		return e.WrapError("can't start http server", err)
	}

	return nil
}
