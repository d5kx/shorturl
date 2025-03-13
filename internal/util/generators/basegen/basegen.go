package basegen

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Gen struct {
}

const (
	shorURLLength = 6
)

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

func (g *Gen) ShortURL() string {
	var b strings.Builder

	rand.NewSource(time.Now().UnixNano())
	ln := len(symbolsDictionary)

	for i := 0; i < shorURLLength; i++ {
		b.WriteByte(symbolsDictionary[rand.Intn(ln)])
	}

	return b.String()
}

func (g *Gen) UID() string {
	rand.NewSource(time.Now().UnixNano())
	return strconv.Itoa(rand.Intn(99999999))

}
