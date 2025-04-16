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
	db  storages.LinkStorage
	log loggers.Logger
	gen generators.Generator
}

func New(storage storages.LinkStorage, generator generators.Generator, logger loggers.Logger) *UseCases {
	return &UseCases{
		db:  storage,
		log: logger,
		gen: generator,
	}
}

// Save сохраняет в базу данных оригинальный url и идентификатор пользователя,
// генерирует сокращенный адрес
func (u *UseCases) Save(ctx context.Context, originalURL string, userId string) (string, error) {
	var (
		shortURL string
		err      error
	)

	isExist := true
	for isExist {
		shortURL = u.gen.ShortURL()
		isExist, err = u.db.IsExist(ctx, shortURL)
		if err != nil {
			u.log.Debug("IsExist() database error", zap.String("sURL", shortURL), zap.Error(err))
			return "", e.WrapError("database error", err)
		}
	}

	var l = link.Link{
		UUID:        userId, /*u.gen.UUID()*/
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}

	err = u.db.Save(ctx, &l)

	if err != nil {
		u.log.Debug("Save() database error", zap.Error(err))
		return "", e.WrapError("database error", err)
	}
	return l.ShortURL, err
}

func (u *UseCases) SaveTx(ctx context.Context, slice []string, userId string) ([]string, error) {
	var (
		shortURL string
		err      error
		links    []*link.Link
		shorts   []string
	)
	for _, v := range slice {
		isExist := true
		for isExist {
			shortURL = u.gen.ShortURL()
			isExist, err = u.db.IsExist(ctx, shortURL)
			if err != nil {
				u.log.Debug("IsExist() database error", zap.String("sURL", shortURL), zap.Error(err))
				return nil, e.WrapError("database error", err)
			}
		}
		links = append(links, &link.Link{
			UUID:        userId,
			OriginalURL: v,
			ShortURL:    shortURL,
		})
		shorts = append(shorts, shortURL)

	}
	err = u.db.SaveTx(ctx, links)

	if err != nil {
		u.log.Debug("Save() database error", zap.Error(err))
		return nil, e.WrapError("database error", err)
	}

	return shorts, nil
}

// Get возвращает указатель на заполненный объект ссылки по сокращенному url
func (u *UseCases) Get(ctx context.Context, shortURL string) (*link.Link, error) {
	var (
		uuid, originalURL string
		deletedFlag       bool
		err               error
	)
	// получаем поля из БД
	uuid, originalURL, deletedFlag, err = u.db.Get(ctx, shortURL)
	if err != nil {
		u.log.Debug("Get() database error", zap.String("sURL", shortURL), zap.Error(err))
		return nil, e.WrapError("database error", err)
	}
	// если ссылка в БД не нашлась
	if originalURL == "" {
		u.log.Debug("short link does not exist in the database", zap.String("short", shortURL))
		return nil, nil
	}
	// возвращаем указатель на заполненный объект ссылки
	return &link.Link{
		OriginalURL: originalURL,
		ShortURL:    shortURL,
		UUID:        uuid,
		DeletedFlag: deletedFlag,
	}, nil
}

// GetShort возвращает указатель на объект ссылки по оригинальному url
func (u *UseCases) GetShort(ctx context.Context, originalURL string) (*link.Link, error) {
	var (
		uuid, shortURL string
		err            error
	)

	uuid, shortURL, err = u.db.GetShort(ctx, originalURL)

	if err != nil {
		u.log.Debug("GetShort() database error", zap.String("sURL", shortURL), zap.Error(err))
		return nil, e.WrapError("database error", err)
	}

	if shortURL == "" {
		u.log.Debug("original URL does not exist in the database", zap.String("original", originalURL))
		return nil, nil
	}

	return &link.Link{
		OriginalURL: originalURL,
		ShortURL:    shortURL,
		UUID:        uuid,
	}, nil
}

// GetUserUrls возвращает слайс указателей на объекты ссылки по идентификатору пользователя
func (u *UseCases) GetUserUrls(ctx context.Context, uuid string) ([]*link.Link, error) {
	// получаем [][]string с результатами запроса
	result, err := u.db.GetUserUrls(ctx, uuid)
	if err != nil {
		u.log.Debug("GetUserUrls() database error", zap.String("uuid", uuid), zap.Error(err))
		return nil, e.WrapError("database error", err)
	}
	// создаем слайс указателей на ссылки
	links := make([]*link.Link, 0)
	// заполняем результирующий слайс, uuid преднамеренно оставляем пустым
	for _, v := range result {
		links = append(links, &link.Link{UUID: "", ShortURL: v[0], OriginalURL: v[1]})
	}

	return links, nil
}

// DeleteUserUrls помечает в БД ссылки пользователя как удаленные
func (u *UseCases) DeleteUserUrls(ctx context.Context, shortUrls []string, uuid string) error {
	// создаем и заполняем слайс ссылок для удаления
	var links []*link.Link
	for _, v := range shortUrls {
		links = append(links, &link.Link{UUID: uuid, ShortURL: v})
	}
	// передаем слайс ссылок в хранилище для асинхронного удаления
	u.db.RemoveUrls(ctx, links)

	return nil
}
