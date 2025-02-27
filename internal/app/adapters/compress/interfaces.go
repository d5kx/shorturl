package compress

import (
	"net/http"
)

type Compressor interface {
	RequestCompress(http.HandlerFunc) http.HandlerFunc
}
