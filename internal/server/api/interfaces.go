package api

import (
	"net/http"
)

type Routable interface {
	Handler() http.Handler
	PathPrefix() string
}
