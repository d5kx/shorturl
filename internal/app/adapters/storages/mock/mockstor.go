package mockstor

import (
	"context"
	"errors"

	"github.com/d5kx/shorturl/internal/app/entities"
)

type Storage struct {
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Save(ctx context.Context, l *link.Link) error {
	if l.OriginalURL == "db_error" {
		return errors.New("db_error")
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, error) {
	if shortURL == "AbCdEf" {
		return "http://ya.ru", nil
	}
	return "", nil
}

func (s *Storage) IsExist(ctx context.Context, shortURL string) (bool, error) {
	return false, nil
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	return nil
}
