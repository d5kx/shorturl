package generators

type Generator interface {
	ShortURL() string
	UUID() string
}
