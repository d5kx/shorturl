package auth

import "net/http"

type Authorizer interface {
	Do(http.HandlerFunc) http.HandlerFunc
}
