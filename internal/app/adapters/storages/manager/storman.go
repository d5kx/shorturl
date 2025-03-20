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
	fdb    storages.ManagedStorage
	qdb    storages.ManagedStorage
	logger loggers.Logger
}

func New(mstor, fstor, qstor storages.ManagedStorage, logger loggers.Logger) *Storage {

	return &Storage{
		mdb:    mstor,
		fdb:    fstor,
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

	if s.fdb.IsActive() {
		err = s.fdb.Bootstrap(context.Background())
		if err != nil {
			s.logger.Info("can't bootstrap file db", zap.Error(err))
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

func (s *Storage) Save(ctx context.Context, l *link.Link) error {
	var err error
	if s.qdb.IsActive() {
		return s.qdb.Save(ctx, l)
	}

	if s.fdb.IsActive() {
		err = s.fdb.Save(ctx, l)
	}
	err = s.mdb.Save(ctx, l)

	return err
}
func (s *Storage) SaveTx(ctx context.Context, links []*link.Link) error {
	var err error
	if s.qdb.IsActive() {
		return s.qdb.SaveTx(ctx, links)
	}

	if s.fdb.IsActive() {
		err = s.fdb.SaveTx(ctx, links)
	}
	err = s.mdb.SaveTx(ctx, links)

	return err
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, error) {
	if s.qdb.IsActive() {
		return s.qdb.Get(ctx, shortURL)
	}
	return s.mdb.Get(ctx, shortURL)
}

func (s *Storage) IsExist(ctx context.Context, shortURL string) (bool, error) {
	if s.qdb.IsActive() {
		return s.qdb.IsExist(ctx, shortURL)
	}
	return s.mdb.IsExist(ctx, shortURL)
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	return nil
}

func (s *Storage) IsActive() bool {
	return true
}

func (s *Storage) Open(name string) error {
	var err error

	if conf.GetPostgreSQLConnectionString() != "" {
		err = s.qdb.Open(conf.GetPostgreSQLConnectionString())
		if err != nil {
			s.logger.Info("can't connect to PostgreSQL db", zap.Error(err))
		}
	}

	if !s.qdb.IsActive() && conf.GetDBFileName() != "" {
		err = s.fdb.Open(conf.GetDBFileName())
		if err != nil {
			s.logger.Info("can't open DB file", zap.Error(err))
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
		zap.Bool("file", s.fdb.IsActive()),
		zap.Bool("sql", s.qdb.IsActive()),
	)
	return err
}

func (s *Storage) Close() error {
	var err error

	if conf.GetPostgreSQLConnectionString() != "" {
		err = s.qdb.Close()
		if err != nil {
			s.logger.Info("can't close connection to PostgreSQL db", zap.Error(err))
		}
	}

	if conf.GetDBFileName() != "" {
		err = s.fdb.Close()
		if err != nil {
			s.logger.Info("can't close connection to file db", zap.Error(err))
		}
	}
	return err
}
