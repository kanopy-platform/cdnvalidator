package server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/kanopy-platform/cdnvalidator/internal/config"
	v1beta1_ds "github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
	"github.com/kanopy-platform/cdnvalidator/internal/server/api/v1beta1"
	"github.com/kanopy-platform/cdnvalidator/internal/server/middleware/authorization"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
	"github.com/kanopy-platform/cdnvalidator/pkg/http/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

//go:embed templates
var embeddedFS embed.FS

type Server struct {
	router              *mux.Router
	template            *template.Template
	distributionService *v1beta1_ds.DistributionService
	authCookieName      string
}

func New(config *config.Config, cloudfront *cloudfront.Client, opts ...Option) (http.Handler, error) {
	s := &Server{
		router:   mux.NewRouter(),
		template: template.Must(template.ParseFS(embeddedFS, "templates/*.html")),
	}

	if config == nil {
		return nil, errors.New("missing required parameter config")
	}
	if cloudfront == nil {
		return nil, errors.New("missing required parameter cloudfront")
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	s.router.Use(prometheus.New())
	s.router.Use(logRequestHandler)
	s.router.HandleFunc("/healthz", s.handleHealthz())
	s.router.Handle("/metrics", promhttp.Handler())
	s.router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger", http.FileServer(http.Dir("swagger"))))

	ui := s.router.PathPrefix("").Subrouter()
	ui.HandleFunc("/", s.handleRoot()).Methods(http.MethodGet)

	authmiddleware := authorization.New(authorization.WithCookieName(s.authCookieName),
		authorization.WithAuthorizationHeader())

	s.distributionService = v1beta1_ds.New(config, cloudfront)
	api := v1beta1.New(s.router, s.distributionService)

	ui.Use(authmiddleware)
	api.Use(authmiddleware)

	return s.router, nil
}

func (s *Server) handleRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := s.distributionService.List(r.Context())
		if err != nil {
			log.WithError(err).Error("error get distribution list")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Distributions": d,
		}

		if err := s.template.ExecuteTemplate(w, "index.html", data); err != nil {
			log.WithError(err).Error("error executing template")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status": "ok",
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(status)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func logRequestHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		// Execute the chain of handlers, while capturing HTTP metrics: code, bytes-written, duration
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		host := r.Header.Get("x-forwarded-for")
		if host == "" {
			// r.RemoteAddr contains port, which we want to remove
			idx := strings.LastIndex(r.RemoteAddr, ":")
			if idx == -1 {
				host = r.RemoteAddr
			} else {
				host = r.RemoteAddr[:idx]
			}
		}

		// Combined log format
		// Using fmt.Fprintf here because logrus prints timestamps and log level by default
		fmt.Fprintf(os.Stderr, "%v %v %v [%v] %q %v %v %q %q %vms\n",
			host,                                   // host
			"-",                                    // user-identity
			"-",                                    // authuser
			t.Format("02/Jan/2006 15:04:05 +0000"), // date
			fmt.Sprintf("%v %v %v", r.Method, r.URL.Path, r.Proto), // request
			metrics.Code,                    // status
			metrics.Written,                 // bytes written
			r.Header.Get("referer"),         // referer
			r.Header.Get("user-agent"),      // user-agent
			metrics.Duration.Milliseconds(), // duration of HTTP handler
		)
	}
	return http.HandlerFunc(fn)
}
