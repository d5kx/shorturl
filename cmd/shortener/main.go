package main

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/adapters/auth/base"
	"github.com/d5kx/shorturl/internal/app/adapters/compress/gzip"
	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/routers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/servers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/simple"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/zap"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/manager"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/mem"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/sql/postgre"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/db"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
	useuser "github.com/d5kx/shorturl/internal/app/usecases/user"
	"github.com/d5kx/shorturl/internal/util/generators/basegen"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

// go install github.com/golang/mock/mockgen@latest
// mockgen -destination=internal/app/adapters/storages/gomock/gomockstor.go -package=gomockstor github.com/d5kx/shorturl/internal/app/usecases LinkStorage,DB

// go run main.go -l debug -f tmp/short-url-db.json -d "host=localhost port=5432 user=postgres password=820610 dbname=shorturl sslmode=disable"
// go run main.go -l debug -f tmp/short-url-db.json -fusers tmp/users-db.json

func init() {
	conf.ParseFlags()
}

func main() {
	sl := simplelogger.New()

	logger, err := zaplogger.New()
	if err != nil {
		sl.Fatal("can't run zap loggers", err)
	}

	m := memstor.New(logger)
	p := postgre.New(logger)
	storage := storman.New(m, p, logger)
	storage.Open("")
	defer storage.Close()
	storage.Bootstrap(context.Background())

	generator := basegen.New()
	linkUse := uselink.New(storage, generator, logger)
	dbUse := usedb.New(p)
	useUser := useuser.New(storage, generator, logger)
	compressor := gzipc.New(logger)
	auth := baseauth.New(generator, logger)
	auth.GenerateTLSCertificate()

	handler := basehandler.New(linkUse, useUser, dbUse, logger)
	router := baserouter.New(handler, compressor, auth, logger)
	server := baseserver.New(router, storage, logger)

	// канал приема системных сигналов
	//quitCh := make(chan os.Signal, 1)
	//signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := server.Run(ctx); err != nil {
		logger.Info("can't run service", zap.Error(err))
		os.Exit(1)
	}
}
