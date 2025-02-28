package memstor

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Storage struct {
	db map[string]string
	//db map[string]link.Link
}

func (s *Storage) GetDB() map[string]string {
	return s.db
}

func New() *Storage {
	return &Storage{db: make(map[string]string)}
}

func (s *Storage) Save(l *link.Link) error {
	s.db[l.ShortURL] = l.OriginalURL
	if conf.GetDBFileName() == "" {
		return nil
	}
	return s.SaveToFile(l)
}

func (s *Storage) Get(shortURL string) (string, error) {
	value, ok := s.db[shortURL]

	if !ok {
		return "", nil
	}
	return value, nil
}

func (s *Storage) IsExist(shortURL string) (bool, error) {
	_, ok := s.db[shortURL]
	return ok, nil
}

func (s *Storage) Remove(shortURL string) error {
	delete(s.db, shortURL)
	return nil
}

func (s *Storage) SaveToFile(l *link.Link) error {
	file, err := os.OpenFile(conf.GetDBFileName(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return e.WrapError("can't open file "+conf.GetDBFileName(), err)
	}
	writer := bufio.NewWriter(file)
	defer func() {
		if err := writer.Flush(); err != nil {
			//добавить запись в лог
		}
		if err := file.Close(); err != nil {
			//добавить запись в лог
		}
	}()

	rand.NewSource(time.Now().UnixNano())
	l.UID = strconv.Itoa(rand.Intn(999999))

	if err = json.NewEncoder(writer).Encode(l); err != nil {
		return e.WrapError("can't encode json when saving to file", err)
	}

	return nil
}

func (s *Storage) LoadFromFile() error {
	file, err := os.OpenFile(conf.GetDBFileName(), os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return e.WrapError("can't open file "+conf.GetDBFileName(), err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			//добавить запись в лог
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()
		l := link.Link{}
		err = json.Unmarshal(data, &l)
		if err != nil {
			return e.WrapError("can't decode json when reading from file", err)
		}

		s.db[l.ShortURL] = l.OriginalURL
	}

	if err := scanner.Err(); err != nil {
		//добавить запись в лог
	}

	return nil
}
