package v1beta1

import (
	"github.com/gorilla/mux"
)

const PathPrefix = "/api/v1beta1"

func New(router *mux.Router) *mux.Router {
	api := router.PathPrefix(PathPrefix).Subrouter()

	// append api handlers here

	return api
}

// define api handler functions here
