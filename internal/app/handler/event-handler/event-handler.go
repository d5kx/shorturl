package event_handler

import (
	"github.com/d5kx/shorturl/internal/app/fetcher"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Handler struct {
	fetcher fetcher.Fetcher
}

func (h *Handler) Run() error {
	err := h.fetcher.Fetch()
	if err != nil {
		return e.WrapError("can't start fetcher", err)
	}

	return nil
}

func New(fetcher fetcher.Fetcher) Handler {
	return Handler{fetcher: fetcher}
}
