package filestor

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"
	"go.uber.org/zap"
	"os"
)

type Storage struct {
	log      loggers.Logger
	file     *os.File
	isActive bool
}

func New(logger loggers.Logger) *Storage {
	return &Storage{
		log: logger,
	}
}

func (s *Storage) Open(filename string) error {
	var err error
	s.file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return e.WrapError("can't open file: "+filename, err)
	}
	s.log.Info("file open:", zap.String("filename", filename))
	s.isActive = true
	return nil
}

func (s *Storage) Close() error {
	var err error

	if err = s.file.Close(); err != nil {
		s.log.Info("file closing error when saving to file", zap.Error(err))
	}
	return err
}

func (s *Storage) Save(ctx context.Context, l *link.Link) error {
	writer := bufio.NewWriter(s.file)
	if err := json.NewEncoder(writer).Encode(l); err != nil {
		return e.WrapError("can't encode json when saving to file", err)
	}
	if err := writer.Flush(); err != nil {
		s.log.Info("error in Flush() when saving to file", zap.Error(err))
	}
	return nil
}
func (s *Storage) SaveTx(ctx context.Context, slice []*link.Link) error {
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, bool, error) {
	return "", "", false, nil
}
func (s *Storage) GetShort(ctx context.Context, shortURL string) (string, string, error) {
	return "", "", nil
}
func (s *Storage) GetUserUrls(ctx context.Context, uuid string) ([][]string, error) {
	return make([][]string, 0), nil
}
func (s *Storage) IsExist(ctx context.Context, shortURL string) (bool, error) {
	return false, nil
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	return nil
}
func (s *Storage) RemoveUrls(ctx context.Context, links []*link.Link) {}
func (s *Storage) IsActive() bool {
	return s.isActive
}
func (s *Storage) Bootstrap(ctx context.Context) error {
	return nil
}
