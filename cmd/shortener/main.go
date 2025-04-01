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
	"github.com/d5kx/shorturl/internal/app/adapters/storages/file"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/manager"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/mem"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/sql/postgre"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/db"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
	"github.com/d5kx/shorturl/internal/util/generators/basegen"
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

	handler := basehandler.New(linkUse, dbUse, zl)
	router := baserouter.New(handler, compressor, auth, zl)
	server := baseserver.New(router, zl)
	if err := server.Run(); err != nil {
		sl.Fatal("can't run service", err)
	}

}
