package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	router *http.ServeMux
}

func New() http.Handler {
	s := &Server{router: http.NewServeMux()}

	s.router.HandleFunc("/", s.handleRoot())
	s.router.HandleFunc("/healthz", s.handleHealthz())
	s.router.HandleFunc("/test", s.handleTest())

	return s.router
}

func (s *Server) handleRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world")
	}
}

func (s *Server) handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status": "ok",
		}

		bytes, err := json.Marshal(status)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(bytes))
	}
}

// Temporary handler for testing
// Remove this once real API endpoints are implemented
func (s *Server) handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := cloudfront.NewClient()
		if err != nil {
			log.Errorf("cloudfront.NewClient failed: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		log.Info("Creating Invalidation...")

		create, err := c.CreateInvalidation("E12AN388E0D4UU", []string{"/"})
		if err != nil {
			log.Errorf("CreateInvalidation failed: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Infof("CreateInvalidation: Id=%v, Status=%v", create.InvalidationId, create.Status)

		get, err := c.GetInvalidation("E12AN388E0D4UU", create.InvalidationId)
		if err != nil {
			log.Errorf("GetInvalidation failed: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Infof("GetInvalidation: CreateTime=%v, Status=%v, Paths=%v", get.CreateTime, get.Status, get.Paths)

		fmt.Fprint(w, "success")
	}
}
