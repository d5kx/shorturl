package event_processor

import (
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/storage"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Processor struct {
	stor storage.Storage
}

func New(storage storage.Storage) Processor {
	return Processor{stor: storage}
}

func (p Processor) Process(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost && req.Header.Get("Content-Type") == "text/plain" && req.ContentLength > 0 {

		b := make([]byte, req.ContentLength)
		_, err := req.Body.Read(b)
		if err != nil {
			e.WrapError("can't process POST request", err)
			return
		}

		var sb strings.Builder
		sb.Write(b)
		var l = storage.Link{URL: sb.String()}

		var sUrl string
		sUrl, err = p.stor.Save(&l)
		if err != nil {
			e.WrapError("can't process POST request", err)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)

		res.Write([]byte("http://" + event_fetcher.ServerAdress + "/" + sUrl + "#" + l.URL))
		return

	}

	if req.Method == http.MethodGet /*&& req.Header.Get("Content-Type") == "text/plain"*/ {
		slice := strings.Split(strings.TrimLeft(req.URL.Path, "/"), "/")
		if len(slice) == 0 {
			return
		}

		l, err := p.stor.Get(slice[0])

		if err != nil {
			e.WrapError("can't process GET request", err)
			return
		}
		if l == nil {
			return
		}

		res.Header().Set("Location", l.URL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		//res.Write([]byte((*l).URL))
		//slice := strings.Split(strings.TrimLeft(req.URL.Path, "/"), "/")
		//res.Write([]byte(strings.TrimLeft(req.URL.Path, "/")))
		//res.Write([]byte(slice[0]))

		return
	}

	res.WriteHeader(http.StatusBadRequest)
	return

}
