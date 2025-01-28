package main

import (
	"github.com/d5kx/shorturl/internal/app/fetcher/event-fetcher"
	"github.com/d5kx/shorturl/internal/app/handler/event-handler"
	"github.com/d5kx/shorturl/internal/app/processor/event-processor"
	"github.com/d5kx/shorturl/internal/app/storage/memory"
	"github.com/d5kx/shorturl/internal/util/err"
)

// curl -v -X POST -H "Content-Type:text/plain" -d "ya.ru" "http://localhost:8080"
func main() {
	/*
		var l = storage.Link{URL: "https://habr.com/ru/articles/457728/"}
		var l1 = storage.Link{URL: "https://habr.com/ru/articles/457728/"}


		st.Save(&l)
		st.Save(&l1)
		fmt.Println(st.GetDB())
	*/
	s := memstorage.New()
	p := event_processor.New(s)
	f := event_fetcher.New(p)

	handler := event_handler.New(f)

	if e := handler.Run(); e != nil {
		err.WrapError("can't run service", e)
	}

}
