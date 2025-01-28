package event_fetcher

import (
	"fmt"
	"net/http"

	"github.com/d5kx/shorturl/internal/app/processor"
	"github.com/d5kx/shorturl/internal/util/err"
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

	e := http.ListenAndServe(ServerAdress, f.mux)

	if e != nil {
		return err.WrapError("can't start http server", e)
	}
	fmt.Println("Fetch() - OK")
	return nil
}
