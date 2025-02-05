package eventfetcher

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/d5kx/shorturl/internal/util/e"
)

type Fetcher struct {
	//mux     *http.ServeMux
	address string
	Router  chi.Router
}

func New(address string) Fetcher {
	var f Fetcher

	//f.mux = http.NewServeMux()
	f.Router = chi.NewRouter()
	f.address = address

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
