package generators

type Generator interface {
	ShortURL() string
	UID() string
}
