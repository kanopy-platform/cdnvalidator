package prometheus

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type middleware struct {
	buckets           []float64
	httpTotalRequests *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
}

func New(opts ...Option) func(http.Handler) http.Handler {
	m := &middleware{
		buckets: []float64{0.5, 1, 2, 3, 5},
	}

	for _, opt := range opts {
		opt(m)
	}

	m.httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_duration_seconds_bucket",
		Help:    "Duration of HTTP requests.",
		Buckets: m.buckets,
	}, []string{"path", "code", "method"})

	m.httpTotalRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of HTTP requests",
	}, []string{"path", "code", "method"})

	return m.handler
}

func (m *middleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		labels := []string{strings.ToLower(path), strconv.Itoa(metrics.Code), strings.ToLower(r.Method)}

		go func() {
			m.httpTotalRequests.WithLabelValues(labels...).Inc()
			m.httpDuration.WithLabelValues(labels...).Observe(metrics.Duration.Seconds())
		}()
	})
}
