package fetcher

import (
	"github.com/d5kx/shorturl/internal/app/processor"
)

type Fetcher interface {
	Fetch() error
	AddHandler(string, processor.Processor)
}
