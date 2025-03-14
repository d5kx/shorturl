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

func (s *Storage) Ping(ctx context.Context) bool {
	cntx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(cntx); err != nil {
		s.log.Debug("DB ping: error", zap.Error(err))
		return false
	}
	s.log.Debug("DB ping: ok")
	return true
}

func (s *Storage) Bootstrap(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.log.Debug("unable to start SQL transaction", zap.Error(err))
		return e.WrapError("unable to start SQL transaction", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE public.links (
			uuid text NOT NULL,
			short_url text NOT NULL,
			original_url text NOT NULL,
			PRIMARY KEY (uuid)
		)
	`)

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			s.log.Debug("unable to rollback transaction", zap.Error(err))
		}
		s.log.Debug("unable to execute SQL transaction", zap.Error(err))
		return e.WrapError("unable to execute SQL transaction", err)
	}
	/*
		ALTER TABLE IF EXISTS public.links
		OWNER to postgres;
	*/
	return tx.Commit()
}

func (s *Storage) Save(ctx context.Context, l *link.Link) error {
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, error) {
	return "", nil
}

func (s *Storage) IsExist(ctx context.Context, shortURL string) (bool, error) {
	return false, nil
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	return nil
}
