package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
)

type Server struct {
	router         *http.ServeMux
	authCookieName string
}

func New(opts ...Option) (http.Handler, error) {
	s := &Server{router: http.NewServeMux()}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	s.router.HandleFunc("/", s.handleRoot())
	s.router.HandleFunc("/healthz", s.handleHealthz())

	/* TODO enable authz middleware on routes.
	a := authorization.New(authorization.WithCookieName(s.authCookieName),
		authorization.WithAuthorizationHeader()) // TODO add with entitlements option
	v1beta1.New().RegisterRoutes(a, s.router)
	*/

	return logRequestHandler(s.router), nil
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
