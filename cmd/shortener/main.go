package main

import (
	"fmt"
	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/handler/event-handler"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/storage"
	"github.com/d5kx/shorturl/internal/app/storage/memory"
	"github.com/d5kx/shorturl/internal/util/e"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "ya.ru" "http://localhost:8080"
func main() {

	var l = storage.Link{URL: "https://habr.com/ru/articles/457728/"}
	s := memstorage.New()
	s.Save(&l)

	fmt.Println(s.GetDB())

	p := event_processor.New(s)
	f := event_fetcher.New(p)

	handler := event_handler.New(f)

	if err := handler.Run(); err != nil {
		e.WrapError("can't run service", err)
	}

}
