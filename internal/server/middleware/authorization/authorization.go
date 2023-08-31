package authorization

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/internal/jwt"
	log "github.com/sirupsen/logrus"
)

type middleware struct {
	authCookieName    string
	authHeaderName    string
	authHeaderEnabled bool
}

func New(opts ...Option) func(http.Handler) http.Handler {
	m := &middleware{}

	for _, opt := range opts {
		opt(m)
	}

	return m.handler
}

func (m *middleware) addClaims(ctx context.Context, claims []string) context.Context {
	return context.WithValue(ctx, core.ContextBoundaryKey, claims)
}

func (m *middleware) getAuthorizationToken(req *http.Request) (string, error) {
	switch {
	case m.authHeaderEnabled:
		if _, ok := req.Header["Authorization"]; ok {
			v := req.Header.Get("Authorization")
			if strings.HasPrefix(v, "Bearer") {
				v = strings.TrimPrefix(v, "Bearer ")
			}
			return v, nil
		}
		fallthrough
	case m.authHeaderName != "":
		protoHeaderName := textproto.CanonicalMIMEHeaderKey(m.authHeaderName)
		if _, ok := req.Header[protoHeaderName]; ok {
			return req.Header.Get(m.authHeaderName), nil
		}

		return "", fmt.Errorf("No %s header found", m.authHeaderName)

	default:
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
}

func (m *middleware) handler(next http.Handler) http.Handler {
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

		// add information to context
		req = req.WithContext(m.addClaims(req.Context(), claims))
		next.ServeHTTP(w, req)
	})
}
