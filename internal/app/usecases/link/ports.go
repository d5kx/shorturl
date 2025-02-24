package uselink

import (
	"github.com/d5kx/shorturl/internal/app/entities"
)

type LinkStorage interface {
	Save(*link.Link) error
	Get(shortURL string) (string, error)
	IsExist(shortURL string) (bool, error)
	Remove(shortURL string) error
}
