package authorization

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/kanopy-platform/cdnvalidator/internal/jwt"
	log "github.com/sirupsen/logrus"
)

type ClaimsKey string

const (
	ContextBoundaryKey ClaimsKey = "claims"
)

type Entitler interface {
	Entitled(req *http.Request, claims []string) bool
}

type Authorizer interface {
	Authz(next http.Handler) http.Handler
}

type Middleware struct {
	entitlementManager Entitler
	authCookieName     string
	authHeaderEnabled  bool
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

func (m *Middleware) getAuthorizationToken(req *http.Request) (string, error) {
	if m.authHeaderEnabled {
		if _, ok := req.Header["Authorization"]; ok {
			v := req.Header.Get("Authorization")
			if strings.HasPrefix(v, "Bearer") {
				v = strings.TrimPrefix(v, "Bearer ")
			}
			return v, nil
		}
	}

	// check cookie
	v, err := req.Cookie(m.authCookieName)
	if err != nil {
		return "", err
	}

	if v.Value == "" {
		return "", fmt.Errorf("token empty")
	}

	return v.Value, nil
}

func (m *Middleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// process entitlement logic

		tokenString, err := m.getAuthorizationToken(req)
		if err != nil {
			log.WithError(err).Error("no authorization token found")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenClaims, err := jwt.TokenClaims(tokenString)
		if err != nil {
			log.WithError(err).Error("unable to parse token claims")
			http.Error(w, "invalid token", http.StatusForbidden)
			return
		}

		claims := append(tokenClaims.Groups, tokenClaims.Scopes...)

		if m.entitlementManager != nil {
			if !m.entitlementManager.Entitled(req, claims) {
				log.WithError(err).Error("unauthorized action")
				http.Error(w, "invalid permissions to perform the requested action", http.StatusForbidden)
				return
			}
		}

		// add information to context
		req = req.WithContext(m.addClaims(req.Context(), claims))
		next.ServeHTTP(w, req)
	})
}
