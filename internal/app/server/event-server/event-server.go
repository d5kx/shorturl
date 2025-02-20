package eventserver

import (
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/log"
	"github.com/d5kx/shorturl/internal/app/routers"
	"github.com/d5kx/shorturl/internal/util/e"
	"go.uber.org/zap"
)

type Server struct {
	router routers.Router
	log    logger.Logger
}

func (s *Server) Run() error {

	s.log.Info("running server",
		zap.String("server address", conf.GetServAdr()),
		zap.String("base address of response", conf.GetResURLAdr()),
		zap.String("log level", conf.GetLoggerLevel()),
	)
	err := s.router.Run()
	if err != nil {
		return e.WrapError("can't start routers", err)
	}

	return nil
}

func New(fetcher routers.Router, logger logger.Logger) *Server {
	return &Server{
		router: fetcher,
		log:    logger,
	}
}
