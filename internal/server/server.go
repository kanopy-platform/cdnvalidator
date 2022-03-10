package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	v1beta1_ds "github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
	"github.com/kanopy-platform/cdnvalidator/internal/server/api/v1beta1"
	"github.com/kanopy-platform/cdnvalidator/internal/server/middleware/authorization"
	"github.com/kanopy-platform/cdnvalidator/pkg/http/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	router         *mux.Router
	authCookieName string
	apiOptions     []v1beta1_ds.Option
}

func New(opts ...Option) (http.Handler, error) {
	s := &Server{router: mux.NewRouter()}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	s.router.Use(prometheus.New())
	s.router.Use(logRequestHandler)
	s.router.HandleFunc("/", s.handleRoot())
	s.router.HandleFunc("/healthz", s.handleHealthz())
	s.router.Handle("/metrics", promhttp.Handler())
	s.router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger", http.FileServer(http.Dir("swagger"))))

	authmiddleware := authorization.New(authorization.WithCookieName(s.authCookieName),
		authorization.WithAuthorizationHeader())

	api, err := v1beta1.New(s.router,
		s.apiOptions...,
	)
	if err != nil {
		return nil, err
	}

	api.Use(authmiddleware)

	return s.router, nil
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
