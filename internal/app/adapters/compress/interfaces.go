package compress

import (
	"net/http"
)

type Compressor interface {
	Do(http.HandlerFunc) http.HandlerFunc
}
