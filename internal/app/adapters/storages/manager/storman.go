package storman

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/adapters/storages"
	"github.com/d5kx/shorturl/internal/app/conf"
	"go.uber.org/zap"

	"github.com/d5kx/shorturl/internal/app/entities"
)

type Storage struct {
	mdb    storages.ManagedStorage
	qdb    storages.ManagedStorage
	logger loggers.Logger
}

func New(mstor, qstor storages.ManagedStorage, logger loggers.Logger) *Storage {

	return &Storage{
		mdb:    mstor,
		qdb:    qstor,
		logger: logger,
	}
}

func (s *Storage) Bootstrap(ctx context.Context) error {
	var err error

	if s.qdb.IsActive() {
		err = s.qdb.Bootstrap(context.Background())
		if err != nil {
			s.logger.Info("can't bootstrap PostgreSQL db", zap.Error(err))
		}
	}

	if s.mdb.IsActive() {
		err = s.mdb.Bootstrap(ctx)
		if err != nil {
			s.logger.Info("can't bootstrap memory DB from file", zap.Error(err))
		}
	}
	return err
}

func (s *Storage) Save(ctx context.Context, l *entities.Link) error {
	if s.qdb.IsActive() {
		return s.qdb.Save(ctx, l)
	}
	if s.mdb.IsActive() {
		return s.mdb.Save(ctx, l)
	}
	return nil
}

func (s *Storage) SaveTx(ctx context.Context, links []*entities.Link) error {
	if s.qdb.IsActive() {
		return s.qdb.SaveTx(ctx, links)
	}
	if s.mdb.IsActive() {
		return s.mdb.SaveTx(ctx, links)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, bool, error) {
	if s.qdb.IsActive() {
		return s.qdb.Get(ctx, shortURL)
	}
	return s.mdb.Get(ctx, shortURL)
}

func (s *Storage) GetShort(ctx context.Context, originalURL string) (string, string, error) {
	if s.qdb.IsActive() {
		return s.qdb.GetShort(ctx, originalURL)
	}
	return s.mdb.GetShort(ctx, originalURL)
}

func (s *Storage) GetUserUrls(ctx context.Context, uuid string) ([][]string, error) {
	if s.qdb.IsActive() {
		return s.qdb.GetUserUrls(ctx, uuid)
	}
	return s.mdb.GetUserUrls(ctx, uuid)
}

func (s *Storage) LinkExist(ctx context.Context, shortURL string) (bool, error) {
	if s.qdb.IsActive() {
		return s.qdb.LinkExist(ctx, shortURL)
	}
	return s.mdb.LinkExist(ctx, shortURL)
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	return nil
}
func (s *Storage) RemoveUrls(ctx context.Context, links []*entities.Link) {
	if s.qdb.IsActive() {
		s.qdb.RemoveUrls(ctx, links)
		return
	}
	s.mdb.RemoveUrls(ctx, links)
}

func (s *Storage) UserExist(ctx context.Context, login string) (bool, error) {
	if s.qdb.IsActive() {
		return s.qdb.UserExist(ctx, login)
	}
	return s.mdb.UserExist(ctx, login)
}

func (s *Storage) UserSave(ctx context.Context, user *entities.User) error {
	if s.qdb.IsActive() {
		return s.qdb.UserSave(ctx, user)
	}
	return s.mdb.UserSave(ctx, user)
}
func (s *Storage) UserGet(ctx context.Context, login string) (string, string, error) {
	if s.qdb.IsActive() {
		return s.qdb.UserGet(ctx, login)
	}
	return s.mdb.UserGet(ctx, login)
}
func (s *Storage) IsActive() bool {
	return true
}
func (s *Storage) Shutdown(ctx context.Context) error {
	var err error
	if s.qdb.IsActive() {
		err = s.qdb.Shutdown(ctx)
	}
	if s.mdb.IsActive() {
		err = s.mdb.Shutdown(ctx)
	}
	return err
}
func (s *Storage) Open(name string) error {
	var err error

	if conf.GetPostgreSQLConnectionString() != "" {
		err = s.qdb.Open(conf.GetPostgreSQLConnectionString())
		if err != nil {
			s.logger.Info("can't connect to PostgreSQL db", zap.Error(err))
		}
	}

	if !s.qdb.IsActive() {
		err = s.mdb.Open("")
		if err != nil {
			s.logger.Info("can't open memory DB", zap.Error(err))
		}
	}

	s.logger.Info("storages used",
		zap.Bool("mem", s.mdb.IsActive()),
		zap.Bool("sql", s.qdb.IsActive()),
	)
	return err
}

func (s *Storage) Close() error {
	if s.qdb.IsActive() {
		err := s.qdb.Close()
		if err != nil {
			s.logger.Info("can't close connection to PostgreSQL db", zap.Error(err))
		}
		return err
	}

	return nil
}
