package event_server

import (
	"github.com/d5kx/shorturl/internal/app/fetcher"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Server struct {
	fetcher fetcher.Fetcher
}

func (s *Server) Run() error {
	err := s.fetcher.Fetch()
	if err != nil {
		return e.WrapError("can't start fetcher", err)
	}

	return nil
}

func New(fetcher fetcher.Fetcher) Server {
	return Server{fetcher: fetcher}
}
