package main

import (
	"fmt"

	"github.com/d5kx/shorturl/internal/app/handler/event-handler"
	"github.com/d5kx/shorturl/internal/app/storage"
	"github.com/d5kx/shorturl/internal/app/storage/memory"
)

func main() {
	handler := eventhandler.New()

	if err := handler.Run(); err != nil {
		panic(err)
	}

	fmt.Println("OK!")

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
