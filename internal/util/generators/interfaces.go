package generators

type Generator interface {
	ShortURL() string
	UUID() string
	HashPassword(passwd string) (string, error)
	ComparePasswordHash(passwd, hash string) bool
}
