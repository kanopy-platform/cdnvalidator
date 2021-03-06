package authorization

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/internal/jwt"
	"github.com/stretchr/testify/assert"
)

type Mock struct {
	Claims []string
}

// e.g. http.HandleFunc("/health-check", HealthCheckHandler)
func (m *Mock) MockContextHandler(w http.ResponseWriter, r *http.Request) {
	// inspect context
	m.Claims = r.Context().Value(core.ContextBoundaryKey).([]string)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"retval": "done"}`)
}

func TestAuthorizationContext(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.

	rawToken, err := jwt.NewTestJWTWithClaims(jwt.Claims{
		Groups: []string{"g1"},
		Scopes: []string{"g2"},
	})

	assert.NoError(t, err)

	req, err := http.NewRequest("GET", "/some-auth-path", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rawToken))
	assert.NoError(t, err)

	m := &Mock{}

	middleware := New(WithAuthorizationHeader())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(m.MockContextHandler)

	// wrap the test handler in the authz middleware
	middleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, []string{"g1", "g2"}, m.Claims)
}

func TestAuthorizationResponses(t *testing.T) {
	rawToken, err := jwt.NewTestJWTWithClaims(jwt.Claims{
		Groups: []string{"yes"},
		Scopes: []string{"g2"},
	})
	assert.NoError(t, err)

	tests := []struct {
		middleware func(http.Handler) http.Handler
		token      string
		want       int
	}{
		{
			token:      "",
			middleware: New(),
			want:       401,
		},

		{
			token:      "invalidtoken",
			middleware: New(),
			want:       403,
		},

		{
			token:      rawToken,
			middleware: New(WithCookieName("testme")),
			want:       200,
		},

		{
			token:      rawToken,
			middleware: New(WithAuthorizationHeader(), WithCookieName("testme")),
			want:       200,
		},

		{
			token:      rawToken,
			middleware: New(WithAuthorizationHeader()),
			want:       200,
		},
	}

	for _, test := range tests {
		m := &Mock{}

		req, err := http.NewRequest("GET", "/some-auth-path", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.token))

		req.AddCookie(&http.Cookie{
			Name:  "testme",
			Value: test.token,
		})

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(m.MockContextHandler)

		// wrap the test handler in the authz middleware
		test.middleware(handler).ServeHTTP(rr, req)
		assert.Equal(t, test.want, rr.Result().StatusCode)
	}
}
