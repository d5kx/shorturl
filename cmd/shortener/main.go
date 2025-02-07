package main

import (
	"log"

	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/server/event-server"
	"github.com/d5kx/shorturl/internal/app/storage/memory"

	"github.com/d5kx/shorturl/cmd/shortener/conf"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "http://ya.ru" "http://localhost:8080"
// curl -v -X GET -H "Content-Type:text/plain" "http://localhost:8080/GlTBlr"
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration1$ -binary-path=C:\go\shorturl\cmd\shortener\shortener.exe
// D:\go_projects\shorturl\cmd\shortener>go vet -vettool=D:\go_projects\statictest-windows-amd64.exe ./..
func init() {
	conf.Config.ParseFlags()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	p := eventprocessor.New(memstorage.New())
	f := eventfetcher.New(conf.Config.GetServAdr(), &p)

	server := eventserver.New(&f)
	if err := server.Run(); err != nil {
		log.Fatal("can't run service", err)
	}
}
