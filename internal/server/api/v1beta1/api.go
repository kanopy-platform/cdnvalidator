package v1beta1

import (
	"github.com/gorilla/mux"
)

const PathPrefix = "/api/v1beta1"

func New(router *mux.Router) *mux.Router {

	// append api handlers here

	return router
}

// define api handler functions here
