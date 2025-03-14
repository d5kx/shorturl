package usecases

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/entities"
)

type LinkStorage interface {
	Save(ctx context.Context, link *link.Link) error
	Get(ctx context.Context, shortURL string) (string, error)
	IsExist(ctx context.Context, shortURL string) (bool, error)
	Remove(ctx context.Context, shortURL string) error
}

type DB interface {
	Ping(ctx context.Context) bool
}
