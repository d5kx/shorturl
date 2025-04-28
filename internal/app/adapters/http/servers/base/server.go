package baseserver

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/d5kx/shorturl/internal/app/adapters/http/routers"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/adapters/storages"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/util/e"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type Server struct {
	httpServer  *http.Server
	httpsServer *http.Server
	router      routers.Router
	log         loggers.Logger
	stor        storages.ManagedStorage
}

func New(
	router routers.Router,
	storage storages.ManagedStorage,
	logger loggers.Logger,
) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    conf.GetServAdr(),
			Handler: router.Mux(),
		},
		httpsServer: &http.Server{
			Addr:    conf.GetTSLServAdr(),
			Handler: router.Mux(),
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
		router: router,
		log:    logger,
		stor:   storage,
	}
}

func (s *Server) Run(ctx context.Context) error {
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

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		<-ctx.Done()
		s.log.Info("received stop signal")

		func() {
			defer wg.Done()
			s.log.Info("storage is shutting down...")
			err := s.stor.Shutdown()
			if err != nil {
				s.log.Info("can't gracefully shutdown storage", zap.Error(err))
			}
			s.log.Info("storage is stopped")
			s.log.Info("HTTP server is shutting down...")
			s.httpServer.SetKeepAlivesEnabled(false)
			err = s.httpServer.Shutdown(ctx)
			if err != nil {
				s.log.Info("can't gracefully shutdown HTTP servers", zap.Error(err))
			}
			s.log.Info("HTTP server is stopped")
		}()

		func() {
			defer wg.Done()
			s.log.Info("HTTPS server is shutting down...")
			s.httpsServer.SetKeepAlivesEnabled(false)
			err := s.httpsServer.Shutdown(ctx)
			if err != nil {
				s.log.Info("can't gracefully shutdown HTTPS servers", zap.Error(err))
			}
			s.log.Info("HTTPS server is stopped")
		}()

	}()

	select {
	case err := <-errorsCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	wg.Wait()
	return nil
}
