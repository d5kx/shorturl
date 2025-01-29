package event_processor

import (
	"log"
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/storage"
)

type Processor struct {
	stor storage.Storage
}

func New(storage storage.Storage) Processor {
	return Processor{stor: storage}
}

func (p Processor) Process(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost && req.Header.Get("Content-Type") == "text/plain" {

		b := make([]byte, req.ContentLength)
		n, _ := req.Body.Read(b)
		if n == 0 {
			log.Println("can't process POST request (no link in request)/")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		var sb strings.Builder
		sb.Write(b)
		var l = storage.Link{URL: sb.String()}

		sUrl, err := p.stor.Save(&l)
		if err != nil {
			log.Println("can't process POST request (short link is not saved in the database)")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte("http://" + event_fetcher.ServerAdress + "/" + sUrl))
		return
	}

	if req.Method == http.MethodGet && req.Header.Get("Content-Type") == "text/plain" {

		l, err := p.stor.Get(strings.TrimPrefix(req.URL.Path, "/"))

		if err != nil || l == nil {

			log.Println("can't process GET request (short link does not exist)")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Location", l.URL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	log.Println("can't process request (request type is not supported)")
	res.WriteHeader(http.StatusBadRequest)
	return
}
