package baseauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/util/e"
	"github.com/d5kx/shorturl/internal/util/generators"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
)

const TOKEN_EXP = time.Hour * 24000
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

// GenerateTLSCertificate генерирует самоподписанный TLS сертификат
func (a *Auth) GenerateTLSCertificate() error {
	// генерируем приватный ключ на основе эллиптических кривых
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		a.log.Debug("failed to generate private key", zap.Error(err))
		return err
	}
	//генерируем серийный номер
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		a.log.Debug("failed to generate serial number", zap.Error(err))
		return err
	}
	//создаем шаблон сертификата
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"My Corp"},
		},
		DNSNames:  []string{"localhost"}, //сертификат действителен для домена localhost
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(6 * time.Hour), // срок действия сертификата 6 часов

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	// Сертификат создается из шаблона и подписывается закрытым ключом.
	// &template передается как для шаблона, так и для родительского параметра CreateCertificate.
	// Последнее делает этот сертификат самоподписанным
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		a.log.Debug("failed to create certificate", zap.Error(err))
		return err
	}

	// сохраняем сертификат в файл
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		a.log.Debug("failed to encode certificate to PEM")
	}

	if err := os.WriteFile(conf.GetTSLCertFileName(), pemCert, 0644); err != nil {
		a.log.Debug("failed to write cert.pem", zap.Error(err))
		return err
	}
	a.log.Debug("wrote " + conf.GetTSLCertFileName())

	// сохраняем приватный ключ в файл
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		a.log.Debug("Failed to encode key to PEM")
	}
	if err := os.WriteFile(conf.GetTSLKeyFileName(), pemKey, 0600); err != nil {
		a.log.Debug("failed to write key.pem", zap.Error(err))
		return err
	}
	a.log.Debug("wrote " + conf.GetTSLKeyFileName())
	return nil
}
