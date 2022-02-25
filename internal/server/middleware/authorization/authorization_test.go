package authorization

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Mock struct {
	Claims []string
}

// e.g. http.HandleFunc("/health-check", HealthCheckHandler)
func (m *Mock) MockContextHandler(w http.ResponseWriter, r *http.Request) {
	// inspect context
	m.Claims = r.Context().Value(ContextBoundaryKey).([]string)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"retval": "done"}`)
}

func TestAuthorizationContext(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/some-auth-path", nil)
	assert.NoError(t, err)

	m := &Mock{}

	a := New()

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(m.MockContextHandler)

	// wrap the test handler in the authz middleware
	a.Authz(handler).ServeHTTP(rr, req)

	assert.Equal(t, []string{"g1", "g2"}, m.Claims)
}
