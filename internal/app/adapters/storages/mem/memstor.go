package memstor

import (
	"bufio"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"os"

	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Storage struct {
	db       map[string]*link.Link
	log      loggers.Logger
	isActive bool
}

func (s *Storage) GetDB() map[string]*link.Link {
	return s.db
}

func New(logger loggers.Logger) *Storage {
	return &Storage{
		db:  make(map[string]*link.Link),
		log: logger,
	}
}

func (s *Storage) Save(ctx context.Context, l *link.Link) error {
	s.db[l.ShortURL] = l /*l.OriginalURL*/
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, error) {
	value, ok := s.db[shortURL]

	if !ok {
		return "", "", nil
	}
	return value.UID, value.OriginalURL, nil
}

func (s *Storage) IsExist(ctx context.Context, shortURL string) (bool, error) {
	_, ok := s.db[shortURL]
	return ok, nil
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	delete(s.db, shortURL)
	return nil
}

/*
	func (s *Storage) SaveToFile(l *link.Link) error {
		file, err := os.OpenFile(conf.GetDBFileName(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return e.WrapError("can't open file "+conf.GetDBFileName(), err)
		}
		writer := bufio.NewWriter(file)
		defer func() {
			if err := writer.Flush(); err != nil {
				s.log.Fatal("error in Flush() when saving to file", err)
			}
			if err := file.Close(); err != nil {
				s.log.Fatal("file closing error when saving to file", err)
			}
		}()

		if err = json.NewEncoder(writer).Encode(l); err != nil {
			return e.WrapError("can't encode json when saving to file", err)
		}

		return nil
	}
*/
func (s *Storage) IsActive() bool {
	return s.isActive
}

func (s *Storage) Open(name string) error {
	s.isActive = true
	return nil
}

func (s *Storage) Close() error {
	return nil
}
func (s *Storage) Bootstrap(ctx context.Context) error {
	file, err := os.OpenFile(conf.GetDBFileName(), os.O_RDONLY, 0666)
	if err != nil {
		return e.WrapError("can't open file "+conf.GetDBFileName(), err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.log.Info("file closing error when load from file", zap.Error(err))
		}
	}()
	var i int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()
		l := link.Link{}
		err = json.Unmarshal(data, &l)
		if err != nil {
			return e.WrapError("can't decode json when reading from file", err)
		}
		s.db[l.ShortURL] = &l /*.OriginalURL*/
		i++
	}

	if err := scanner.Err(); err != nil {
		s.log.Info("file scanning error when loaf from file", zap.Error(err))
	}
	s.log.Info("loaded from file", zap.Int("records", i))

	return nil
}
