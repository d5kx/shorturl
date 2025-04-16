package storages

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/entities"
)

type LinkStorage interface {
	Save(ctx context.Context, link *link.Link) error
	SaveTx(ctx context.Context, links []*link.Link) error
	Get(ctx context.Context, shortURL string) (uuid string, originalURL string, deletedFlag bool, err error)
	GetShort(ctx context.Context, originalURL string) (uuid string, shortURL string, err error)
	GetUserUrls(ctx context.Context, uuid string) ([][]string, error)
	IsExist(ctx context.Context, shortURL string) (bool, error)
	Remove(ctx context.Context, shortURL string) error
	RemoveUrls(ctx context.Context, links []*link.Link)
}

type DB interface {
	Ping(ctx context.Context) bool
}
