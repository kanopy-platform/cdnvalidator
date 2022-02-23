package cloudfront

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	cf "github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/stretchr/testify/assert"
)

type mockCloudFrontClient struct {
	cloudfrontiface.CloudFrontAPI
	returnError     bool
	createTime      time.Time
	paths           []string
	distributionId  string
	invalidationId  string
	callerReference string
	status          string
}

func (m *mockCloudFrontClient) CreateInvalidationWithContext(aws.Context, *cf.CreateInvalidationInput, ...request.Option) (*cf.CreateInvalidationOutput, error) {
	if m.returnError {
		return nil, fmt.Errorf("mock cloudfront error")
	}

	output := &cf.CreateInvalidationOutput{
		Invalidation: &cf.Invalidation{
			CreateTime: aws.Time(m.createTime),
			Id:         aws.String(m.invalidationId),
			InvalidationBatch: &cf.InvalidationBatch{
				CallerReference: aws.String(m.callerReference),
				Paths: &cf.Paths{
					Items:    aws.StringSlice(m.paths),
					Quantity: aws.Int64(int64(len(m.paths))),
				},
			},
			Status: aws.String(m.status),
		},
		Location: aws.String(""),
	}

	return output, nil
}

func (m *mockCloudFrontClient) GetInvalidationWithContext(aws.Context, *cf.GetInvalidationInput, ...request.Option) (*cf.GetInvalidationOutput, error) {
	if m.returnError {
		return nil, fmt.Errorf("mock cloudfront error")
	}

	output := &cf.GetInvalidationOutput{
		Invalidation: &cf.Invalidation{
			CreateTime: aws.Time(m.createTime),
			Id:         aws.String(m.invalidationId),
			InvalidationBatch: &cf.InvalidationBatch{
				CallerReference: aws.String(m.callerReference),
				Paths: &cf.Paths{
					Items:    aws.StringSlice(m.paths),
					Quantity: aws.Int64(int64(len(m.paths))),
				},
			},
			Status: aws.String(m.status),
		},
	}

	return output, nil
}

func newMockClient(cfClient cloudfrontiface.CloudFrontAPI) *Client {
	return &Client{
		cfApi: cfClient,
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
			distributionId:  "E28KL8GAIQAN03",
			invalidationId:  "I1JEZI55SHT2W3",
			callerReference: currentTimeSeconds(),
			status:          "Completed",
		},
	}

	for _, test := range tests {
		client := newMockClient(test)

		output, err := client.CreateInvalidation(test.distributionId, test.paths)
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
			distributionId: "E28KL8GAIQAN03",
			invalidationId: "I1JEZI55SHT2W3",
			status:         "Completed",
		},
	}

	for _, test := range tests {
		client := newMockClient(test)

		output, err := client.GetInvalidation(test.distributionId, test.invalidationId)
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
