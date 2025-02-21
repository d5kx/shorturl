package main

import (
	"github.com/d5kx/shorturl/internal/app/adapters/http/routers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/http/servers/base"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/simple"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers/zap"
	"github.com/d5kx/shorturl/internal/app/adapters/storages/mem"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/handlers/base"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:application/json" -d "{\"url\": \"https://practicum.yandex.ru\"}" "http://localhost:8080/api/shorten"
// curl -v -X GET -H "Content-Type:text/plain" "http://localhost:8080/GlTBlr"
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration1$ -binary-path=C:\go\shorturl\cmd\shortener\shortener.exe
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration2$ -source-path=C:\go\shorturl\internal\app\handlers\event-handlers\event-processor_test.go
func init() {
	conf.ParseFlags()
}

func main() {
	sl := simplelogger.New()

	zl, err := zaplogger.New()
	if err != nil {
		sl.Fatal("can't run zap loggers", err)
	}

	p := basehandler.New(memstor.New(), zl)
	f := baserouter.New(p, zl)

	server := baseserver.New(f, zl)
	if err := server.Run(); err != nil {
		sl.Fatal("can't run service", err)
	}
}
