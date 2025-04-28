package auth

import "net/http"

type Authorizer interface {
	Do(next http.HandlerFunc) http.HandlerFunc
	SendUserAuthCookie(next http.HandlerFunc) http.HandlerFunc
}
