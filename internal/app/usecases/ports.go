package storages

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/entities"
)

type LinkStorage interface {
	Save(ctx context.Context, link *entities.Link) error
	SaveTx(ctx context.Context, links []*entities.Link) error
	Get(ctx context.Context, shortURL string) (uuid string, originalURL string, deletedFlag bool, err error)
	GetShort(ctx context.Context, originalURL string) (uuid string, shortURL string, err error)
	GetUserUrls(ctx context.Context, uuid string) ([][]string, error)
	LinkExist(ctx context.Context, shortURL string) (bool, error)
	Remove(ctx context.Context, shortURL string) error
	RemoveUrls(ctx context.Context, links []*entities.Link)
}

type UserStorage interface {
	UserExist(ctx context.Context, login string) (bool, error)
	UserSave(ctx context.Context, user *entities.User) error
	UserGet(ctx context.Context, login string) (uuid string, passwd string, err error)
}

type DB interface {
	Ping(ctx context.Context) bool
}
