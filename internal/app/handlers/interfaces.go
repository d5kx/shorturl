package handlers

import (
	"net/http"
)

type Handler interface {
	Get(res http.ResponseWriter, req *http.Request)
	Post(res http.ResponseWriter, req *http.Request)
	PostAPIShorten(res http.ResponseWriter, req *http.Request)
	BadRequest(res http.ResponseWriter, req *http.Request)
}
