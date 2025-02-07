package eventfetcher

import (
	"net/http"

	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Fetcher struct {
	mux     *http.ServeMux
	address string
}

func New(address string) Fetcher {
	var f Fetcher

	f.mux = http.NewServeMux()
	f.address = address

	return f
}

func (f *Fetcher) Fetch() error {

	err := http.ListenAndServe(f.address, f.mux)

	if err != nil {
		return e.WrapError("can't start http server", err)
	}

	return nil
}

func (f *Fetcher) AddHandler(pattern string, handler *eventprocessor.Processor) {
	f.mux.HandleFunc(pattern, handler.Process)
	handler.AddAddress(f.address)
}
