package eventhandler

import (
	"github.com/d5kx/shorturl/internal/app/fetcher"
	"github.com/d5kx/shorturl/internal/app/processor"
)

type Handler struct {
	fetcher   fetcher.Fetcher
	processor processor.Processor
}

func (h *Handler) Run() error {
	err := h.fetcher.Fetch()
	if err != nil {
		return err
	}

	err = h.processor.Process()
	if err != nil {
		return err
	}
	return nil
}

func New(fetcher fetcher.Fetcher, processor processor.Processor) Handler {
	return Handler{
		fetcher:   fetcher,
		processor: processor,
	}
}
