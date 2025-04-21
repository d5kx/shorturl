package main

import (
	"context"
	"errors"
	"github.com/d5kx/shorturl/internal/app/adapters/auth/base"
	"github.com/d5kx/shorturl/internal/app/adapters/compress/gzip"
	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/routers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/servers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/simple"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/zap"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/file"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/manager"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/mem"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/sql/postgre"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/db"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
	"github.com/d5kx/shorturl/internal/util/generators/basegen"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// go install github.com/golang/mock/mockgen@latest
// mockgen -destination=internal/app/adapters/storages/gomock/gomockstor.go -package=gomockstor github.com/d5kx/shorturl/internal/app/usecases LinkStorage,DB

// go run main.go -l debug -f tmp/short-url-db.json -d "host=localhost port=5432 user=postgres password=820610 dbname=shorturl sslmode=disable"

func init() {
	conf.ParseFlags()
}

func main() {
	sl := simplelogger.New()

	zl, err := zaplogger.New()
	if err != nil {
		sl.Fatal("can't run zap loggers", err)
	}

	m := memstor.New(zl)
	f := filestor.New(zl)
	p := postgre.New(zl)
	manager := storman.New(m, f, p, zl)
	manager.Open("")
	defer manager.Close()
	manager.Bootstrap(context.Background())

	gen := basegen.New()
	linkUse := uselink.New(manager, gen, zl)
	dbUse := usedb.New(p)
	compressor := gzipc.New(zl)
	auth := baseauth.New(gen, zl)
	auth.GenerateTLSCertificate()

	handler := basehandler.New(linkUse, dbUse, zl)
	router := baserouter.New(handler, compressor, auth, zl)
	server := baseserver.New(router, zl)

	// канал приема системных сигналов
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM)

	errorsCh := make(chan error)
	httpDoneCh := make(chan bool)
	httpsDoneCh := make(chan bool)
	defer func() {
		close(errorsCh)
		close(httpDoneCh)
		close(httpsDoneCh)
		close(quitCh)
	}()

	go func() {
		s := <-quitCh
		zl.Info("received signal", zap.String("code", s.String()))
		stopErrCh := server.Shutdown(httpDoneCh, httpsDoneCh)
		select {
		case stopErr := <-stopErrCh:
			zl.Info("can't stop service", zap.Error(stopErr))
		}
	}()

	// запускаем сервер с обработкой ошибки с канала
	errorsCh = server.Run()

	select {
	case startErr := <-errorsCh:
		if !errors.Is(startErr, http.ErrServerClosed) {
			zl.Info("can't run service", zap.Error(startErr))
		}
	}

	for i := 0; i < 2; {
		select {
		case <-httpDoneCh:
			zl.Info("HTTP server is stopped")
			i++
			//close(httpDoneCh)

		case <-httpsDoneCh:
			zl.Info("HTTPS server is stopped")
			i++
			//close(httpsDoneCh)
		}
	}
}
