package main

import (
	"log"

	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/server/event-server"
	"github.com/d5kx/shorturl/internal/app/storage/memory"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "ya.ru" "http://localhost:8080"
// curl -v -X GET -H "Content-Type:text/plain" "http://localhost:8080/GlTBlr"
// shortenertest-windows-amd64 -test.v -test.run=^TestIteration1$ -binary-path=C:\go\shorturl\cmd\shortener\shortener.exe
// D:\go_projects\shorturl\cmd\shortener>go vet -vettool=D:\go_projects\statictest-windows-amd64.exe ./..
func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	p := eventprocessor.New(memstorage.New())
	p.SetAddress("localhost:8080")

	f := eventfetcher.New("localhost:8080")
	f.Router.Post(`/`, p.Post)
	f.Router.Get(`/{id}`, p.Get)
	f.Router.NotFound(p.BadRequest)
	f.Router.MethodNotAllowed(p.BadRequest)
	//f.AddHandler(`/`, &p)

	server := eventserver.New(&f)

	if err := server.Run(); err != nil {
		log.Fatal("can't run service", err)
	}
}
