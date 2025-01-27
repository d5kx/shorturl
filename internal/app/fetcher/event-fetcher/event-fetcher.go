package event_fetcher

import "fmt"

type Fetcher struct {
}

func New() Fetcher {
	return Fetcher{}
}

func (f Fetcher) Fetch() error {
	fmt.Println("Fetch() - OK")
	return nil
}
