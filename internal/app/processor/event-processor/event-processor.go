package eventprocessor

import (
	"log"
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/link"
	"github.com/d5kx/shorturl/internal/app/storage"
)

type Processor struct {
	db      storage.Storage
	address string
}

func New(storage storage.Storage) Processor {
	return Processor{db: storage}
}
func (p *Processor) AddAddress(address string) {
	p.address = address
}
func (p *Processor) Process(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		p.methodPostHandleFunc(res, req)
		return
	}

	if req.Method == http.MethodGet {
		p.methodGetHandleFunc(res, req)
		return
	}

	log.Println("can't process request (request type is not supported)")
	res.WriteHeader(http.StatusBadRequest)
}

func (p *Processor) methodGetHandleFunc(res http.ResponseWriter, req *http.Request) {
	l, err := p.db.Get(strings.TrimPrefix(req.URL.Path, "/"))

	if err != nil || l == nil {

		log.Println("can't process GET request (short link does not exist)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", l.URL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (p *Processor) methodPostHandleFunc(res http.ResponseWriter, req *http.Request) {
	if !strings.Contains(req.Header.Get("Content-Type"), "text/plain") {
		log.Println("can't process POST request (wrong Content-Type)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	b := make([]byte, req.ContentLength)
	n, _ := req.Body.Read(b)
	defer req.Body.Close()
	if n == 0 {
		log.Println("can't process POST request (no link in request)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var sb strings.Builder
	sb.Write(b)
	var l = link.Link{URL: sb.String()}

	sURL, err := p.db.Save(&l)
	if err != nil {
		log.Println("can't process POST request (short link is not saved in the database)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://" + p.address + "/" + sURL))
}
