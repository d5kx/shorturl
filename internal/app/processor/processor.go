package processor

import (
	"net/http"
)

type Processor interface {
	Get(res http.ResponseWriter, req *http.Request)
	Post(res http.ResponseWriter, req *http.Request)
	PostAPIShorten(res http.ResponseWriter, req *http.Request)
	BadRequest(res http.ResponseWriter, req *http.Request)
}
