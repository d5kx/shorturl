package storage

import (
	"math/rand"
	"strings"
	"time"

	"github.com/d5kx/shorturl/internal/util/e"
)

const (
	shorURLLength = 6
)

var symbolDic = []byte{
	'A', 'b', 'C', 'd', 'E', 'f', 'G', 'h', 'I', 'j',
	'a', 'B', 'c', 'D', 'e', 'F', 'g', 'H', 'i', 'J',
	'K', 'l', 'M', 'n', 'O', 'p', 'Q', 'r', 'S', 't',
	'k', 'L', 'm', 'N', 'o', 'P', 'q', 'R', 's', 'T',
	'u', 'V', 'w', 'X', 'y', 'Z', 'U', 'v', 'W', 'x',
	'Y', 'z',
}

type Storage interface {
	Save(*Link) (string, error)
	Get(shortURL string) (*Link, error)
	IsExist(shortURL string) (bool, error)
	Remove(shortURL string) error
}

type Link struct {
	URL string
}

func (l Link) ShortURL() (string, error) {
	var b strings.Builder
	var err error

	rand.NewSource(time.Now().UnixNano())
	ln := len(symbolDic)

	for i := 0; i < shorURLLength; i++ {

		err = b.WriteByte(symbolDic[rand.Intn(ln)])
		if err != nil {
			return "", e.WrapError("can't generate short link", err)
		}
	}

	return b.String(), nil
}
