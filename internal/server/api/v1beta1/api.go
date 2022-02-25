package v1beta1

import (
	"net/http"

	"github.com/kanopy-platform/cdnvalidator/internal/server/api"
)

type APIV1Beta1 struct {
	router *http.ServeMux
	prefix string
}

func New() api.Routable {
	a := &APIV1Beta1{
		router: http.NewServeMux(),
		prefix: "/api/v1beta1",
	}

	// append api handlers

	return a
}

func (a *APIV1Beta1) Handler() http.Handler {
	return a.router
}

func (a *APIV1Beta1) PathPrefix() string {
	return a.prefix
}

// define api handler functions here
