package storages

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/usecases"
)

type CommonStorage interface {
	storages.LinkStorage
	storages.UserStorage
	Open(string) error
	Close() error
	IsActive() bool
	Bootstrap(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type StorageManager interface {
	Open() error
	Close() error
	Bootstrap(ctx context.Context) error
	Shutdown(ctx context.Context) error
	WorkingStorage() CommonStorage
}
