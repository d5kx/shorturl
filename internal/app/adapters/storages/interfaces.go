package storages

import (
	"github.com/d5kx/shorturl/internal/app/entities"
)

type Storage interface {
	Save(*link.Link) (string, error)
	Get(shortURL string) (*link.Link, error)
	IsExist(shortURL string) (bool, error)
	Remove(shortURL string) error
}
