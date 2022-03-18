package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kanopy-platform/cdnvalidator/internal/config"
	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
	"github.com/stretchr/testify/assert"
)

func addClaims(ctx context.Context, claims []string) context.Context {
	return context.WithValue(ctx, core.ContextBoundaryKey, claims)
}

func newTestConfig() (*config.Config, error) {
	configYaml := `---
distributions:
  dis1:
    id: "123"
    prefix: "/foo"
  dis2:
    id: "456"
    prefix: "/bar"
entitlements:
  grp1:
    - dis1
    - dis2
  grp2:
    - dis2
`
	return config.NewTestConfigWithYaml([]byte(configYaml))
}

func TestGetDistribution(t *testing.T) {
	testConfig, err := newTestConfig()
	assert.NoError(t, err)

	testCf := cloudfront.NewTestCloudfrontClient(&cloudfront.MockCloudFrontClient{})

	ds := New(testConfig, testCf)

	tests := []struct {
		// inputs
		claims           []string
		distributionName string
		// outputs
		want *config.Distribution
		err  error
	}{
		{
			// success
			claims:           []string{"grp1"},
			distributionName: "dis1",
			want:             &config.Distribution{ID: "123", Prefix: "/foo"},
			err:              nil,
		},
		{
			// non existant distribution
			claims:           []string{"grp1"},
			distributionName: "dis3",
			want:             nil,
			err:              NewInvalidationError(ResourceNotFoundErrorCode, errors.New("distribution dis3 not found"), "dis3"),
		},
		{
			// user claim is not entitled to distribution
			claims:           []string{"grp2"},
			distributionName: "dis1",
			want:             nil,
			err:              NewInvalidationError(InvalidationUnauthorizedErrorCode, errors.New("distribution unauthorized"), "dis1"),
		},
	}

	for _, test := range tests {
		ctx := addClaims(context.Background(), test.claims)

		ret, err := ds.getDistribution(ctx, test.distributionName)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.want, ret)
		}
	}
}

func TestList(t *testing.T) {
	testConfig, err := newTestConfig()
	assert.NoError(t, err)

	testCf := cloudfront.NewTestCloudfrontClient(&cloudfront.MockCloudFrontClient{})

	ds := New(testConfig, testCf)

	tests := []struct {
		// inputs
		claims []string
		// outputs
		want []string
		err  error
	}{
		{
			// success
			claims: []string{"grp1"},
			want:   []string{"dis1", "dis2"},
			err:    nil,
		},
		{
			// success
			claims: []string{"grp2"},
			want:   []string{"dis2"},
			err:    nil,
		},
		{
			// empty claims
			claims: []string{},
			want:   []string{},
			err:    errors.New("no claims present"),
		},
		{
			// claim doesn't exist in entitlement but expect back empty list
			claims: []string{"grp3"},
			want:   []string{},
			err:    nil,
		},
	}

	for _, test := range tests {
		ctx := addClaims(context.Background(), test.claims)

		ret, err := ds.List(ctx)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.want, ret)
		}
	}
}

