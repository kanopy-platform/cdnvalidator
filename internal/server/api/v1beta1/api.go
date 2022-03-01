package v1beta1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kanopy-platform/cdnvalidator/internal/core"
	log "github.com/sirupsen/logrus"
)

const ErrUserNotEntitled = "User is not entitled to the CloudFront Invalidation service"

const PathPrefix = "/api/v1beta1"

func New(router *mux.Router, ds core.DistributionService) *mux.Router {
	api := router.PathPrefix(PathPrefix).Subrouter()

	// append api handlers here
	api.HandleFunc("/distributions", getDistributions(ds)).Methods(http.MethodGet)
	api.HandleFunc("/distributions/{name}/invalidations", createInvalidation(ds)).Methods(http.MethodPost)
	api.HandleFunc("/distributions/{name}/invalidations/{id}", getInvalidation(ds)).Methods(http.MethodGet)

	return api
}

// GET Distributions /api/v1beta/distributions
func getDistributions(ds core.DistributionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := ds.List(r.Context())
		if err != nil {
			log.WithError(err).Error("unexpected error listing distributions")
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}

		if len(d) == 0 {
			http.Error(w, ErrUserNotEntitled, http.StatusUnauthorized)
			return
		}

		err = json.NewEncoder(w).Encode(map[string]interface{}{"distributions": d})
		if err != nil {
			log.WithError(err).Error("unexpected encoding error")
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}
	}
}

// POST Distributions /api/v1beta/distributions/{name}
func createInvalidation(ds core.DistributionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// GET Invalidation /api/v1beta/distributions/{name}/invalidations/{id}
func getInvalidation(ds core.DistributionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
