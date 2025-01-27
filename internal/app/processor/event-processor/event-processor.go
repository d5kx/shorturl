package event_processor

import "fmt"

type Processor struct {
}

func New() Processor {
	return Processor{}
}

func (p Processor) Process() error {
	fmt.Println("Process() - OK")
	return nil
}
