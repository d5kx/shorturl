package mockstor

import (
	"errors"

	"github.com/d5kx/shorturl/internal/app/link"
)

type Storage struct {
}

func New() *Storage {
	return &Storage{}
}

func (s Storage) Save(l *link.Link) (string, error) {
	if l.OriginalURL == "db_error" {
		return "", errors.New("db_error")
	}
	return "AbCdEf", nil
}

func (s Storage) Get(shortURL string) (*link.Link, error) {
	if shortURL == "AbCdEf" {
		return &(link.Link{OriginalURL: "http://ya.ru"}), nil
	}
	return nil, nil
}

func (s Storage) IsExist(shortURL string) (bool, error) {
	return false, nil
}

func (s Storage) Remove(shortURL string) error {
	return nil
}
