package postgre

import (
	"context"
	"database/sql"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
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
		s.log.Info("DB not close", zap.Error(err))
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
	query := `
		CREATE SCHEMA IF NOT EXISTS public;
		CREATE TABLE IF NOT EXISTS public.links (
			uuid text NOT NULL,
			short_url text NOT NULL,
			original_url text NOT NULL,
			deleted_flag boolean NOT NULL DEFAULT false
		);
		CREATE UNIQUE INDEX IF NOT EXISTS original ON public.links (original_url);`

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return e.WrapError("unable to execute SQL query", err)
	}
	s.log.Debug("bootstrap: execute SQL query",
		zap.String("query", "CREATE SCHEMA..., CREATE TABLE..., CREATE UNIQUE INDEX..."),
	)

	return nil
}

func (s *Storage) Save(ctx context.Context, l *link.Link) error {
	query := `
		INSERT INTO public.links
		(uuid, short_url, original_url)
		VALUES ($1, $2, $3)
		`
	_, err := s.db.ExecContext(ctx, query, l.UUID, l.ShortURL, l.OriginalURL)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return e.WrapError("unable to execute SQL query", err)
	}
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.Any("link", l),
	)
	return nil
}

func (s *Storage) SaveTx(ctx context.Context, links []*link.Link) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.log.Debug("unable to start SQL transaction", zap.Error(err))
		return e.WrapError("unable to start SQL transaction", err)
	}

	for _, v := range links {
		err = s.Save(ctx, v)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.log.Debug("unable to rollback transaction", zap.Error(err))
			} else {
				s.log.Debug("rollback transaction")
			}
			s.log.Debug("unable to execute SQL transaction", zap.Error(err))
			return e.WrapError("unable to execute SQL transaction", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		s.log.Debug("unable to execute SQL transaction", zap.Error(err))
		return e.WrapError("unable to execute SQL transaction", err)
	}
	s.log.Debug("execute SQL transaction")
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, error) {
	query := `SELECT uuid, original_url FROM  public.links WHERE short_url=$1`
	var uuid, originalURL string
	row := s.db.QueryRowContext(ctx, query, shortURL)
	err := row.Scan(&uuid, &originalURL)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return "", "", e.WrapError("unable to execute SQL query", err)
	}
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.String("shortURL", shortURL),
	)
	return uuid, originalURL, nil
}

func (s *Storage) GetShort(ctx context.Context, originalURL string) (string, string, error) {
	query := `SELECT uuid, short_url FROM  public.links WHERE original_url=$1`
	var uuid, shortURL string
	row := s.db.QueryRowContext(ctx, query, originalURL)
	err := row.Scan(&uuid, &shortURL)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return "", "", e.WrapError("unable to execute SQL query", err)
	}
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.String("originalURL", originalURL),
	)
	return uuid, shortURL, nil
}

func (s *Storage) GetUserUrls(ctx context.Context, uuid string) ([][]string, error) {
	var (
		rows *sql.Rows
		err  error
	)
	query := `SELECT  short_url, original_url FROM  public.links WHERE uuid=$1`
	rows, err = s.db.QueryContext(ctx, query, uuid)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return nil, e.WrapError("unable to execute SQL query", err)
	}
	defer rows.Close()
	result := make([][]string, 0)
	for rows.Next() {
		var shortURL, originalURL string
		if err = rows.Scan(&shortURL, &originalURL); err != nil {
			s.log.Debug("unable to scan SQL query result", zap.Error(err))
			return nil, e.WrapError("unable to scan SQL query result", err)
		}
		result = append(result, []string{shortURL, originalURL})
	}
	err = rows.Close()
	if err != nil {
		s.log.Debug("unable to scan SQL query result", zap.Error(err))
		return nil, e.WrapError("unable to close rows SQL query result", err)
	}

	if err := rows.Err(); err != nil {
		s.log.Debug("unable to scan SQL query result", zap.Error(err))
		return nil, e.WrapError("unable to scan SQL query result", err)
	}
	return result, nil
}

func (s *Storage) IsExist(ctx context.Context, shortURL string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM  public.links WHERE short_url=$1)`
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
