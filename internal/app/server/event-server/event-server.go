package eventserver

import (
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/logger/zaplogger"
	"github.com/d5kx/shorturl/internal/util/e"
	"go.uber.org/zap"
)

type Server struct {
	fetcher *eventfetcher.Fetcher
	log     *zaplogger.ZapLogger
}

func (s *Server) Run() error {
	if err := s.log.Init(conf.GetLoggerLevel()); err != nil {
		return e.WrapError("can't start logger", err)
	}

	s.log.Zap().Info("running server",
		zap.String("server address", conf.GetServAdr()),
		zap.String("base address of responce", conf.GetResURLAdr()),
	)
	err := s.fetcher.Fetch()
	if err != nil {
		return e.WrapError("can't start fetcher", err)
	}

	return nil
}

func New(fetcher *eventfetcher.Fetcher) Server {
	return Server{
		fetcher: fetcher,
		log:     zaplogger.GetInstance(),
	}
}
