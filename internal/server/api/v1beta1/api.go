package v1beta1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
	log "github.com/sirupsen/logrus"
)

const ErrUserNotEntitled = "User is not entitled to the CloudFront Invalidation service"

const PathPrefix = "/api/v1beta1"

func New(router *mux.Router, ds *v1beta1.DistributionService) *mux.Router {
	api := router.PathPrefix(PathPrefix).Subrouter()

	// append api handlers here
	api.HandleFunc("/distributions", getDistributions(ds)).Methods(http.MethodGet)
	api.HandleFunc("/distributions/{name}/invalidations", createInvalidation(ds)).Methods(http.MethodPost)
	api.HandleFunc("/distributions/{name}/invalidations/{id}", getInvalidation(ds)).Methods(http.MethodGet)

	return api
}

// swagger:route GET /api/v1beta1/distributions GetDistributions get-distributions
//
// Get a list of distributions you are entitled to perform invalidations.
//
//     Security:
//       jwt:
//
// responses:
//   200: DistributionResponse
//   401: ErrorResponse
//   500: ErrorResponse
func getDistributions(ds DistributionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := ds.List(r.Context())
		if err != nil {
			logError(w, err, "unexpected error listing distributions", http.StatusInternalServerError)
			return
		}

		if len(d) == 0 {
			http.Error(w, ErrUserNotEntitled, http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]interface{}{"distributions": d}, http.StatusOK)
	}
}

// swagger:route POST  /api/v1beta1/distributions/{name}/invalidations SubmitInvalidation submit-invalidation
//
// Submit an Invalidation Request
//
//     Security:
//       jwt:
//
// responses:
//   200: InvalidationResponse
//   400: InvalidationError
//   403: ErrorResponse
//   404: ErrorResponse
//   500: ErrorResponse
func createInvalidation(ds DistributionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		name := vars["name"]

		invalidationReq := v1beta1.InvalidationRequest{}
		if err := json.NewDecoder(r.Body).Decode(&invalidationReq); err != nil {
			logError(w, err, "unexpected encoding error", http.StatusInternalServerError)
			return
		}

		if len(invalidationReq.Paths) == 0 {
			writeJSON(w, &v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "'paths' is a required field.",
				},
			}, http.StatusBadRequest)
			return
		}

		status, err := ds.CreateInvalidation(r.Context(), name, invalidationReq.Paths)
		if err != nil {
			if v1beta1.ErrorBadRequest(err) {
				writeJSON(w, err, http.StatusBadRequest)
				return
			}

			if v1beta1.ErrorResourceNotFound(err) {
				writeJSON(w, err, http.StatusNotFound)
				return
			}

			if v1beta1.ErrorIsUnauthorized(err) {
				writeJSON(w, err, http.StatusForbidden)
				return
			}

			logError(w, err, "unexpected error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, status, http.StatusCreated)
	}
}

// swagger:route GET  /api/v1beta1/distributions/{name}/invalidations/{id} InvalidationResponse get-invalidation-status
//
// Get an Invalidation Request
//
//     Security:
//       jwt:
//
// responses:
//   200: InvalidationResponse
//   400: InvalidationError
//   403: ErrorResponse
//   404: ErrorResponse
//   500: ErrorResponse
func getInvalidation(ds DistributionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		name := vars["name"]
		invalidationID := vars["id"]

		result, err := ds.GetInvalidationStatus(r.Context(), name, invalidationID)
		if err != nil {
			if v1beta1.ErrorResourceNotFound(err) {
				writeJSON(w, err, http.StatusNotFound)
				return
			}

			if v1beta1.ErrorIsUnauthorized(err) {
				writeJSON(w, err, http.StatusForbidden)
				return
			}
			logError(w, err, "unexpected error", http.StatusInternalServerError)

			return
		}

		writeJSON(w, result, http.StatusOK)
	}
}

func logError(w http.ResponseWriter, err error, msg string, statusCode int) {
	log.WithError(err).Error(msg)
	http.Error(w, "unexpected error", statusCode)
}

func writeJSON(w http.ResponseWriter, out interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		logError(w, err, "unexpected encoding error", http.StatusInternalServerError)
	}
}
