package processor

import (
	"net/http"
)

type Processor interface {
	Process(res http.ResponseWriter, req *http.Request)
}
