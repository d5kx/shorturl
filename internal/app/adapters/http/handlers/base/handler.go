package basehandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/d5kx/shorturl/internal/app/usecases/db"
	"github.com/d5kx/shorturl/internal/util/e"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
	"strings"

	"github.com/d5kx/shorturl/internal/app/adapters/http/models"
	"github.com/d5kx/shorturl/internal/app/adapters/loggers"
	"github.com/d5kx/shorturl/internal/app/conf"
	"github.com/d5kx/shorturl/internal/app/usecases/link"

	"go.uber.org/zap"
)

type Handler struct {
	linkUse *uselink.UseCases
	dbUse   *usedb.UseCases
	log     loggers.Logger
}

func New(useCase *uselink.UseCases, dbUse *usedb.UseCases, logger loggers.Logger) *Handler {
	return &Handler{
		linkUse: useCase,
		log:     logger,
		dbUse:   dbUse,
	}
}

func (h *Handler) GetHTTPS(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "Proudly served with Go and HTTPS!")

}

// Get хендлер для получения оригинального адреса ссылки
func (h *Handler) Get(res http.ResponseWriter, req *http.Request) {
	// получаем короткую ссылку из адреса запроса
	short := strings.TrimPrefix(req.URL.Path, "/")
	// получаем заполненный объект ссылки
	l, err := h.linkUse.Get(req.Context(), short)
	if err != nil || l == nil {
		h.log.Debug("can't process GET request",
			zap.String("short", short),
			zap.Error(err),
		)
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// если ссылка помечена в БД как удаленная
	if l.DeletedFlag {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusGone)
		return
	}

	// пишем ответ, переадресация по оригинальному адресу
	res.Header().Set("Location", l.OriginalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// GetUserUrls хендлер для получения списка всех ссылок пользователя
func (h *Handler) GetUserUrls(res http.ResponseWriter, req *http.Request) {
	// получаем идентификатор пользователя из контекста
	v := req.Context().Value("user_id")
	// получаем массив указателей на ссылки пользователя
	links, err := h.linkUse.GetUserUrls(req.Context(), v.(string))
	if err != nil {
		h.log.Debug("can't process GET request",
			zap.String("user_id", v.(string)),
			zap.Error(err),
		)
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// если сохраненных ссылок у пользователя нет
	if len(links) == 0 {
		http.Error(res, "links not found", http.StatusNoContent)
		return
	}
	// добавляем адрес сервера
	pref := "http://" + conf.GetServAdr() + "/"
	for k, _ := range links {
		links[k].ShortURL = pref + links[k].ShortURL
	}
	// сериализуем в JSON массив ссылок, uuid получаем пустым, в сериализацию не попадает
	jsonByte, err := json.Marshal(links)
	if err != nil {
		h.logBadRequest(res, "can't process GET request (can't encode response)", err)
		return
	}
	// пишем заголовки и тело ответа
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(jsonByte)
	if err != nil {
		h.logBadRequest(res, "can't process GET request (can't write response JSON body)", err)
		return
	}
}

// Post хендлер для добавления одной ссылки и получения одной короткой ссылки
func (h *Handler) Post(res http.ResponseWriter, req *http.Request) {
	// проверяем содержание требуемого типа в строке Content-type запроса
	if !h.checkContentType(req, "text/plain") && !h.checkContentType(req, "application/x-gzip") {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// читаем оригинальную ссылку из тела запроса
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	// если данных в теле запроса нет
	if buf.Len() == 0 {
		h.logBadRequest(res, "can't process POST request (body is empty)", nil)
		return
	}
	// получаем оригинальную ссылку в виде строки
	originalUrl := buf.String()
	// получаем user_id из контекста запроса
	userId := req.Context().Value("user_id")
	//пытаемся сохранить оригинальную ссылку с идентификатором пользователя
	sURL, err := h.linkUse.Save(req.Context(), originalUrl, userId.(string))
	if err != nil {
		var pgErr *pgconn.PgError
		// если ошибки нарушения уникальности, оригинальная ссылка уже существует
		if errors.Is(err, e.ErrSaveUniqueViolation) || (errors.As(err, &pgErr) &&
			pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)) {
			// получаем существующую короткую ссылку
			l, err := h.linkUse.GetShort(req.Context(), originalUrl)
			if err != nil || l == nil {
				h.logBadRequest(res, "can't process GetShort request: original ="+originalUrl, err)
				return
			}
			// пишем ответ с существующей короткой ссылкой
			h.writePostResponse(res, l.ShortURL, http.StatusConflict)
			return
		}
		// если неконкретизированные ошибки
		h.logBadRequest(res, "can't process POST request (short link is not saved)", err)
		return
	}
	// ошибок нет, пишем ответ с новой короткой ссылкой
	h.writePostResponse(res, sURL, http.StatusCreated)
}

// PostAPIShorten хендлер для добавления одной ссылки и получения одной короткой ссылки в json формате
func (h *Handler) PostAPIShorten(res http.ResponseWriter, req *http.Request) {
	// проверяем содержание требуемого типа в строке Content-type запроса
	if !h.checkContentType(req, "application/json") {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// десериализируем запрос в структуру модели
	var request models.RequestJSON
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&request); err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}
	// получаем user_id из контекста запроса
	userId := req.Context().Value("user_id")
	//пытаемся сохранить оригинальную ссылку с идентификатором пользователя
	sURL, err := h.linkUse.Save(req.Context(), request.URL, userId.(string))
	if err != nil {
		var pgErr *pgconn.PgError
		// если ошибки нарушения уникальности, оригинальная ссылка уже существует
		if errors.Is(err, e.ErrSaveUniqueViolation) || (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)) {
			// получаем существующую короткую ссылку
			l, err := h.linkUse.GetShort(req.Context(), request.URL)
			if err != nil || l == nil {
				h.logBadRequest(res, "can't process GetShort request: original ="+request.URL, err)
				return
			}
			// пишем ответ с существующей короткой ссылкой
			h.writePostJsonResponse(res, l.ShortURL, http.StatusConflict)
			return
		}
		// если неконкретизированные ошибки
		h.logBadRequest(res, "can't process POST request (short link is not saved in the database)", err)
		return
	}
	// ошибок нет, пишем ответ с новой короткой ссылкой
	h.writePostJsonResponse(res, sURL, http.StatusCreated)
}

// PostAPIShortenBatch хендлер для добавления пачки ссылок и получения пачки коротких ссылок в json формате
func (h *Handler) PostAPIShortenBatch(res http.ResponseWriter, req *http.Request) {
	// проверяем содержание требуемого типа в строке Content-type запроса
	if !h.checkContentType(req, "application/json") {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var (
		responseSlice []models.ResponseJSONBatch // слайс моделей для ответа
		originalURLs  []string                   // слайс оригинальных ссылок
	)

	dec := json.NewDecoder(req.Body)
	// читаем "[" в json массиве
	_, err := dec.Token()
	if err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}
	// читаем поэлементно json массив
	for dec.More() {
		var request models.RequestJSONBatch
		if err := dec.Decode(&request); err != nil {
			h.logBadRequest(res, "can't decode request JSON body", err)
			return
		}
		// сохраняем оригинальные ссылки
		originalURLs = append(originalURLs, request.OriginalURL)
		// заполняем слайс моделей для ответа, пока без коротких ссылок
		responseSlice = append(responseSlice, models.ResponseJSONBatch{
			CorrelationId: request.CorrelationId,
			ShortURL:      "",
		})
	}
	// читаем "]" в json массиве
	_, err = dec.Token()
	if err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}
	// запрос и ответ пустые
	if len(responseSlice) == 0 {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// получаем user_id из контекста запроса
	userId := req.Context().Value("user_id")
	// пытаемся сохранить транзакцией в базу данных оригинальные ссылки
	// в ответ получаем слайс коротких ссылок
	sURLs, err := h.linkUse.SaveTx(req.Context(), originalURLs, userId.(string))
	if err != nil {
		h.logBadRequest(res, "can't process POST request (short link is not saved in the database)", err)
		return
	}
	// заполняем короткие ссылки в слайсе моделей ответа
	for k, _ := range responseSlice {
		responseSlice[k].ShortURL = sURLs[k]
	}
	// сериализируем в json слайс моделей ответа
	jsonByte, err := json.Marshal(responseSlice)
	if err != nil {
		h.logBadRequest(res, "can't process POST request (can't encode response)", err)
		return
	}
	// пишем заголовки и тело ответа
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(jsonByte)
	if err != nil {
		h.logBadRequest(res, "can't process POST request (can't write response JSON body)", err)
		return
	}
}

// DeleteUserUrls помечает ссылки в БД как удаленные для данного пользователя
// Ссылки приходят в теле запроса в виде json массива
func (h *Handler) DeleteUserUrls(res http.ResponseWriter, req *http.Request) {
	// проверяем содержание требуемого типа в строке Content-type запроса
	if !h.checkContentType(req, "application/json") {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// декодер json
	dec := json.NewDecoder(req.Body)
	// читаем "[" в json массиве
	_, err := dec.Token()
	if err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}
	//var buf bytes.Buffer
	var shortUrls []string
	// читаем и декодируем поэлементно json массив
	for dec.More() {
		var request string
		if err := dec.Decode(&request); err != nil {
			h.logBadRequest(res, "can't decode request JSON body", err)
			return
		}
		shortUrls = append(shortUrls, request)
		//buf.WriteString(request)
	}
	// читаем "]" в json массиве
	_, err = dec.Token()
	if err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}

	//получаем user_id из контекста запроса
	userId := req.Context().Value("user_id")
	// отправляем массив коротких ссылок на "удаление"
	err = h.linkUse.DeleteUserUrls(req.Context(), shortUrls, userId.(string))
	if err != nil {
		h.logBadRequest(res, "can't process DELETE request", err)
		return
	}
	// пишем заголовки и тело ответа
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusAccepted)
	//res.Write(buf.Bytes())
}

