package basegen

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strings"
	"time"
)

type Gen struct {
}

const (
	shorURLLength = 9 // Длина короткой ссылки
)

// Набор символов для генерации коротких ссылок
var symbolsDictionary = []byte{
	'A', 'b', 'C', 'd', 'E', 'f', 'G', 'h', 'I', 'j',
	'a', 'B', 'c', 'D', 'e', 'F', 'g', 'H', 'i', 'J',
	'K', 'l', 'M', 'n', 'O', 'p', 'Q', 'r', 'S', 't',
	'k', 'L', 'm', 'N', 'o', 'P', 'q', 'R', 's', 'T',
	'u', 'V', 'w', 'X', 'y', 'Z', 'U', 'v', 'W', 'x',
	'Y', 'z',
}

func New() *Gen {
	return &Gen{}
}

// ShortURL генерирует короткую ссылку
func (g *Gen) ShortURL() string {
	var b strings.Builder

	rand.NewSource(time.Now().UnixNano())
	ln := len(symbolsDictionary)

	for i := 0; i < shorURLLength; i++ {
		b.WriteByte(symbolsDictionary[rand.Intn(ln)])
	}

	return b.String()
}

// UUID генерирует идентификатор пользователя
func (g *Gen) UUID() string {
	return uuid.New().String()
}

// HashPassword вычисляет хеш пароля
func (g *Gen) HashPassword(passwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(passwd), 16)
	return string(bytes), err
}

// ComparePasswordHash проверяет, является ли хеш образованным от пароля
func (g *Gen) ComparePasswordHash(passwd, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwd))
	return err == nil
}
