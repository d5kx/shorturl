package event_processor

import (
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/storage"
	"github.com/d5kx/shorturl/internal/util/err"
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
		req.Body.Read(b)

		var sb strings.Builder
		sb.Write(b)
		var l = storage.Link{URL: sb.String()}

		sUrl, e := p.stor.Save(&l)
		if e != nil {
			err.WrapError("can't process POST request", e)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)

		res.Write([]byte("http://" + event_fetcher.ServerAdress + "/" + sUrl + "#" + l.URL))
		return

	}

	if req.Method == http.MethodGet {

		l, e := p.stor.Get("")

		if e != nil {
			err.WrapError("can't process GET request", e)
			return
		}
		if l == nil {
			//return
		}

		//res.Header().Set("Location", l.URL)
		//res.WriteHeader(http.StatusTemporaryRedirect)
		//res.Write([]byte((*l).URL))
		res.Write([]byte(req.URL.Scheme))

		return
	}

	res.WriteHeader(http.StatusBadRequest)
	return

}