// PingDB хендлер для проверки соединения с базой данных
func (h *Handler) PingDB(res http.ResponseWriter, req *http.Request) {
	// пинганули успешно, шлем 200 ОК, выходим
	if h.dbUse.Ping(req.Context()) {
		res.WriteHeader(http.StatusOK)
		return
	}
	// не пинганули, отправляем ошибку
	http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (h *Handler) BadRequest(res http.ResponseWriter, req *http.Request) {
	http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func (h *Handler) checkContentType(req *http.Request, t string) bool {
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, t) {
		h.log.Debug("can't process POST request (wrong Content-Type)",
			zap.String("actual", contentType),
			zap.String("expected", t),
		)
		return false
	}
	return true
}

func (h *Handler) logBadRequest(res http.ResponseWriter, mes string, err error) {
	h.log.Debug(mes, zap.Error(err))
	http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func (h *Handler) writePostResponse(res http.ResponseWriter, data string, successStatus int) {
	var buf bytes.Buffer

	buf.WriteString(conf.GetResURLAdr() + "/")
	buf.WriteString(data)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(successStatus)
	_, err := res.Write(buf.Bytes())
	if err != nil {
		h.logBadRequest(res, "can't process POST request (can't write response body)", err)
	}
}

func (h *Handler) writePostJsonResponse(res http.ResponseWriter, data string, successStatus int) {
	var response = models.ResponseJSON{
		Result: conf.GetResURLAdr() + "/" + data,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(successStatus)
	// сериализуем ответ сервера
	jsonByte, err := json.Marshal(response)
	if err != nil {
		h.logBadRequest(res, "can't process POST request (json marshal error)", err)
		return
	}
	_, err = res.Write(jsonByte)
	if err != nil {
		h.logBadRequest(res, "can't process POST request (can't write response JSON body)", err)
		return
	}
}
