package postgre

import (
	"context"
	"database/sql"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
)

type Storage struct {
	log loggers.Logger
	db  *sql.DB
}

func New(logger loggers.Logger) *Storage {
	return &Storage{
		log: logger,
	}
}

func (s *Storage) Connect(connectionString string) error {
	var err error
	s.db, err = sql.Open("pgx", connectionString)
	if err != nil {
		s.log.Debug("DB not open", zap.Error(err))
		return e.WrapError("can't open DB", err)
	}
	s.log.Debug("DB open", zap.String("PostgreSQL", connectionString))
	return nil
}

func (s *Storage) Close() error {
	err := s.db.Close()
	if err != nil {
		s.log.Debug("DB not close", zap.Error(err))
		return e.WrapError("can't close DB", err)
	}
	return nil
}

func (s *Storage) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		s.log.Debug("DB ping: error", zap.Error(err))
		return false
	}
	s.log.Debug("DB ping: ok")
	return true
}

func (s *Storage) Save(l *link.Link) error {
	return nil
}

func (s *Storage) Get(shortURL string) (string, error) {
	return "", nil
}

func (s *Storage) IsExist(shortURL string) (bool, error) {
	return false, nil
}

func (s *Storage) Remove(shortURL string) error {
	return nil
}
