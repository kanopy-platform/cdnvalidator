package cloudfront

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cf "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func NewTestCloudfrontClient(cfClient cfClientAPI) *Client {
	return &Client{
		cfClient: cfClient,
	}
}

type MockCloudFrontClient struct {
	Err            error
	CreateTime     time.Time
	InvalidationId string
	Status         string
	// only used by GetInvalidation
	Paths           []string
	CallerReference string
}

func (m *MockCloudFrontClient) CreateInvalidation(ctx context.Context, params *cf.CreateInvalidationInput, optFns ...func(*cf.Options)) (*cf.CreateInvalidationOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	output := &cf.CreateInvalidationOutput{
		Invalidation: &types.Invalidation{
			CreateTime: aws.Time(m.CreateTime),
			Id:         aws.String(m.InvalidationId),
			InvalidationBatch: &types.InvalidationBatch{
				CallerReference: params.InvalidationBatch.CallerReference,
				Paths: &types.Paths{
					Items:    params.InvalidationBatch.Paths.Items,
					Quantity: aws.Int32(int32(len(params.InvalidationBatch.Paths.Items))),
				},
			},
			Status: aws.String(m.Status),
		},
		Location: aws.String(""),
	}

	return output, nil
}

func (m *MockCloudFrontClient) GetInvalidation(ctx context.Context, params *cf.GetInvalidationInput, optFns ...func(*cf.Options)) (*cf.GetInvalidationOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	output := &cf.GetInvalidationOutput{
		Invalidation: &types.Invalidation{
			CreateTime: aws.Time(m.CreateTime),
			Id:         params.Id,
			InvalidationBatch: &types.InvalidationBatch{
				CallerReference: aws.String(m.CallerReference),
				Paths: &types.Paths{
					Items:    m.Paths,
					Quantity: aws.Int32(int32(len(m.Paths))),
				},
			},
			Status: aws.String(m.Status),
		},
	}

	return output, nil
}
