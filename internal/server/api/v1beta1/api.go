package v1beta1

import (
	"net/http"
)

const PathPrefix = "/api/v1beta1"

func New() http.Handler {
	router := http.NewServeMux()

	// append api handlers here

	return router
}

// define api handler functions here
