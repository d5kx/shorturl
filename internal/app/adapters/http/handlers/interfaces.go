package handlers

import (
	"net/http"
)

type Handler interface {
	GetHTTPS(res http.ResponseWriter, req *http.Request)
	Get(res http.ResponseWriter, req *http.Request)
	GetUserUrls(res http.ResponseWriter, req *http.Request)
	Post(res http.ResponseWriter, req *http.Request)
	PostAPIShorten(res http.ResponseWriter, req *http.Request)
	PostAPIShortenBatch(res http.ResponseWriter, req *http.Request)
	PostAPIUserRegister(http.HandlerFunc) http.HandlerFunc
	PostAPIUserLogin(http.HandlerFunc) http.HandlerFunc
	BadRequest(res http.ResponseWriter, req *http.Request)
	PingDB(res http.ResponseWriter, req *http.Request)
	DeleteUserUrls(res http.ResponseWriter, req *http.Request)
}
