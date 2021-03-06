package v1beta1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
	"github.com/stretchr/testify/assert"
)

func addClaims(ctx context.Context, claims []string) context.Context {
	return context.WithValue(ctx, core.ContextBoundaryKey, claims)
}

func TestGetDistributions(t *testing.T) {
	fake := v1beta1.NewFake()

	tests := []struct {
		claims []string
		want   int
	}{
		{
			claims: []string{"gr1"},
			want:   http.StatusOK,
		},
		{
			claims: []string{},
			want:   http.StatusInternalServerError,
		},
		{
			claims: []string{"gr2"},
			want:   http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", "/distributions", nil)
		req = req.WithContext(addClaims(req.Context(), test.claims))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(getDistributions(fake))
		handler.ServeHTTP(rr, req)

		t.Logf("response: %s", rr.Body.String())
		assert.Equal(t, test.want, rr.Code)
	}
}

func TestCreateInvalidation(t *testing.T) {
	fake := v1beta1.NewFake()

	tests := []struct {
		claims       []string
		name         string
		body         v1beta1.InvalidationRequest
		wantCode     int
		wantResponse v1beta1.InvalidationResponse
	}{
		{
			claims: []string{"gr1"},
			name:   "notfound",
			body: v1beta1.InvalidationRequest{
				Paths: []string{"/*"},
			},
			wantCode: 404,
			wantResponse: v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "Resource not found: notfound",
				},
			},
		},
		{
			claims: []string{"gr1"},
			name:   "unauthorized distribution",
			body: v1beta1.InvalidationRequest{
				Paths: []string{"/test/*"},
			},
			wantCode: 403,
			wantResponse: v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "User is not entitled to invalidate distribution: unauthorized distribution",
				},
			},
		},
		{
			claims: []string{"gr1"},
			name:   "dr1",
			body: v1beta1.InvalidationRequest{
				Paths: []string{"/test/*"},
			},
			wantCode: 201,
			wantResponse: v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "OK",
				},
			},
		},
		{
			claims:   []string{"gr1"},
			name:     "dr1",
			body:     v1beta1.InvalidationRequest{},
			wantCode: 400,
			wantResponse: v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "'paths' is a required field.",
				},
			},
		},
		{
			claims: []string{},
			name:   "dr1",
			body: v1beta1.InvalidationRequest{
				Paths: []string{"/test/*"},
			},
			wantCode:     500,
			wantResponse: v1beta1.InvalidationResponse{},
		},
	}

	for _, test := range tests {
		body, err := json.Marshal(test.body)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("/distributions/%s/invalidations", test.name), bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{
			"name": test.name,
		})
		req = req.WithContext(addClaims(req.Context(), test.claims))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(createInvalidation(fake))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, test.wantCode, rr.Code)

		if test.wantCode != 500 {
			resp := v1beta1.InvalidationResponse{}
			assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			assert.Equal(t, test.wantResponse, resp)
		}

	}
}

func TestGetInvalidation(t *testing.T) {
	fake := v1beta1.NewFake()

	tests := []struct {
		name      string
		invalidID string

		wantCode     int
		wantResponse v1beta1.InvalidationResponse
	}{
		{
			name:      "notfound",
			invalidID: "id",
			wantCode:  404,
			wantResponse: v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "Resource not found: notfound",
				},
			},
		},

		{
			name:      "unauthorized distribution",
			invalidID: "id",
			wantCode:  403,
			wantResponse: v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "User is not entitled to invalidate distribution: unauthorized distribution",
				},
			},
		},

		{
			name:      "d1",
			invalidID: "id",
			wantCode:  200,
			wantResponse: v1beta1.InvalidationResponse{
				Created: time.Unix(0, 0).UTC(),
				ID:      "1",
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "Complete",
				},
			},
		},

		{
			name:      "d1",
			invalidID: "notfound",
			wantCode:  404,
			wantResponse: v1beta1.InvalidationResponse{
				InvalidationMeta: v1beta1.InvalidationMeta{
					Status: "Resource not found: mocking AWS Error cloudfront.ErrCodeNoSuchInvalidation",
				},
			},
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", fmt.Sprintf("/distributions/%s/invalidations/%s", test.name, test.invalidID), nil)
		req = mux.SetURLVars(req, map[string]string{
			"name": test.name,
			"id":   test.invalidID,
		})
		req = req.WithContext(addClaims(req.Context(), []string{"test"}))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(getInvalidation(fake))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, test.wantCode, rr.Code)

		resp := v1beta1.InvalidationResponse{}
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, test.wantResponse, resp)

	}
}
