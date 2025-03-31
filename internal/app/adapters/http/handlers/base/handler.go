package basehandler

import (
	"bytes"
	"encoding/json"
	"errors"
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

// Get хендлер для получения оригинального адреса ссылки
func (h *Handler) Get(res http.ResponseWriter, req *http.Request) {
	// получаем короткую ссылку из адреса запроса
	short := strings.TrimPrefix(req.URL.Path, "/")
	// получаем оригинальный адрес
	l, err := h.linkUse.Get(req.Context(), short)
	if err != nil || l == nil {
		h.log.Debug("can't process GET request",
			zap.String("short", short),
			zap.Error(err),
		)
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	// пишем ответ, переадресация с новым адресом
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
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	// если сохраненных ссылок у пользователя нет
	if len(links) == 0 {
		http.Error(res, "links not found", http.StatusNoContent)
		return
	}
	// сериализуем в JSON массив ссылок, uuid получаем пустым, в сериализацию не попадает
	jsonByte, err := json.Marshal(links)
	if err != nil {
		h.logBadRequest(res, "can't process GET request (can't encode response)", err)
		return
	}
	// пишем заголовки ответа
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	// пишем тело ответа
	_, err = res.Write(jsonByte)
	if err != nil {
		h.logBadRequest(res, "can't process GET request (can't write response JSON body)", err)
		return
	}
}

func (h *Handler) Post(res http.ResponseWriter, req *http.Request) {
	if !h.checkContentType(req, "text/plain") && !h.checkContentType(req, "application/x-gzip") {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	data := buf.String()
	if buf.Len() == 0 {
		h.logBadRequest(res, "can't process POST request (body is empty)", nil)
		return
	}
	// получаем user_id из контекста запроса
	v := req.Context().Value("user_id")
	sURL, err := h.linkUse.Save(req.Context(), data, v.(string))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &e.ErrSaveUniqueViolation) || (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)) {
			l, err := h.linkUse.GetShort(req.Context(), data)
			if err != nil || l == nil {
				h.logBadRequest(res, "can't process GetShort request: original ="+data, err)
				return
			}
			h.writePostResponse(res, l.ShortURL, http.StatusConflict)
			return
		}
		h.logBadRequest(res, "can't process POST request (short link is not saved)", err)
		return
	}
	h.writePostResponse(res, sURL, http.StatusCreated)
}

func (h *Handler) PostAPIShorten(res http.ResponseWriter, req *http.Request) {
	if !h.checkContentType(req, "application/json") {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// десериализуем запрос в структуру модели
	var request models.RequestJSON
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&request); err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}
	// получаем user_id из контекста запроса
	v := req.Context().Value("user_id")
	sURL, err := h.linkUse.Save(req.Context(), request.URL, v.(string))
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &e.ErrSaveUniqueViolation) || (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)) {
			l, err := h.linkUse.GetShort(req.Context(), request.URL)
			if err != nil || l == nil {
				h.logBadRequest(res, "can't process GetShort request: original ="+request.URL, err)
				return
			}
			h.writePostJsonResponse(res, l.ShortURL, http.StatusConflict)
			return
		}
		h.logBadRequest(res, "can't process POST request (short link is not saved in the database)", err)
		return
	}
	h.writePostJsonResponse(res, sURL, http.StatusCreated)
}

func (h *Handler) PostAPIShortenBatch(res http.ResponseWriter, req *http.Request) {
	if !h.checkContentType(req, "application/json") {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var slice []models.ResponseJSONBatch
	var originalURLs []string

	dec := json.NewDecoder(req.Body)
	_, err := dec.Token()
	if err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}

	for dec.More() {
		var request models.RequestJSONBatch
		if err := dec.Decode(&request); err != nil {
			h.logBadRequest(res, "can't decode request JSON body", err)
			return
		}
		originalURLs = append(originalURLs, request.OriginalURL)

		slice = append(slice, models.ResponseJSONBatch{
			CorrelationId: request.CorrelationId,
			ShortURL:      "",
		})
	}

	_, err = dec.Token()
	if err != nil {
		h.logBadRequest(res, "can't decode request JSON body", err)
		return
	}

	if len(slice) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	sURLs, err := h.linkUse.SaveTx(req.Context(), originalURLs)
	if err != nil {
		h.logBadRequest(res, "can't process POST request (short link is not saved in the database)", err)
		return
	}

	for k, _ := range slice {
		slice[k].ShortURL = sURLs[k]
	}

	jsonByte, err := json.Marshal(slice)
	if err != nil {
		h.logBadRequest(res, "can't process POST request (can't encode response)", err)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	_, err = res.Write(jsonByte)
	if err != nil {
		h.logBadRequest(res, "can't process POST request (can't write response JSON body)", err)
		return
	}
}

// PingDB хендлер для проверки соединения с базой данных
func (h *Handler) PingDB(res http.ResponseWriter, req *http.Request) {
	if h.dbUse.Ping(req.Context()) {
		res.WriteHeader(http.StatusOK)
		return
	}
	res.WriteHeader(http.StatusInternalServerError)
}

func (h *Handler) BadRequest(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusBadRequest)
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
	http.Error(res, mes, http.StatusBadRequest)
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
