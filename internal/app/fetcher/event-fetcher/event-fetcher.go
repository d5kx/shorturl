package eventfetcher

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Fetcher struct {
	//mux     *http.ServeMux
	address string
	Router  chi.Router
	proc    *eventprocessor.Processor
}

func New(address string, processor *eventprocessor.Processor) Fetcher {
	var f Fetcher

	//f.mux = http.NewServeMux()
	f.Router = chi.NewRouter()
	f.proc = processor
	f.address = address
	f.proc.SetAddress(address)

	f.Router.Post(`/`, f.proc.Post)
	f.Router.Get(`/{id}`, f.proc.Get)
	f.Router.NotFound(f.proc.BadRequest)
	f.Router.MethodNotAllowed(f.proc.BadRequest)

	return f
}

func (f *Fetcher) Fetch() error {
	//err := http.ListenAndServe(f.address, f.mux)
	err := http.ListenAndServe(f.address, f.Router)

	if err != nil {
		return e.WrapError("can't start http server", err)
	}

	return nil
}

/*
func (f *Fetcher) AddHandler(pattern string, handler *eventprocessor.Processor) {
	f.mux.HandleFunc(pattern, handler.Process)
	handler.SetAddress(f.address)
}
*/
