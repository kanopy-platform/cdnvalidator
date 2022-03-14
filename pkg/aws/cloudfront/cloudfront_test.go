package cloudfront

import (
	"context"
	"errors"
	"flag"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateInvalidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		distributionId string
		paths          []string
		cfClient       *MockCloudFrontClient
	}{
		{
			// error response from cloudfront
			distributionId: "ABCD1234ABCDEF",
			paths:          []string{"/*"},
			cfClient:       &MockCloudFrontClient{Err: errors.New("mock cloudfront error")},
		},
		{
			// success
			distributionId: "ABCD1234ABCDEF",
			paths:          []string{"/docs", "/docs-qa"},
			cfClient: &MockCloudFrontClient{
				Err:            errors.New("mock cloudfront error"),
				CreateTime:     time.Now(),
				InvalidationId: "I1JEZI55SHT2W3",
				Status:         "Completed",
			},
		},
	}

	for _, test := range tests {
		client := NewTestCloudfrontClient(test.cfClient)

		output, err := client.CreateInvalidation(context.Background(), test.distributionId, test.paths)
		if test.cfClient.Err != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.cfClient.InvalidationId, output.InvalidationID)
			assert.Equal(t, test.cfClient.Status, output.Status)
		}
	}
}

func TestGetInvalidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		distributionId string
		invalidationId string
		cfClient       *MockCloudFrontClient
	}{
		{
			// error response from cloudfront
			distributionId: "ABCD1234ABCDEF",
			invalidationId: "I1JEZI55SHT2W3",
			cfClient:       &MockCloudFrontClient{Err: errors.New("mock cloudfront error")},
		},
		{
			// success
			distributionId: "ABCD1234ABCDEF",
			invalidationId: "I1JEZI55SHT2W3",
			cfClient: &MockCloudFrontClient{
				Err:        errors.New("mock cloudfront error"),
				CreateTime: time.Now(),
				Status:     "Completed",
				Paths:      []string{"/docs", "/docs-qa"},
			},
		},
	}
	for _, test := range tests {
		client := NewTestCloudfrontClient(test.cfClient)

		output, err := client.GetInvalidation(context.Background(), test.distributionId, test.invalidationId)
		if test.cfClient.Err != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.cfClient.CreateTime, output.CreateTime)
			assert.Equal(t, test.cfClient.Status, output.Status)
			assert.Equal(t, test.cfClient.Paths, output.Paths)
		}
	}
}

var distributionID = flag.String("distribution", "", "A Cloudfront distribution ID to perform an invalidation against.")
var pathsArg = flag.String("paths", "", "Comma separated list of paths")
var accessID = flag.String("access-id", "", "Default uses local aws profile")
var accessSecret = flag.String("access-secret", "", "Default uses local aws profile")

func TestIntegrationCloudfrontInvalidation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	if *distributionID == "" {
		t.Fatal("-distribution missing")
	}

	if *pathsArg == "" {
		t.Fatal("-paths missing")
	}

	opts := []Option{
		WithAWSRegion("us-east-1"),
		WithTimeout(time.Duration(30) * time.Second),
	}

	if *accessID != "" && *accessSecret != "" {
		opts = append(opts, WithStaticCredentials(*accessID, *accessSecret))
	}

	c, err := New(opts...)
	require.NoError(t, err)

	log.Info("Creating Invalidation...")

	paths := strings.Split(*pathsArg, ",")

	create, err := c.CreateInvalidation(context.Background(), *distributionID, paths)
	require.NoError(t, err)

	log.Infof("Created Invalidation: Id=%v, Status=%v", create.InvalidationID, create.Status)

	get, err := c.GetInvalidation(context.Background(), *distributionID, create.InvalidationID)
	require.NoError(t, err)

	log.Infof("Got Invalidation %v: CreateTime=%v, Status=%v, Paths=%v", create.InvalidationID, get.CreateTime, get.Status, get.Paths)
}
