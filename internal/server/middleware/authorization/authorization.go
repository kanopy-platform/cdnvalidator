package authorization

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ClaimsKey string

const (
	ContextBoundaryKey ClaimsKey = "claims"
)

type Entitler interface {
	Entitled(req http.Request, claims []string) bool
}

type Authorizer interface {
	Authz(next http.Handler) http.Handler
}

type Middleware struct {
	entitlementManager Entitler
}

func New(opts ...Option) *Middleware {
	m := &Middleware{}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) addClaims(ctx context.Context, claims []string) context.Context {
	return context.WithValue(ctx, ContextBoundaryKey, claims)
}

func (m *Middleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// process entitlement logic
		log.Info("authz middleware")

		claims := []string{"g1", "g2"}
		// todo parse JWT

		// add information to context

		req = req.WithContext(m.addClaims(req.Context(), claims))

		next.ServeHTTP(w, req)
	})
}
