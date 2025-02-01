package eventserver

import (
	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Server struct {
	fetcher *eventfetcher.Fetcher
}

func (s *Server) Run() error {
	err := s.fetcher.Fetch()
	if err != nil {
		return e.WrapError("can't start fetcher", err)
	}

	return nil
}

func New(fetcher *eventfetcher.Fetcher) Server {
	return Server{fetcher: fetcher}
}
