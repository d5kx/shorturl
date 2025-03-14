package usedb

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/usecases"
)

type UseCases struct {
	db usecases.DB
}

func New(db usecases.DB) *UseCases {
	return &UseCases{db: db}
}

func (u *UseCases) Ping(ctx context.Context) bool {
	return u.db.Ping(ctx)
}