func TestCreateInvalidation(t *testing.T) {
	testConfig, err := newTestConfig()
	assert.NoError(t, err)

	tests := []struct {
		// inputs
		claims           []string
		distributionName string
		paths            []string
		mockCf           *cloudfront.MockCloudFrontClient
		// outputs
		want *InvalidationResponse
		err  error
	}{
		{
			// success
			claims:           []string{"grp1"},
			distributionName: "dis1",
			paths:            []string{"/foo/*", "/foo/a/*", "/foo/a%20bb%2Ec/bar/*", "/foo/bar/../*", "/foo/bar/%2e%2e%2f/*"},
			mockCf: &cloudfront.MockCloudFrontClient{
				Err:            nil,
				CreateTime:     time.Unix(0, 0).UTC(),
				InvalidationId: "ABC123",
				Status:         "In Progress",
			},
			want: &InvalidationResponse{
				InvalidationMeta: InvalidationMeta{
					Status: "In Progress",
				},
				ID:      "ABC123",
				Created: time.Unix(0, 0).UTC(),
				Paths:   []string{"/foo/*", "/foo/a/*", "/foo/a%20bb%2Ec/bar/*", "/foo/*", "/foo/bar/%2e%2e%2f/*"}, // result should be cleaned paths with encoding
			},
			err: nil,
		},
		{
			// error, unauthorized paths
			claims:           []string{"grp1"},
			distributionName: "dis1",
			paths:            []string{"/a/*", "/foo/a/b", "/a/../*", ".."},
			mockCf:           &cloudfront.MockCloudFrontClient{},
			want:             nil,
			err:              NewInvalidationError(BadRequestErrorCode, errors.New("unauthorized paths"), fmt.Errorf("unauthorized paths: %v", []string{"/a/*", "/a/../*", ".."})),
		},
		{
			// error, unauthorized paths with URL encoding
			claims:           []string{"grp1"},
			distributionName: "dis1",
			paths:            []string{"/foo/%2e%2e%2f/*", "/foo/a/..%2f/%2e%2e/*"},
			mockCf:           &cloudfront.MockCloudFrontClient{},
			want:             nil,
			err:              NewInvalidationError(BadRequestErrorCode, errors.New("unauthorized paths"), fmt.Errorf("unauthorized paths: %v", []string{"/foo/%2e%2e%2f/*", "/foo/a/..%2f/%2e%2e/*"})),
		},
		{
			// error invalid URL encoding
			claims:           []string{"grp1"},
			distributionName: "dis1",
			paths:            []string{"/foo/ab%2/*"},
			mockCf:           &cloudfront.MockCloudFrontClient{},
			want:             nil,
			err:              NewInvalidationError(BadRequestErrorCode, errors.New("invalid encoded path"), fmt.Errorf("invalid encoded path: %v", "/foo/ab%2/*")),
		},
		{
			// error from cloudfront api
			claims:           []string{"grp2"},
			distributionName: "dis2",
			paths:            []string{"/bar/*"},
			mockCf:           &cloudfront.MockCloudFrontClient{Err: errors.New("mock cloudfront error")},
			want:             nil,
			err:              NewInvalidationError(BadRequestErrorCode, errors.New("cloudfront CreateInvalidation failed"), errors.New("mock cloudfront error")),
		},
	}

	for _, test := range tests {
		cfClient := cloudfront.NewTestCloudfrontClient(test.mockCf)
		ds := New(testConfig, cfClient)

		ctx := addClaims(context.Background(), test.claims)

		ret, err := ds.CreateInvalidation(ctx, test.distributionName, test.paths)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.want, ret)
		}
	}
}

func TestGetInvalidationStatus(t *testing.T) {
	testConfig, err := newTestConfig()
	assert.NoError(t, err)

	tests := []struct {
		// inputs
		claims           []string
		distributionName string
		invalidationId   string
		mockCf           *cloudfront.MockCloudFrontClient
		// outputs
		want *InvalidationResponse
		err  error
	}{
		{
			// success
			claims:           []string{"grp1"},
			distributionName: "dis1",
			invalidationId:   "ABC123",
			mockCf: &cloudfront.MockCloudFrontClient{
				Err:            nil,
				CreateTime:     time.Unix(0, 0).UTC(),
				InvalidationId: "ABC123",
				Status:         "Completed",
				Paths:          []string{"/*", "/foo/*"},
			},
			want: &InvalidationResponse{
				InvalidationMeta: InvalidationMeta{
					Status: "Completed",
				},
				ID:      "ABC123",
				Created: time.Unix(0, 0).UTC(),
				Paths:   []string{"/*", "/foo/*"},
			},
			err: nil,
		},
		{
			// error from cloudfront api
			claims:           []string{"grp2"},
			distributionName: "dis2",
			invalidationId:   "ABC123",
			mockCf:           &cloudfront.MockCloudFrontClient{Err: errors.New("mock cloudfront error")},
			want:             nil,
			err:              NewInvalidationError(BadRequestErrorCode, fmt.Errorf("cloudfront GetInvalidation failed"), errors.New("mock cloudfront error")),
		},
	}

	for _, test := range tests {
		cfClient := cloudfront.NewTestCloudfrontClient(test.mockCf)
		ds := New(testConfig, cfClient)

		ctx := addClaims(context.Background(), test.claims)

		ret, err := ds.GetInvalidationStatus(ctx, test.distributionName, test.invalidationId)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.want, ret)
		}
	}
}
