package v1beta1

import (
	"net/http"

	"github.com/gorilla/mux"
)

const PathPrefix = "/api/v1beta1"

func New(router *mux.Router, config interface{}) *mux.Router {
	api := router.PathPrefix(PathPrefix).Subrouter()

	// append api handlers here
	api.HandleFunc("/distributions", getDistributions(config)).Methods(http.MethodGet)
	api.HandleFunc("/distributions/{name}/invalidations", createInvalidation(config)).Methods(http.MethodPost)
	api.HandleFunc("/distributions/{name}/invalidations/{id}", getInvalidation(config)).Methods(http.MethodGet)

	return api
}

// GET Distributions /api/v1beta/distributions
func getDistributions(config interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// GET Distributions /api/v1beta/distributions/{name}
func createInvalidation(config interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// GET Invalidation /api/v1beta/distributions/{name}/invalidations/{id}
func getInvalidation(config interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
