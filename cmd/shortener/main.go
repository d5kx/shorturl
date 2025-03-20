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
	"github.com/d5kx/shorturl/internal/app/adapters/storages/manager"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/mem"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/sql/postgre"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/db"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
	"github.com/d5kx/shorturl/internal/util/generators/basegen"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:application/json"  -H "Accept-Encoding:gzip" --output "-" -d "{\"url\": \"https://practicum.yandex.ru\"}" "http://localhost:9090/api/shorten"
// curl -v -X GET -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" "http://localhost:8080/GlTBlr"
// curl -v -X GET "http://localhost:8080/ping"
// curl -v -X POST -H "Content-Type:application/json" -d "[{\"correlation_id\":\"id=1\",\"original_url\":\"https://ya1.ru\"},{\"correlation_id\":\"id=2\",\"original_url\":\"https://ya2.ru\"}]", "http://localhost:8080/api/shorten/batch"

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

	m := memstor.New(zl)
	f := filestor.New(zl)
	p := postgre.New(zl)
	manager := storman.New(m, f, p, zl)
	manager.Open("")
	defer manager.Close()
	manager.Bootstrap(context.Background())

	u := uselink.New(manager, basegen.New(), zl)
	postgUse := usedb.New(p)
	compressor := gzipc.New(zl)

	handler := basehandler.New(u, postgUse, zl)
	router := baserouter.New(handler, compressor, zl)
	server := baseserver.New(router, zl)
	if err := server.Run(); err != nil {
		sl.Fatal("can't run service", err)
	}

}
