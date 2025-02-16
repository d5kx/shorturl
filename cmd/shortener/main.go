package main

import (
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/logger/simplelogger"
	"github.com/d5kx/shorturl/internal/app/logger/zaplogger"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/server/event-server"
	"github.com/d5kx/shorturl/internal/app/storage/memory"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X GET -H "Content-Type:text/plain" "http://localhost:8080/GlTBlr"
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration1$ -binary-path=C:\go\shorturl\cmd\shortener\shortener.exe
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration2$ -source-path=C:\go\shorturl\internal\app\processor\event-processor\event-processor_test.go
// D:\go_projects\shorturl\cmd\shortener>go vet -vettool=D:\go_projects\statictest-windows-amd64.exe ./..
func init() {
	conf.ParseFlags()
}

func main() {

	sl := simplelogger.GetInstance()

	if err := zaplogger.Init(conf.GetLoggerLevel()); err != nil {
		sl.Fatal("can't run zap logger", err)
	}
	l := zaplogger.GetInstance()

	p := eventprocessor.New(memstorage.New(), sl)
	f := eventfetcher.New(&p, l)

	server := eventserver.New(&f, l)

	if err := server.Run(); err != nil {
		sl.Fatal("can't run service", err)
	}
}
