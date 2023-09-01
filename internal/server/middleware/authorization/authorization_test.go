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
		name       string
		token      string
		skipBearer bool
		want       int
	}{
		{
			token:      "",
			name:       "no providers",
			middleware: New(),
			want:       401,
		},

		{
			token:      "invalidtoken",
			middleware: New(),
			name:       "invalid token, no cookie",
			want:       401,
		},

		{
			token:      rawToken,
			middleware: New(WithCookieName("testme")),
			name:       "valid named cookie",
			want:       200,
		},
		{
			token:      "",
			middleware: New(WithCookieName("testme")),
			name:       "empty named cookie",
			want:       401,
		},
		{
			token:      rawToken,
			middleware: New(WithAuthorizationHeader(), WithCookieName("testme")),
			name:       "valid AuthorizationHeade, valid cookie",
			want:       200,
		},

		{
			token:      rawToken,
			middleware: New(WithAuthorizationHeader()),
			name:       "valid Authorization BearerHeader",
			want:       200,
		},
		{
			token:      rawToken,
			middleware: New(WithHeaderName("testy")),
			name:       "valid Named Header",
			want:       200,
		},
		{
			token:      "invalid",
			middleware: New(WithHeaderName("testy")),
			name:       "invalid Named Header",
			want:       403,
		},
		{
			token:      rawToken,
			middleware: New(WithHeaderName("empty")),
			name:       "empty Named Header",
			want:       401,
		},
		{
			token:      rawToken,
			middleware: New(WithCookieName("testme"), WithHeaderName("testy")),
			name:       "valid Named Header, valid cookie",
			want:       200,
		},
		{
			token:      rawToken,
			middleware: New(WithAuthorizationHeader(), WithCookieName("testme"), WithHeaderName("testy")),
			name:       "invalid Authorization Bearer header, valid Named Header, valid cookie",
			skipBearer: true,
			want:       200,
		},
		{
			token:      rawToken,
			middleware: New(WithAuthorizationHeader(), WithCookieName("testme"), WithHeaderName("testy")),
			name:       "valid Authorization Bearer header, valid Named Header, valid cookie",
			want:       200,
		},
		{
			token:      rawToken,
			middleware: New(WithAuthorizationHeader(), WithCookieName("invalid"), WithHeaderName("testy")),
			name:       "valid Authorization Bearer header, valid Named Header, invalid cookie",
			want:       200,
		},
		{
			token:      rawToken,
			middleware: New(WithCookieName("testme"), WithHeaderName("invalid")),
			name:       "empty Named Header, invalid cookie",
			want:       401,
		},
		{
			token:      "invalid",
			middleware: New(WithCookieName("testme"), WithHeaderName("testy")),
			name:       "invalid Named Header, valid cookie",
			want:       403,
		},
	}

	for _, test := range tests {
		m := &Mock{}

		req, err := http.NewRequest("GET", "/some-auth-path", nil)
		assert.NoError(t, err)

		if !test.skipBearer {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.token))
		}
		req.Header.Set("testy", test.token)

		req.AddCookie(&http.Cookie{
			Name:  "testme",
			Value: test.token,
		})

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(m.MockContextHandler)

		// wrap the test handler in the authz middleware
		test.middleware(handler).ServeHTTP(rr, req)
		assert.Equal(t, test.want, rr.Result().StatusCode, test.name)
	}
}
