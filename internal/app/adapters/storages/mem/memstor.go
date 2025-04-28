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
	db       map[string]*entities.Link
	log      loggers.Logger
	isActive bool
}

func (s *Storage) GetDB() map[string]*entities.Link {
	return s.db
}

func New(logger loggers.Logger) *Storage {
	return &Storage{
		db:  make(map[string]*entities.Link),
		log: logger,
	}
}

func (s *Storage) Save(ctx context.Context, l *entities.Link) error {
	for _, v := range s.db {
		if v.OriginalURL == l.OriginalURL {
			return e.ErrSaveUniqueViolation
		}
	}

	s.db[l.ShortURL] = l /*l.OriginalURL*/
	return nil
}

func (s *Storage) SaveTx(ctx context.Context, slice []*entities.Link) error {
	for _, v := range slice {
		s.Save(ctx, v)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, bool, error) {
	value, ok := s.db[shortURL]

	if !ok {
		return "", "", false, nil
	}
	return value.UUID, value.OriginalURL, value.DeletedFlag, nil
}

func (s *Storage) GetShort(ctx context.Context, originalURL string) (string, string, error) {
	for _, v := range s.db {
		if v.OriginalURL == originalURL {
			return v.UUID, v.ShortURL, nil
		}
	}
	return "", "", nil
}

func (s *Storage) GetUserUrls(ctx context.Context, uuid string) ([][]string, error) {
	return make([][]string, 0), nil
}

func (s *Storage) LinkExist(ctx context.Context, shortURL string) (bool, error) {
	_, ok := s.db[shortURL]
	return ok, nil
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	delete(s.db, shortURL)
	return nil
}

func (s *Storage) RemoveUrls(ctx context.Context, links []*entities.Link) {}

func (s *Storage) IsActive() bool {
	return s.isActive
}

func (s *Storage) UserExist(ctx context.Context, login string) (bool, error) {
	return false, nil
}
func (s *Storage) UserSave(ctx context.Context, user *entities.User) error {
	return nil
}
func (s *Storage) UserGet(ctx context.Context, login string) (string, string, error) {
	return "", "", nil
}
func (s *Storage) Shutdown() error { return nil }
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
		l := entities.Link{}
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
