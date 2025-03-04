package mockstor

import (
	"errors"

	"github.com/d5kx/shorturl/internal/app/entities"
)

type Storage struct {
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Save(l *link.Link) error {
	if l.OriginalURL == "db_error" {
		return errors.New("db_error")
	}
	return nil
}

func (s *Storage) Get(shortURL string) (string, error) {
	if shortURL == "AbCdEf" {
		return "http://ya.ru", nil
	}
	return "", nil
}

func (s *Storage) IsExist(shortURL string) (bool, error) {
	return false, nil
}

func (s *Storage) Remove(shortURL string) error {
	return nil
}

func (s *Storage) SaveToFile(filename string) error {
	return nil
}
