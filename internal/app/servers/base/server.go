package baseserver

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

	s.log.Info("running servers",
		zap.String("servers address", conf.GetServAdr()),
		zap.String("base address of response", conf.GetResURLAdr()),
		zap.String("log level", conf.GetLoggerLevel()),
	)
	err := s.router.Run()
	if err != nil {
		return e.WrapError("can't start routers", err)
	}

	return nil
}

func New(router routers.Router, logger logger.Logger) *Server {
	return &Server{
		router: router,
		log:    logger,
	}
}
