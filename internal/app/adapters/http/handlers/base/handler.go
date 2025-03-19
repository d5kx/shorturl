package basehandler

import (
	"bytes"
	"encoding/json"
	usedb "github.com/d5kx/shorturl/internal/app/usecases/db"
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

func (h *Handler) Get(res http.ResponseWriter, req *http.Request) {
	short := strings.TrimPrefix(req.URL.Path, "/")
	l, err := h.linkUse.Get(req.Context(), short)
	if err != nil || l == nil {
		h.log.Debug("can't process GET request",
			zap.String("short", short),
			zap.Error(err),
		)
		res.WriteHeader(http.StatusBadRequest)

		return
	}

	res.Header().Set("Location", l.OriginalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) Post(res http.ResponseWriter, req *http.Request) {
	if !h.checkContentType(req, "text/plain") && !h.checkContentType(req, "application/x-gzip") {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(req.Body)
	defer req.Body.Close()
	if buf.Len() == 0 {
		h.log.Debug("can't process POST request (no link in body request)", zap.String("body", buf.String()))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	sURL, err := h.linkUse.Save(req.Context(), buf.String())
	if err != nil {
		h.log.Debug("can't process POST request (short link is not saved)", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	buf.Reset()
	buf.WriteString(conf.GetResURLAdr() + "/")
	buf.WriteString(sURL)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(buf.Bytes())
	if err != nil {
		h.log.Debug("can't process POST request (can't write response body)", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}
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
		h.log.Debug("can't decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	sURL, err := h.linkUse.Save(req.Context(), request.URL)
	if err != nil {
		h.log.Debug("can't process POST request (short link is not saved in the database)", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	// заполняем модель ответа
	var response = models.ResponseJSON{
		Result: conf.GetResURLAdr() + "/" + sURL,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	// сериализуем ответ сервера
	jsonByte, err := json.Marshal(response)
	if err != nil {
		h.log.Debug("can't process POST request (can't encode response)", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = res.Write(jsonByte)
	if err != nil {
		h.log.Debug("can't process POST request (can't write response JSON body)", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	/*
		enc := json.NewEncoder(res)
		if err := enc.Encode(response); err != nil {
			p.loggers.Debug("can't encode response", zap.Error(err))
			res.WriteHeader(http.StatusBadRequest)
			return
		}*/
}

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
