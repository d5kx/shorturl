package storage

import (
	"math/rand"
	"strings"
	"time"

	"github.com/d5kx/shorturl/internal/util/err"
)

const (
	shorURLLength = 8
	maxURLLength  = 256
)

var symbolDic = []byte{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
	'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
	'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
}

type Storage interface {
	Save(*Link) error
	Get(shortURL string) (*Link, error)
	IsExist(shortURL string) (bool, error)
	Remove(shortURL string) error
}

type Link struct {
	URL string
}

func (l Link) ShortURL() (string, error) {
	var b strings.Builder
	var e error

	rand.Seed(time.Now().UnixNano())
	ln := len(symbolDic)

	for i := 0; i < shorURLLength; i++ {

		e = b.WriteByte(symbolDic[rand.Intn(ln)])
		if e != nil {
			return "", err.WrapError("can't generate short link", e)
		}
	}

	return b.String(), nil
}
