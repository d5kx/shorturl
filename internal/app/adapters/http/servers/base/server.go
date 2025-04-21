package baseserver

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/adapters/http/routers"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/util/e"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	httpServer  *http.Server
	httpsServer *http.Server
	router      routers.Router
	log         loggers.Logger
}

func New(router routers.Router, logger loggers.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    conf.GetServAdr(),
			Handler: router.Mux(),
		},
		httpsServer: &http.Server{
			Addr:    conf.GetTSLServAdr(),
			Handler: router.Mux(),
		},
		router: router,
		log:    logger,
	}
}

func (s *Server) Run() chan error {
	// канал сбора ошибок
	errorsCh := make(chan error)

	//запускаем http сервер
	go func() {
		s.log.Info("HTTP server is starting...",
			zap.String("address", conf.GetServAdr()),
			zap.String("address of response", conf.GetResURLAdr()),
			zap.String("log level", conf.GetLoggerLevel()),
		)
		err := s.httpServer.ListenAndServe()
		// ошибки отправляем в канал
		if err != nil {
			errorsCh <- e.WrapError("can't start HTTP servers", err)
		}
	}()

	// запускаем https сервер
	go func() {
		s.log.Info("HTTPS server is starting...",
			zap.String("address", conf.GetTSLServAdr()),
			zap.String("log level", conf.GetLoggerLevel()),
		)
		err := s.httpsServer.ListenAndServeTLS(conf.GetTSLCertFileName(), conf.GetTSLKeyFileName())
		// ошибки отправляем в канал
		if err != nil {
			errorsCh <- e.WrapError("can't start HTTPS servers", err)
		}
	}()

	return errorsCh
}

func (s *Server) Shutdown(httpDoneCh chan bool, httpsDoneCh chan bool) chan error {
	// канал сбора ошибок
	errorsCh := make(chan error)

	go func() {
		s.log.Info("HTTP server is shutting down...")
		s.httpServer.SetKeepAlivesEnabled(false)
		err := s.httpServer.Shutdown(context.Background())
		if err != nil {
			errorsCh <- e.WrapError("can't stop HTTP servers", err)
		}
		httpDoneCh <- true
	}()

	go func() {
		s.log.Info("HTTPS server is shutting down...")
		s.httpsServer.SetKeepAlivesEnabled(false)
		err := s.httpsServer.Shutdown(context.Background())
		if err != nil {
			errorsCh <- e.WrapError("can't stop HTTPS servers", err)
		}
		httpsDoneCh <- true
	}()

	return errorsCh
}
