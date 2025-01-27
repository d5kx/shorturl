package memstorage

import (
	"github.com/d5kx/shorturl/internal/app/storage"
	"github.com/d5kx/shorturl/internal/util/err"
)

type Storage struct {
	db map[string]storage.Link
}

func (s *Storage) GetDB() map[string]storage.Link {
	return s.db
}

func New() Storage {
	return Storage{db: make(map[string]storage.Link)}
}

func (s *Storage) Save(l *storage.Link) error {

	var sUrl string
	var e error
	isExist := true

	for isExist {
		sUrl, e = l.ShortURL()
		isExist, e = s.IsExist(sUrl)

		if e != nil {
			return err.WrapError("can't save link", e)
		}
	}
	s.db[sUrl] = *l

	return nil

}
func (s *Storage) Get(shortURL string) (*storage.Link, error) {
	value, ok := s.db[shortURL]

	if ok {
		return &value, nil
	}
	return nil, nil

}
func (s *Storage) IsExist(shortURL string) (bool, error) {
	_, ok := s.db[shortURL]
	return ok, nil
}
func (s *Storage) Remove(shortURL string) error {
	delete(s.db, shortURL)
	return nil

}
