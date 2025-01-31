package link

import (
	"math/rand"
	"strings"
	"time"
)

const (
	shorURLLength = 6
)

var symbolDictionary = []byte{
	'A', 'b', 'C', 'd', 'E', 'f', 'G', 'h', 'I', 'j',
	'a', 'B', 'c', 'D', 'e', 'F', 'g', 'H', 'i', 'J',
	'K', 'l', 'M', 'n', 'O', 'p', 'Q', 'r', 'S', 't',
	'k', 'L', 'm', 'N', 'o', 'P', 'q', 'R', 's', 'T',
	'u', 'V', 'w', 'X', 'y', 'Z', 'U', 'v', 'W', 'x',
	'Y', 'z',
}

type Link struct {
	URL string
}

func (l Link) ShortURL() string {
	var b strings.Builder

	rand.NewSource(time.Now().UnixNano())
	ln := len(symbolDictionary)

	for i := 0; i < shorURLLength; i++ {
		b.WriteByte(symbolDictionary[rand.Intn(ln)])
	}

	return b.String()
}
