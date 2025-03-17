package uselink

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/app/usecases"
	"github.com/d5kx/shorturl/internal/util/e"
	"github.com/d5kx/shorturl/internal/util/generators"

	"go.uber.org/zap"
)

type UseCases struct {
	mdb    usecases.LinkStorage
	fdb    usecases.LinkStorage
	qdb    usecases.LinkStorage
	logger loggers.Logger
	gen    generators.Generator
}

func New(mstor, fstor, qstor usecases.LinkStorage, generator generators.Generator, logger loggers.Logger) *UseCases {
	logger.Info("storage activity",
		zap.Bool("mem", mstor.IsActive()),
		zap.Bool("file", fstor.IsActive()),
		zap.Bool("sql", qstor.IsActive()),
	)
	return &UseCases{
		mdb:    mstor,
		fdb:    fstor,
		qdb:    qstor,
		logger: logger,
		gen:    generator,
	}
}

func (u *UseCases) Save(ctx context.Context, originalURL string) (string, error) {
	var (
		shortURL string
		err      error
	)

	isExist := true
	for isExist {
		shortURL = u.gen.ShortURL()
		if u.qdb.IsActive() {
			isExist, err = u.qdb.IsExist(ctx, shortURL)
		} else {
			isExist, err = u.mdb.IsExist(ctx, shortURL)
		}
		if err != nil {
			u.logger.Debug("IsExist() database error", zap.String("sURL", shortURL), zap.Error(err))
			return "", e.WrapError("database error", err)
		}
	}

	var l = link.Link{
		UID:         u.gen.UUID(),
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}

	if u.qdb.IsActive() {
		err = u.qdb.Save(ctx, &l)
	} else {
		if u.fdb.IsActive() {
			err = u.fdb.Save(ctx, &l)
		}
		err = u.mdb.Save(ctx, &l)
	}

	if err != nil {
		u.logger.Debug("Save() database error", zap.Error(err))
		return "", e.WrapError("database error", err)
	}
	return l.ShortURL, err
}

func (u *UseCases) Get(ctx context.Context, shortURL string) (*link.Link, error) {
	var (
		originalURL string
		err         error
	)

	if u.qdb.IsActive() {
		originalURL, err = u.qdb.Get(ctx, shortURL)
	} else {
		originalURL, err = u.mdb.Get(ctx, shortURL)
	}

	if err != nil {
		u.logger.Debug("Get() database error", zap.String("sURL", shortURL), zap.Error(err))
		return nil, e.WrapError("database error", err)
	}

	if originalURL == "" {
		u.logger.Debug("short link does not exist in the database", zap.String("short", shortURL))
		return nil, nil
	}

	return &link.Link{
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}, nil
}
