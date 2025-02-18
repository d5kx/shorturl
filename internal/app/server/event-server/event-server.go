package eventserver

import (
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/log"
	"github.com/d5kx/shorturl/internal/util/e"
	"go.uber.org/zap"
)

type Server struct {
	fetcher *eventfetcher.Fetcher
	log     logger.Logger
}

func (s *Server) Run() error {
	s.log.Info("running server",
		zap.String("server address", conf.GetServAdr()),
		zap.String("base address of responce", conf.GetResURLAdr()),
	)
	err := s.fetcher.Fetch()
	if err != nil {
		return e.WrapError("can't start fetcher", err)
	}

	return nil
}

func New(fetcher *eventfetcher.Fetcher, logger logger.Logger) Server {
	return Server{
		fetcher: fetcher,
		log:     logger,
	}
}
