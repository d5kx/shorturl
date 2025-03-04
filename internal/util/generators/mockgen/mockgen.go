package mockgen

type Gen struct {
}

func New() *Gen {
	return &Gen{}
}

func (g *Gen) ShortURL() string {
	return "AbCdEf"
}
