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
	//var l = storage.Link{URL: "https://habr.com/ru/articles/457728/"}
	s := memstorage.New()
	//s.Save(&l)

	//fmt.Println(s.GetDB())

	p := eventprocessor.New(s)
	f := eventfetcher.New(p)

	handler := eventserver.New(f)

	if err := handler.Run(); err != nil {
		log.Fatal("can't run service", err)
	}
}
