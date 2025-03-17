package main

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/adapters/compress/gzip"
	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/routers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/servers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/simple"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/zap"
	filestor "github.com/d5kx/shorturl/internal/app/adapters/storages/file"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/mem"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/sql/postgre"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/db"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
	"github.com/d5kx/shorturl/internal/util/generators/basegen"
)

// curl -v -X POST -H "Content-Type:text/plain" -H -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:application/json"  -H "Accept-Encoding:gzip" --output "-" -d "{\"url\": \"https://practicum.yandex.ru\"}" "http://localhost:9090/api/shorten"
// curl -v -X GET -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" "http://localhost:8080/GlTBlr"
// curl -v -X GET "http://localhost:8080/ping"

// shortenertest-windows-amd64 -test.v -test.run=^TestIteration1$ -binary-path=C:\go\shorturl\cmd\shortener\shortener.exe
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration2$ -source-path=C:\go\shorturl\internal\app\handlers\event-handlers\event-processor_test.go

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

	postg := postgre.New(zl)
	err = postg.Open(conf.GetPostgreSQLConnectionString())
	if err != nil {
		sl.Info("can't connect to PostgreSQL db", err)
	}
	defer postg.Close()

	postg.Bootstrap(context.Background())

	postgUse := usedb.New(postg)

	m := memstor.New(zl)
	err = m.LoadFromFile()
	if err != nil {
		sl.Info("can't load DB from file", err)
	}

	f := filestor.New(zl)
	err = f.Open(conf.GetDBFileName())
	if err != nil {
		sl.Info("can't open DB file", err)
	}
	defer f.Close()

	u := uselink.New(m, f, postg, basegen.New(), zl)
	c := gzipc.New(zl)

	handler := basehandler.New(u, postgUse, zl)
	router := baserouter.New(handler, c, zl)
	server := baseserver.New(router, zl)
	if err := server.Run(); err != nil {
		sl.Fatal("can't run service", err)
	}

}
