package baseauth

import (
	"context"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/util/e"
	"github.com/d5kx/shorturl/internal/util/generators"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const TOKEN_EXP = time.Hour * 240
const SECRET_KEY = "supersecretkey"

type Auth struct {
	log loggers.Logger
	gen generators.Generator
}

type claims struct {
	jwt.RegisteredClaims
	UserID string
}

func New(generator generators.Generator, logger loggers.Logger) *Auth {
	return &Auth{
		log: logger,
		gen: generator,
	}
}

func (a *Auth) Do(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		cookie, err := req.Cookie("user_id")
		//куку из запроса не получили, генерируем новый user_id и выставляем куку с ним
		if err != nil {
			//генерируем подписанный токен
			userId, err := a.buildJWTString(a.gen.UUID())
			if err != nil {
				a.log.Debug("can't build auth token", zap.Error(err))
				http.Error(res, "can't build auth token", http.StatusInternalServerError)
				return
			}
			//устанавливаем куку с подписанным токеном
			a.setAuthCookie(res, userId)
			ctx = context.WithValue(ctx, "user_id", userId)

		} else { // куку из запроса получили, user_id передаем дальше по контексту
			userId, err := a.getUserID(cookie.Value)
			/*if err == e.ErrAuthTokenNotValid || err == e.ErrUnexpSigningMethod {
				a.log.Debug("invalid auth token", zap.Error(err))
				http.Error(res, "invalid auth token", http.StatusUnauthorized)
				return
			}*/
			if err != nil {
				a.log.Debug("invalid auth token", zap.Error(err))
				http.Error(res, "invalid auth token", http.StatusUnauthorized)
				return
			}
			ctx = context.WithValue(ctx, "user_id", userId)
			a.log.Debug("read cookie", zap.Any("cookie", userId))
		}
		//будем передавать user_id по цепочке middleware через контекст
		next.ServeHTTP(res, req.WithContext(ctx))
	}
}

func (a *Auth) setAuthCookie(res http.ResponseWriter, uuid string) {
	cookie := &http.Cookie{
		Name:     "user_id",
		Value:    uuid,
		Path:     "/",
		HttpOnly: true,                    // Доступ только через HTTP, защита от XSS
		Secure:   true,                    // Только HTTPS
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}

	http.SetCookie(res, cookie)
	a.log.Debug("set cookie", zap.Any("cookie", cookie))
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (a *Auth) buildJWTString(uuid string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: uuid,
	})
	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Теперь попробуем получить из строки токена полезную нагрузку, а именно — UserID
func (a *Auth) getUserID(tokenString string) (string, error) {
	// создаём экземпляр структуры с утверждениями
	claimsObj := &claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claimsObj,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, e.ErrUnexpSigningMethod
			}
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", e.ErrAuthTokenNotValid
	}
	return claimsObj.UserID, nil
}
