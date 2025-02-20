package main

import (
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/log/simple"
	"github.com/d5kx/shorturl/internal/app/log/zap"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/routers/base"
	"github.com/d5kx/shorturl/internal/app/servers/base"
	"github.com/d5kx/shorturl/internal/app/stor/mem"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X POST -H "Content-Type:application/json" -d "{\"url\": \"https://practicum.yandex.ru\"}" "http://localhost:8080/api/shorten"
// curl -v -X GET -H "Content-Type:text/plain" "http://localhost:8080/GlTBlr"
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration1$ -binary-path=C:\go\shorturl\cmd\shortener\shortener.exe
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration2$ -source-path=C:\go\shorturl\internal\app\processor\event-processor\event-processor_test.go
func init() {
	conf.ParseFlags()
}

func main() {
	sl := simplelogger.New()

	zl, err := zaplogger.New()
	if err != nil {
		sl.Fatal("can't run zap log", err)
	}

	p := eventprocessor.New(memstor.New(), zl)
	f := baserouter.New(p, zl)

	server := baseserver.New(f, zl)
	if err := server.Run(); err != nil {
		sl.Fatal("can't run service", err)
	}
}
