package storages

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/usecases"
)

type ManagedStorage interface {
	storages.LinkStorage
	storages.UserStorage
	Open(string) error
	Close() error
	IsActive() bool
	Bootstrap(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
