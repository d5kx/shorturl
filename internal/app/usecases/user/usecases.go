package useuser

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
	db  storages.UserStorage
	log loggers.Logger
	gen generators.Generator
}

func New(
	storage storages.UserStorage,
	generator generators.Generator,
	logger loggers.Logger,
) *UseCases {
	return &UseCases{
		db:  storage,
		log: logger,
		gen: generator,
	}
}

// UserExist проверяет, существует ли пользователь с таким логином
func (u *UseCases) UserExist(ctx context.Context, login string) (bool, error) {
	//делаем запрос к БД
	return u.db.UserExist(ctx, login)
}

// UserSave сохраняет пользователя в БД.
// Генерирует идентификатор пользователя и хеш пароля
func (u *UseCases) UserSave(ctx context.Context, login string, passwd string) (*entities.User, error) {
	// вычисляем хеш пароля
	passwdHash, err := u.gen.HashPassword(passwd)
	if err != nil {
		return nil, e.WrapError("passwd hash generation error", err)
	}

	// создаем объект user
	user := &entities.User{
		UUID:       u.gen.UUID(),
		Login:      login,
		PasswdHash: passwdHash,
	}

	//делаем запрос к БД на добавление пользователя
	err = u.db.UserSave(ctx, user)
	if err != nil {
		u.log.Debug("UserSave() database error", zap.Error(err))
		return nil, e.WrapError("database error", err)
	}

	return user, nil
}

// UserAuth достает пользователя из БД и проверяет валидность пароля
func (u *UseCases) UserAuth(ctx context.Context, login string, passwd string) (*entities.User, bool, error) {
	//делаем запрос к БД на извлечение данных пользователя
	uuid, passwdHash, err := u.db.UserGet(ctx, login)
	if err != nil {
		u.log.Debug("UserGet() database error", zap.Error(err))
		return nil, false, e.WrapError("database error", err)
	}
	// создаем объект user
	user := &entities.User{
		UUID:       uuid,
		Login:      login,
		PasswdHash: passwdHash,
	}
	return user, u.gen.ComparePasswordHash(passwd, user.PasswdHash), nil
}
