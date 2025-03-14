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
	db     usecases.LinkStorage
	logger loggers.Logger
	gen    generators.Generator
}

func New(storage usecases.LinkStorage, generator generators.Generator, logger loggers.Logger) *UseCases {
	return &UseCases{
		db:     storage,
		logger: logger,
		gen:    generator,
	}
}

func (u *UseCases) Save(ctx context.Context, originalURL string) (string, error) {
	var shortURL string
	var err error

	isExist := true
	for isExist {
		shortURL = u.gen.ShortURL()
		isExist, err = u.db.IsExist(ctx, shortURL)
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
	err = u.db.Save(ctx, &l)

	if err != nil {
		u.logger.Debug("Save() database error", zap.Error(err))
		return "", e.WrapError("database error", err)
	}
	return l.ShortURL, err
}

func (u *UseCases) Get(ctx context.Context, shortURL string) (*link.Link, error) {
	originalURL, err := u.db.Get(ctx, shortURL)
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
