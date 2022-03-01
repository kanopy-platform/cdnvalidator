package v1beta1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestGetDistributions(t *testing.T) {

	fake := core.NewFake()

	req, err := http.NewRequest("GET", "/distributions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getDistributions(fake))
	handler.ServeHTTP(rr, req)

	t.Logf("response: %s", rr.Body.String())
	assert.Equal(t, http.StatusOK, rr.Code)

}
