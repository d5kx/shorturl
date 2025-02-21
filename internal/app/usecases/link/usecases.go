package usecases

import (
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"

	"go.uber.org/zap"
)

type UseCases struct {
	db     LinkStorage
	logger loggers.Logger
}

func New(storage LinkStorage, logger loggers.Logger) *UseCases {
	return &UseCases{
		db:     storage,
		logger: logger,
	}
}

func (u *UseCases) Save(OriginalURL string) (string, error) {
	var sURL string
	var err error

	isExist := true
	for isExist {
		sURL = link.ShortURL()
		isExist, err = u.db.IsExist(sURL)
		if err != nil {
			u.logger.Debug("database error", zap.Error(err))
			return "", e.WrapError("database error", err)
		}
	}
	var l = link.Link{
		OriginalURL: OriginalURL,
		ShortURL:    sURL,
	}

	err = u.db.Save(&l)
	if err != nil {
		u.logger.Debug("database error", zap.Error(err))
		return "", e.WrapError("database error", err)
	}
	return sURL, err
}

func (u *UseCases) Get(shortURL string) (*link.Link, error) {
	originalURL, err := u.db.Get(shortURL)
	if err != nil {
		u.logger.Debug("database error", zap.Error(err))
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
