package main

import (
	"fmt"

	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/handler/event-handler"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/storage"
	"github.com/d5kx/shorturl/internal/app/storage/memory"
)

func main() {
	f := event_fetcher.New()
	p := event_processor.New()

	handler := eventhandler.New(f, p)

	if err := handler.Run(); err != nil {
		panic(err)
	}

	var l = storage.Link{URL: "https://habr.com/ru/articles/457728/"}
	var l1 = storage.Link{URL: "https://habr.com/ru/articles/457728/"}

	//s, _ := l.ShortURL()

	//fmt.Printf(s)
	st := memstorage.New()
	st.Save(&l)
	st.Save(&l1)
	fmt.Println(st.GetDB())
	//var db map[string]storage.Link
}
