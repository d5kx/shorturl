package storages

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/usecases"
)

type ManagedStorage interface {
	storages.LinkStorage
	Open(string) error
	Close() error
	Bootstrap(ctx context.Context) error
	IsActive() bool
}
