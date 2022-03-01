package cloudfront

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cf "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCloudFrontClient struct {
	returnError     bool
	createTime      time.Time
	paths           []string
	distributionId  string
	invalidationId  string
	callerReference string
	status          string
}

func (m *mockCloudFrontClient) CreateInvalidation(ctx context.Context, params *cf.CreateInvalidationInput, optFns ...func(*cf.Options)) (*cf.CreateInvalidationOutput, error) {
	if m.returnError {
		return nil, fmt.Errorf("mock cloudfront error")
	}

	output := &cf.CreateInvalidationOutput{
		Invalidation: &types.Invalidation{
			CreateTime: aws.Time(m.createTime),
			Id:         aws.String(m.invalidationId),
			InvalidationBatch: &types.InvalidationBatch{
				CallerReference: aws.String(m.callerReference),
				Paths: &types.Paths{
					Items:    m.paths,
					Quantity: aws.Int32(int32(len(m.paths))),
				},
			},
			Status: aws.String(m.status),
		},
		Location: aws.String(""),
	}

	return output, nil
}

func (m *mockCloudFrontClient) GetInvalidation(ctx context.Context, params *cf.GetInvalidationInput, optFns ...func(*cf.Options)) (*cf.GetInvalidationOutput, error) {
	if m.returnError {
		return nil, fmt.Errorf("mock cloudfront error")
	}

	output := &cf.GetInvalidationOutput{
		Invalidation: &types.Invalidation{
			CreateTime: aws.Time(m.createTime),
			Id:         aws.String(m.invalidationId),
			InvalidationBatch: &types.InvalidationBatch{
				CallerReference: aws.String(m.callerReference),
				Paths: &types.Paths{
					Items:    m.paths,
					Quantity: aws.Int32(int32(len(m.paths))),
				},
			},
			Status: aws.String(m.status),
		},
	}

	return output, nil
}

func newMockClient(cfClient cfClientAPI) *Client {
	return &Client{
		cfClient: cfClient,
	}
}

func TestCreateInvalidation(t *testing.T) {
	t.Parallel()

	tests := []*mockCloudFrontClient{
		{
			// error response from cloudfront
			returnError: true,
		},
		{
			// success
			returnError:     false,
			createTime:      time.Now(),
			paths:           []string{"/docs", "/docs-qa"},
			distributionId:  "ABCD1234ABCDEF",
			invalidationId:  "I1JEZI55SHT2W3",
			callerReference: currentTimeSeconds(),
			status:          "Completed",
		},
	}

	for _, test := range tests {
		client := newMockClient(test)

		output, err := client.CreateInvalidation(context.Background(), test.distributionId, test.paths)
		if test.returnError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.invalidationId, output.InvalidationId)
			assert.Equal(t, test.status, output.Status)
		}
	}
}

func TestGetInvalidation(t *testing.T) {
	t.Parallel()

	tests := []*mockCloudFrontClient{
		{
			// error response from cloudfront
			returnError: true,
		},
		{
			// success
			returnError:    false,
			createTime:     time.Now(),
			paths:          []string{"/docs", "/docs-qa"},
			distributionId: "ABCD1234ABCDEF",
			invalidationId: "I1JEZI55SHT2W3",
			status:         "Completed",
		},
	}

	for _, test := range tests {
		client := newMockClient(test)

		output, err := client.GetInvalidation(context.Background(), test.distributionId, test.invalidationId)
		if test.returnError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.createTime, output.CreateTime)
			assert.Equal(t, test.status, output.Status)
			assert.Equal(t, test.paths, output.Paths)
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
		WithAwsRegion("us-east-1"),
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

	log.Infof("Created Invalidation: Id=%v, Status=%v", create.InvalidationId, create.Status)

	get, err := c.GetInvalidation(context.Background(), *distributionID, create.InvalidationId)
	require.NoError(t, err)

	log.Infof("Got Invalidation %v: CreateTime=%v, Status=%v, Paths=%v", create.InvalidationId, get.CreateTime, get.Status, get.Paths)
}
