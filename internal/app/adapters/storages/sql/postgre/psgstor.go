package postgre

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type Storage struct {
	log          loggers.Logger
	db           *sql.DB
	isActive     bool
	delChan      chan *entities.Link // Буферизированный канал для отложенного удаления ссылок
	forceDelChan chan struct{}       // Канал для отправки сигнала принудительного удаления ссылок из очереди

}

func New(logger loggers.Logger) *Storage {
	return &Storage{
		log:          logger,
		delChan:      make(chan *entities.Link, 256), // установим каналу удаления буфер в 256 ссылок
		forceDelChan: make(chan struct{}),
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
			id bigserial NOT NULL,
		    uuid uuid NOT NULL,
			short_url text NOT NULL,
			original_url text NOT NULL,
			deleted_flag boolean NOT NULL DEFAULT false,
			PRIMARY KEY (id)
		);
		CREATE TABLE IF NOT EXISTS public.users (
    		uuid uuid NOT NULL,
    		login text NOT NULL,
    		passwd text NOT NULL,
    		PRIMARY KEY (uuid)
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

	// запустим горутину с фоновым удалением ссылок
	go s.flushDelete()

	return nil
}

func (s *Storage) Save(ctx context.Context, l *entities.Link) error {
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

// UserSave сохраняет пользователя в БД
func (s *Storage) UserSave(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO public.users
		(uuid, login, passwd)
		VALUES ($1, $2, $3)
		`
	_, err := s.db.ExecContext(ctx, query, user.UUID, user.Login, user.PasswdHash)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return e.WrapError("unable to execute SQL query", err)
	}
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.Any("user", user),
	)
	return nil
}

func (s *Storage) SaveTx(ctx context.Context, links []*entities.Link) error {
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

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, bool, error) {
	query := `SELECT uuid, original_url, deleted_flag FROM  public.links WHERE short_url=$1`
	var (
		uuid, originalURL string
		deletedFlag       bool
	)
	row := s.db.QueryRowContext(ctx, query, shortURL)
	err := row.Scan(&uuid, &originalURL, &deletedFlag)
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return "", "", false, e.WrapError("unable to execute SQL query", err)
	}
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.String("shortURL", shortURL),
	)
	return uuid, originalURL, deletedFlag, nil
}

// UserGet достает из БД uuid и хеш пароля пользователя
func (s *Storage) UserGet(ctx context.Context, login string) (string, string, error) {
	query := `SELECT uuid, passwd FROM  public.users WHERE login=$1`
	var (
		uuid, passwd string
	)
	row := s.db.QueryRowContext(ctx, query, login)
	err := row.Scan(&uuid, &passwd)
	// другие ошибки
	if err != nil {
		s.log.Debug("unable to execute SQL query", zap.String("query", query), zap.Error(err))
		return "", "", e.WrapError("unable to execute SQL query", err)
	}
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.String("login", login),
	)
	return uuid, passwd, nil
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

// LinkExist проверяет наличие короткой ссылки в БД
func (s *Storage) LinkExist(ctx context.Context, shortURL string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM  public.links WHERE short_url=$1)`
	var isExist bool
	row := s.db.QueryRowContext(ctx, query, shortURL)
	err := row.Scan(&isExist)
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.String("shortURL", shortURL),
	)
	return isExist, err
}

// UserExist проверяет наличие логина пользователя в БД
func (s *Storage) UserExist(ctx context.Context, login string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM  public.users WHERE login=$1)`
	var isExist bool
	row := s.db.QueryRowContext(ctx, query, login)
	err := row.Scan(&isExist)
	s.log.Debug("execute SQL query",
		zap.String("query", query),
		zap.String("login", login),
	)
	return isExist, err
}

// RemoveUrls ставит ссылки в очередь на асинхронное удаление.
func (s *Storage) RemoveUrls(ctx context.Context, urls []*entities.Link) {
	// отправляем ссылки в очередь на сохранение
	for _, v := range urls {
		s.delChan <- v
	}
}

// flushDelete с определённым интервалом помечает записи в БД как удаленные
func (s *Storage) flushDelete() {
	// будем сохранять сообщения, накопленные за последние 15 секунд
	ticker := time.NewTicker(15 * time.Second)
	//слайс для ссылок для удаления
	var links []*entities.Link

	for {
		select {
		//в канал поступила ссылка на удаление
		case data := <-s.delChan:
			//добавляем ссылку в слайс ссылок для удаления
			links = append(links, data)

		// сработал таймер
		case <-ticker.C:
			//s.log.Debug("deleted_ticker", zap.String("time", time.Now().String()))
			s.deleteLinksFromSlice(links)

		// пришел сигнал принудительного срабатывания, например из Shutdown
		case <-s.forceDelChan:
			//s.log.Debug("deleted_forced", zap.String("time", time.Now().String()))
			s.deleteLinksFromSlice(links)
		}
	}
}

// deleteLinksFromSlice помечает записи содержащиеся в links как удаленные в БД
func (s *Storage) deleteLinksFromSlice(links []*entities.Link) {
	if len(links) == 0 {
		return
	}
	var (
		values []string // слайс параметров
		args   []any    // слайс аргументов
	)
	// заполняем слайсы параметров и аргументов
	for i, v := range links {
		base := i * 2
		values = append(values, fmt.Sprintf("($%d,$%d)", base+1, base+2))
		args = append(args, v.UUID, v.ShortURL)
	}
	// составляем строку запроса
	query := `
				UPDATE public.links AS t
					SET deleted_flag = true
						FROM (VALUES` + strings.Join(values, ",") +
		`) AS v(uuid, short_url)
					WHERE t.uuid = v.uuid AND t.short_url = v.short_url;`

	// выполняем запрос на обновление данных в БД
	_, err := s.db.ExecContext(context.Background(), query, args...)
	if err != nil {
		s.log.Debug("unable to execute SQL query",
			zap.String("query", query),
			zap.Error(err))
	}
	if err == nil {
		s.log.Debug("execute SQL query",
			zap.String("query", query),
			zap.Any("args", args),
		)
	}
	// очистим слайс после удаления
	links = nil
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	return nil
}

func (s *Storage) IsActive() bool {
	return s.isActive
}

func (s *Storage) Shutdown(ctx context.Context) error {
	// посылаем сигнал на удаление ссылок из очереди и очищение очереди
	s.forceDelChan <- struct{}{}
	return nil
}
