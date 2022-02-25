package authorization

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type EntitlementKey string
type BoundaryKey string

const (
	ContextEntitlementKey EntitlementKey = "entitlement"
	ContextBoundaryKey    BoundaryKey    = "boundaries"
)

type EntitlementGetter interface {
	GetEntitlements(distrbution string, boundaries ...string) interface{} // todo, grab correct type from config
}

type Authorizer interface {
	Authz(next http.Handler) http.Handler
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

// TODO swap entitlement interface{} for type
func (m *Middleware) addEntitlement(ctx context.Context, entitlement interface{}) context.Context {
	return context.WithValue(ctx, ContextEntitlementKey, entitlement)
}

func (m *Middleware) addBoundaries(ctx context.Context, boundaries []string) context.Context {
	return context.WithValue(ctx, ContextBoundaryKey, boundaries)
}

func (m *Middleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// process entitlement logic
		log.Info("authz middleware")

		boundaries := []string{"g1", "g2"}
		// todo parse JWT

		// add information to context
		entitlement := "hello world" // todo swap out for type
		req = req.WithContext(m.addEntitlement(req.Context(), entitlement))
		req = req.WithContext(m.addBoundaries(req.Context(), boundaries))

		next.ServeHTTP(w, req)
	})
}
