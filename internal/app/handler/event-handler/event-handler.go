package event_handler

import (
	"github.com/d5kx/shorturl/internal/app/fetcher"
	"github.com/d5kx/shorturl/internal/util/err"
)

type Handler struct {
	fetcher fetcher.Fetcher
}

func (h *Handler) Run() error {
	e := h.fetcher.Fetch()
	if e != nil {
		return err.WrapError("can't start fetcher", e)
	}

	return nil
}

func New(fetcher fetcher.Fetcher) Handler {
	return Handler{fetcher: fetcher}
}
