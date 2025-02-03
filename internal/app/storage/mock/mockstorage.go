package mockstorage

import (
	"github.com/d5kx/shorturl/internal/app/link"
)

type Storage struct {
}

func New() Storage {

	return Storage{}
}

func (s Storage) Save(l *link.Link) (string, error) {

	return "AbCdEf", nil

}

func (s Storage) Get(shortURL string) (*link.Link, error) {

	return &(link.Link{URL: "http://ya.ru"}), nil
}

func (s Storage) IsExist(shortURL string) (bool, error) {

	return false, nil
}

func (s Storage) Remove(shortURL string) error {

	return nil
}
