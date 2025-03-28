package baseauth

import (
	"context"
	"fmt"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/util/generators"
	"go.uber.org/zap"
	"net/http"
)

type Auth struct {
	log loggers.Logger
	gen generators.Generator
}

func New(generator generators.Generator, logger loggers.Logger) *Auth {
	return &Auth{
		log: logger,
		gen: generator,
	}
}

func (a *Auth) Do(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//будем передавать user_id по цепочке middleware через контекст
		ctx := req.Context()
		cookie, err := req.Cookie("user_id")
		if err != nil {
			userId := a.gen.UUID()
			a.setAuthCookie(res, userId)
			ctx = context.WithValue(ctx, "user_id", userId)

		} else {
			fmt.Println(cookie.Value)
			ctx = context.WithValue(ctx, "user_id", cookie.Value)
			a.log.Debug("read cookie", zap.Any("cookie", cookie))
		}

		next.ServeHTTP(res, req.WithContext(ctx))
	}
}

func (a *Auth) setAuthCookie(res http.ResponseWriter, uuid string) {
	cookie := &http.Cookie{
		Name:     "user_id",
		Value:    a.gen.UUID(),
		Path:     "/",
		HttpOnly: true,                    // Доступ только через HTTP, защита от XSS
		Secure:   true,                    // Только HTTPS
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}

	http.SetCookie(res, cookie)
	a.log.Debug("set cookie", zap.Any("cookie", cookie))
}
