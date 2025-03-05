package mockgen

import (
	"strconv"
)

type Gen struct {
}

func New() *Gen {
	return &Gen{}
}

func (g *Gen) ShortURL() string {
	return "AbCdEf"
}

func (g *Gen) UID() string {
	return strconv.Itoa(rand.Intn(23456789))

}
