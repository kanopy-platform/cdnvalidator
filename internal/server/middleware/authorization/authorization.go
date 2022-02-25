package authorization

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type EntitlementGetter interface {
	GetEntitlements(e ...string) interface{}
}

type Middleware struct {
	entitlement EntitlementGetter
}

func New(opts ...Option) *Middleware {
	m := &Middleware{}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// process entitlement logic
		log.Info("authz middleware")
		next.ServeHTTP(w, req)
	})
}
