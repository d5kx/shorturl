package memstor

import (
	"time"

	"github.com/d5kx/shorturl/internal/app/entities"
)

type Storage struct {
	db map[string]string
	//db map[string]link.Link
}

func (s *Storage) GetDB() map[string]string {
	return s.db
}

func New() *Storage {
	return &Storage{db: make(map[string]string)}
}

func (s *Storage) Save(l *link.Link) error {
	s.db[l.ShortURL] = l.OriginalURL

	return nil
}

func (s *Storage) Get(shortURL string) (string, error) {
	value, ok := s.db[shortURL]
	//для тестирование, потом удалить
	time.Sleep(8 * time.Millisecond)
	if !ok {
		return "", nil
	}
	return value, nil
}

func (s *Storage) IsExist(shortURL string) (bool, error) {
	_, ok := s.db[shortURL]
	return ok, nil
}

func (s *Storage) Remove(shortURL string) error {
	delete(s.db, shortURL)
	return nil
}
