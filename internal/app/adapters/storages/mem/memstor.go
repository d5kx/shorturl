package memstor

import (
	"bufio"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"reflect"
	"sync"

	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/entities"
	"github.com/d5kx/shorturl/internal/util/e"
)

type Storage struct {
	linksDB  map[string]*entities.Link //Хранилище ссылок
	usersDB  map[string]*entities.User //Хранилище пользователей
	log      loggers.Logger
	isActive bool
}

func (s *Storage) GetDB() map[string]*entities.Link {
	return s.linksDB
}

func New(logger loggers.Logger) *Storage {
	return &Storage{
		linksDB: make(map[string]*entities.Link),
		usersDB: make(map[string]*entities.User),
		log:     logger,
	}
}

func (s *Storage) Save(ctx context.Context, l *entities.Link) error {
	for _, v := range s.linksDB {
		if v.OriginalURL == l.OriginalURL {
			return e.ErrSaveUniqueViolation
		}
	}

	s.linksDB[l.ShortURL] = l /*l.OriginalURL*/
	return nil
}

func (s *Storage) SaveTx(ctx context.Context, slice []*entities.Link) error {
	for _, v := range slice {
		s.Save(ctx, v)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, shortURL string) (string, string, bool, error) {
	value, ok := s.linksDB[shortURL]

	if !ok {
		return "", "", false, nil
	}
	return value.UUID, value.OriginalURL, value.DeletedFlag, nil
}

func (s *Storage) GetShort(ctx context.Context, originalURL string) (string, string, error) {
	for _, v := range s.linksDB {
		if v.OriginalURL == originalURL {
			return v.UUID, v.ShortURL, nil
		}
	}
	return "", "", nil
}

func (s *Storage) GetUserUrls(ctx context.Context, uuid string) ([][]string, error) {
	return make([][]string, 0), nil
}

func (s *Storage) LinkExist(ctx context.Context, shortURL string) (bool, error) {
	_, ok := s.linksDB[shortURL]
	return ok, nil
}

func (s *Storage) Remove(ctx context.Context, shortURL string) error {
	delete(s.linksDB, shortURL)
	return nil
}

func (s *Storage) RemoveUrls(ctx context.Context, links []*entities.Link) {}

func (s *Storage) IsActive() bool {
	return s.isActive
}

func (s *Storage) UserExist(ctx context.Context, login string) (bool, error) {
	_, ok := s.usersDB[login]
	return ok, nil
}

func (s *Storage) UserSave(ctx context.Context, user *entities.User) error {
	s.usersDB[user.Login] = user
	return nil
}
func (s *Storage) UserGet(ctx context.Context, login string) (string, string, error) {
	return "", "", nil
}
func (s *Storage) Shutdown(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)

	// записываем файл со ссылками
	go func() {
		defer wg.Done()
		err := writeMapToFile[entities.Link](conf.GetDBFileName(), s.linksDB, s.log)
		if err != nil {
			return
		}
	}()

	// записываем файл с пользователями
	go func() {
		defer wg.Done()
		err := writeMapToFile[entities.User](conf.GetUsersFileName(), s.usersDB, s.log)
		if err != nil {
			return
		}
	}()
	//ждем все записывающие горутины
	wg.Wait()
	return nil
}
func (s *Storage) Open(name string) error {
	s.isActive = true
	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Bootstrap(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)

	//загружаем файл со ссылками
	go func() {
		defer wg.Done()
		err := readFileToMap[entities.Link](conf.GetDBFileName(), s.linksDB, 1, s.log)
		if err != nil {
			return
		}
	}()
	//загружаем файл с пользователями
	go func() {
		defer wg.Done()
		err := readFileToMap[entities.User](conf.GetUsersFileName(), s.usersDB, 1, s.log)
		if err != nil {
			return
		}
	}()
	//ждем читающие горутины
	wg.Wait()
	return nil
}

// writeMapToFile дженерик, записывает содержимое мапы в файл
func writeMapToFile[T any](fileName string, data map[string]*T, log loggers.Logger) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Debug("can't open file", zap.String("file", fileName), zap.Error(err))
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Info("file closing error when save file", zap.String("file", fileName), zap.Error(err))
		}
	}()
	// создаем декодер и врайтер
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)

	// пишем в файл слайс ссылок
	var c int //Счетчик сохраненных записей
	for i, _ := range data {
		if err = encoder.Encode(data[i]); err != nil {
			log.Debug("can't encode json when saving to file", zap.String("file", fileName), zap.Error(err))
			return err
		}
		c++
	}
	//сбрасываем буфер в файл
	if err = writer.Flush(); err != nil {
		log.Info("error in Flush() when saving to file ", zap.String("file", fileName), zap.Error(err))
		return err
	}
	// логируем результаты записи в файл
	log.Info("file was written without errors", zap.String("file", conf.GetDBFileName()), zap.Int("records", c))

	return nil
}

// readFileToMap дженерик, читает содержимое файла в мапу, использует рефлексию,
// keyField - индекс поля структуры для ключа мапы
func readFileToMap[T any](fileName string, dataMap map[string]*T, keyField int, log loggers.Logger) error {

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		log.Info("can't open file", zap.String("file", fileName), zap.Error(err))
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Info("file closing error when load from file", zap.String("file", fileName), zap.Error(err))
		}
	}()
	var i int // Счетчик прочитанных записей
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()
		var l T
		err = json.Unmarshal(data, &l)
		if err != nil {
			log.Debug("can't decode json when reading from file", zap.String("file", fileName), zap.Error(err))
			return err
		}
		v := reflect.ValueOf(l).Field(keyField)
		dataMap[v.String()] = &l
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Info("file scanning error when loaf from file", zap.String("file", fileName), zap.Error(err))
		return err
	}
	// логируем результаты загрузки из файла
	log.Info("file was loaded without errors", zap.String("file", fileName), zap.Int("records", i))

	return nil
}
