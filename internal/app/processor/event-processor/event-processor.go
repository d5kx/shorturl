package event_processor

import (
	event_fetcher "github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"log"
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/storage"
)

type Processor struct {
	db storage.Storage
}

func New(storage storage.Storage) Processor {
	return Processor{db: storage}
}

func (p Processor) Process(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost && strings.Contains(req.Header.Get("Content-Type"), "text/plain") {
		p.methodPostHandleFunc(res, req)
		return
	}

	if req.Method == http.MethodGet {
		p.methodGetHandleFunc(res, req)
		return
	}

	log.Println("can't process request (request type is not supported)")
	res.WriteHeader(http.StatusBadRequest)
	return
}

func (p Processor) methodGetHandleFunc(res http.ResponseWriter, req *http.Request) {
	l, err := p.db.Get(strings.TrimPrefix(req.URL.Path, "/"))

	if err != nil || l == nil {

		log.Println("can't process GET request (short link does not exist)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", l.URL)
	res.WriteHeader(http.StatusTemporaryRedirect)
	return
}

func (p Processor) methodPostHandleFunc(res http.ResponseWriter, req *http.Request) {
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

	sUrl, err := p.db.Save(&l)
	if err != nil {
		log.Println("can't process POST request (short link is not saved in the database)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://" + event_fetcher.ServerAddress + "/" + sUrl))
	return
}
