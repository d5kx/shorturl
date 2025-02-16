package eventprocessor

import (
	"github.com/d5kx/shorturl/internal/app/logger"
	"io"
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/link"
	"github.com/d5kx/shorturl/internal/app/storage"
)

type Processor struct {
	db  storage.Storage
	log logger.Logger
}

func New(storage storage.Storage, logger logger.Logger) Processor {
	return Processor{
		db:  storage,
		log: logger,
	}
}

func (p *Processor) Get(res http.ResponseWriter, req *http.Request) {
	l, err := p.db.Get(strings.TrimPrefix(req.URL.Path, "/"))

	if err != nil || l == nil {

		p.log.Info("can't process GET request (short link does not exist in the database)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", l.OriginalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (p *Processor) Post(res http.ResponseWriter, req *http.Request) {
	if !strings.Contains(req.Header.Get("Content-Type"), "text/plain") {
		p.log.Info("can't process POST request (wrong Content-Type)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	b, _ := io.ReadAll(req.Body)
	defer req.Body.Close()
	if len(b) == 0 {
		p.log.Info("can't process POST request (no link in body request)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var sb strings.Builder
	sb.Write(b)
	var l = link.Link{OriginalURL: sb.String()}

	sURL, err := p.db.Save(&l)
	if err != nil {
		p.log.Info("can't process POST request (short link is not saved in the database)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(strings.Join([]string{conf.GetResURLAdr(), "/", sURL}, "")))
	if err != nil {
		p.log.Info("can't process POST request (can't write response body)")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (p *Processor) BadRequest(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusBadRequest)
}
