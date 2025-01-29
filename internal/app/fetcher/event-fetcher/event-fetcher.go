package event_fetcher

import (
	"fmt"
	"net/http"

	"github.com/d5kx/shorturl/internal/app/processor"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Fetcher struct {
	mux  *http.ServeMux
	proc processor.Processor
}

const (
	ServerAdress = "localhost:8080"
)

func New(processor processor.Processor) Fetcher {
	var f Fetcher

	f.proc = processor
	f.mux = http.NewServeMux()
	f.mux.HandleFunc(`/`, f.proc.Process)

	return f
}

func (f Fetcher) Fetch() error {

	err := http.ListenAndServe(ServerAdress, f.mux)

	if err != nil {
		return e.WrapError("can't start http server", err)
	}

	return nil
}
