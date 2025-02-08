package processor

import (
	"net/http"
)

type Processor interface {
	//Process(res http.ResponseWriter, req *http.Request)
	Get(res http.ResponseWriter, req *http.Request)
	Post(res http.ResponseWriter, req *http.Request)
	BadRequest(res http.ResponseWriter, req *http.Request)
}
