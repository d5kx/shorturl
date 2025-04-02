package baseauth

import (
	"context"
	"errors"
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
		var (
			needGenerateToken bool
			userId            string
			err               error
			cookie            *http.Cookie
		)
		//захватываем контекст запроса
		ctx := req.Context()
		// пытаемся получить куку из запроса
		cookie, err = req.Cookie("user_id")

		if err != nil { //куку из запроса не получили
			needGenerateToken = true
		} else { // куку из запроса получили
			// получаем user_id из куки
			userId, err = a.getUserID(cookie.Value)
			// если токен перестал быть валидным
			if errors.Is(err, e.ErrAuthTokenNotValid) {
				needGenerateToken = true
			}
			// если ошибка парсинга строки jwt токена
			if errors.Is(err, e.ErrAuthTokenParse) {
				a.log.Debug("invalid auth token", zap.Error(err))
				http.Error(res, "invalid auth token", http.StatusUnauthorized)
				return
			}
			a.log.Debug("read cookie", zap.Any("userId", cookie))
		}
		// если нужно генерируем новый user_id и выставляем куку с ним
		if needGenerateToken {
			// генерируем подписанный токен
			userId = a.gen.UUID()
			var userIdJWT string
			userIdJWT, err = a.buildJWTString(userId)
			if err != nil {
				a.log.Debug("can't build auth token", zap.Error(err))
				http.Error(res, "can't build auth token", http.StatusInternalServerError)
				return
			}
			//устанавливаем куку с подписанным токеном
			a.setAuthCookie(res, userIdJWT)
		}
		//будем передавать user_id по цепочке middleware через контекст
		a.log.Debug("send to middleware", zap.String("user_ud", userId))
		ctx = context.WithValue(ctx, "user_id", userId)
		next.ServeHTTP(res, req.WithContext(ctx))
	}
}

// setAuthCookie создает куку и устанавливает её в заголовок ответа
func (a *Auth) setAuthCookie(res http.ResponseWriter, uuidJWT string) {
	cookie := &http.Cookie{
		Name:     "user_id",
		Value:    uuidJWT,
		Path:     "/",
		HttpOnly: true,                    // Доступ только через HTTP, защита от XSS
		Secure:   true,                    // Только HTTPS
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}

	http.SetCookie(res, cookie)
	a.log.Debug("set cookie", zap.Any("cookie", cookie))
}

// BuildJWTString создаёт токен и возвращает его в виде строки
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
	a.log.Debug("set user_id", zap.String("user_id", uuid))
	return tokenString, nil
}

// getUserID попробует получить из строки токена полезную нагрузку, а именно — UserID
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
	// возвращаем свою ошибку парсинга токена
	if err != nil {
		return "", e.ErrAuthTokenParse
	}
	// возвращаем свою ошибку валидации токена
	if !token.Valid {
		return "", e.ErrAuthTokenNotValid
	}
	a.log.Debug("read user_id", zap.String("user_id", claimsObj.UserID))
	return claimsObj.UserID, nil
}
