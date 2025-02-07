package memstorage

import (
	"github.com/d5kx/shorturl/internal/app/link"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Storage struct {
	db map[string]link.Link
}

func (s Storage) GetDB() map[string]link.Link {
	return s.db
}

func New() Storage {
	return Storage{db: make(map[string]link.Link)}
}

func (s Storage) Save(l *link.Link) (string, error) {
	var sURL string
	var err error
	isExist := true

	for isExist {
		sURL = l.ShortURL()
		isExist, err = s.IsExist(sURL)

		if err != nil {
			return "", e.WrapError("can't save link", err)
		}
	}
	s.db[sURL] = *l

	return sURL, nil
}

func (s Storage) Get(shortURL string) (*link.Link, error) {
	value, ok := s.db[shortURL]

	if !ok {
		return nil, nil
	}
	return &value, nil
}

func (s Storage) IsExist(shortURL string) (bool, error) {
	_, ok := s.db[shortURL]
	return ok, nil
}

func (s Storage) Remove(shortURL string) error {
	delete(s.db, shortURL)
	return nil
}
