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
	log      loggers.Logger
	db       *sql.DB
	isActive bool
}

func New(logger loggers.Logger) *Storage {
	return &Storage{
		log: logger,
	}
}

func (s *Storage) Open(connectionString string) error {
	var err error
	s.db, err = sql.Open("pgx", connectionString)
	if err != nil {
		s.log.Info("DB not open", zap.Error(err))
		return e.WrapError("can't open DB", err)
	}
	s.db.SetMaxOpenConns(100)
	s.db.SetMaxIdleConns(100)
	s.db.SetConnMaxIdleTime(time.Minute * 4)
	s.db.SetConnMaxLifetime(time.Minute * 3)

	s.log.Info("DB open", zap.String("PostgreSQL", connectionString))
	s.Ping(context.Background())
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
	s.isActive = true
	return true
}

func (s *Storage) Bootstrap(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.log.Debug("unable to start SQL transaction", zap.Error(err))
		return e.WrapError("unable to start SQL transaction", err)
	}

	query := `CREATE SCHEMA IF NOT EXISTS public`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
	}

	query = `
		CREATE TABLE IF NOT EXISTS public.links (
			uuid text NOT NULL,
			short_url text NOT NULL,
			original_url text NOT NULL,
			PRIMARY KEY (uuid)
		)`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
	}

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			s.log.Debug("unable to rollback transaction", zap.Error(err))
		} else {
			s.log.Debug("rollback transaction")
		}

		s.log.Debug("unable to execute SQL transaction", zap.Error(err))
		return e.WrapError("unable to execute SQL transaction", err)
	}

	return tx.Commit()
}

func (s *Storage) Save(ctx context.Context, l *link.Link) error {
	query := `
		INSERT INTO public.links
		(uuid, short_url, original_url)
		VALUES ($1, $2, $3)
		`
	_, err := s.db.ExecContext(ctx, query, l.UID, l.ShortURL, l.OriginalURL)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
	}
	return err
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, error) {
	return "", nil
}

func (s *Storage) IsExist(ctx context.Context, shortURL string) (bool, error) {
	query := `SELECT EXISTS( SELECT 1 FROM  public.links WHERE short_url=$1)`
	var isExist bool
	row := s.db.QueryRowContext(ctx, query, shortURL)
	err := row.Scan(&isExist)
	return isExist, err
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	return nil
}

func (s *Storage) IsActive() bool {
	return s.isActive
}
