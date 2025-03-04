package main

import (
	"github.com/d5kx/shorturl/internal/app/adapters/compress/gzip"
	"github.com/d5kx/shorturl/internal/app/adapters/http/handlers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/routers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/servers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/simple"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/zap"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/mem"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/link"
	"github.com/d5kx/shorturl/internal/util/generators/basegen"
)

// curl -v -X POST -H "Content-Type:text/plain" -H -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:application/json"  -H "Accept-Encoding:gzip" --output "-" -d "{\"url\": \"https://practicum.yandex.ru\"}" "http://localhost:9090/api/shorten"
// curl -v -X GET -H "Content-Type:text/plain" -H "Accept-Encoding:gzip" --output "-" "http://localhost:8080/GlTBlr"
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration1$ -binary-path=C:\go\shorturl\cmd\shortener\shortener.exe
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration2$ -source-path=C:\go\shorturl\internal\app\handlers\event-handlers\event-processor_test.go
// mockgen -destination=internal/app/adapters/storages/gomock/gomockstor.go -package=gomockstor github.com/d5kx/shorturl/internal/app/usecases/link LinkStorage
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
	err = m.LoadFromFile()
	if err != nil {
		sl.Fatal("can't load DB from file", err)
	}

	u := uselink.New(m, basegen.New(), zl)
	c := gzipc.New(zl)
	p := basehandler.New(u, zl)
	f := baserouter.New(p, c, zl)

	server := baseserver.New(f, zl)
	if err := server.Run(); err != nil {
		sl.Fatal("can't run service", err)
	}

}
