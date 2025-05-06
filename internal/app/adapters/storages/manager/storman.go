package storman

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/adapters/storages"
	"github.com/d5kx/shorturl/internal/app/conf"
	"go.uber.org/zap"
)

type Storage struct {
	mdb    storages.CommonStorage
	qdb    storages.CommonStorage
	logger loggers.Logger
}

func New(memoryStorage, SQLStorage storages.CommonStorage, logger loggers.Logger) *Storage {

	return &Storage{
		mdb:    memoryStorage,
		qdb:    SQLStorage,
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
func (s *Storage) Open() error {
	var err error

	if conf.GetPostgreSQLConnectionString() != "" {
		err = s.qdb.Open(conf.GetPostgreSQLConnectionString())
	}

	if !s.qdb.IsActive() {
		err = s.mdb.Open("")
	}

	s.logger.Info("storages used",
		zap.Bool("mem", s.mdb.IsActive()),
		zap.Bool("sql", s.qdb.IsActive()),
	)
	return err
}

func (s *Storage) Close() error {
	if s.qdb.IsActive() {
		return s.qdb.Close()
	}
	return nil
}

func (s *Storage) WorkingStorage() storages.CommonStorage {
	if s.qdb.IsActive() {
		return s.qdb
	}
	return s.mdb
}
