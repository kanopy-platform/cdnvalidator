package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cf "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func New(opts ...Option) (*Client, error) {
	client := &Client{}

	// default options
	o := []Option{
		WithAWSRegion("us-east-1"),
		WithTimeout(30 * time.Second),
	}

	opts = append(o, opts...)

	for _, opt := range opts {
		opt(client)
	}

	// By default, if no StaticCredentials are provided, LoadDefaultConfig will use environment variables
	// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
	cfg, err := config.LoadDefaultConfig(context.Background(), client.awsCfgOptions...)
	if err != nil {
		return nil, err
	}

	client.cfClient = cf.NewFromConfig(cfg)

	return client, nil
}

// Returns the current UTC time formatted using the format string
//	"20060102150405"
// which means: "2006-01-02 15:04:05"
func currentTimeSeconds() string {
	t := time.Now().UTC()
	return t.Format("20060102150405")
}

// Verifies the Invalidation struct's pointers are non-nil
func checkInvalidationStruct(invalidation *types.Invalidation) error {
	if invalidation == nil {
		return fmt.Errorf("invalidation nil pointer")
	}

	fieldNilError := fmt.Errorf("invalidation struct contains a nil field")

	if invalidation.CreateTime == nil ||
		invalidation.Id == nil ||
		invalidation.InvalidationBatch == nil ||
		invalidation.Status == nil {
		return fieldNilError
	}

	if invalidation.InvalidationBatch.CallerReference == nil ||
		invalidation.InvalidationBatch.Paths == nil ||
		invalidation.InvalidationBatch.Paths.Quantity == nil {
		return fieldNilError
	}

	return nil
}

// Creates an Invalidation request
func (c *Client) CreateInvalidation(ctx context.Context, distributionId string, paths []string) (*CreateInvalidationOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	output, err := c.cfClient.CreateInvalidation(ctx, &cf.CreateInvalidationInput{
		DistributionId: aws.String(distributionId),
		InvalidationBatch: &types.InvalidationBatch{
			// Using CallerReference as unique identifier for Invalidation request.
			// Effectively rate limits CreateInvalidation requests for the same (Distribution_Id, Paths)
			// to once per second.
			CallerReference: aws.String(currentTimeSeconds()),
			Paths: &types.Paths{
				Items:    paths,
				Quantity: aws.Int32(int32(len(paths))),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if err := checkInvalidationStruct(output.Invalidation); err != nil {
		return nil, err
	}
	invalidation := *output.Invalidation

	response := &CreateInvalidationOutput{
		InvalidationID: aws.ToString(invalidation.Id),
		Status:         aws.ToString(invalidation.Status),
		CreateTime:     aws.ToTime(invalidation.CreateTime),
		Paths:          invalidation.InvalidationBatch.Paths.Items,
	}

	return response, nil
}

// Retrieves information about the given invalidation request
func (c *Client) GetInvalidation(ctx context.Context, distributionId string, invalidationId string) (*GetInvalidationOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	output, err := c.cfClient.GetInvalidation(ctx, &cf.GetInvalidationInput{
		DistributionId: aws.String(distributionId),
		Id:             aws.String(invalidationId),
	})
	if err != nil {
		return nil, err
	}

	if err := checkInvalidationStruct(output.Invalidation); err != nil {
		return nil, err
	}
	invalidation := *output.Invalidation

	response := &GetInvalidationOutput{
		InvalidationID: aws.ToString(invalidation.Id),
		Status:         aws.ToString(invalidation.Status),
		CreateTime:     aws.ToTime(invalidation.CreateTime),
		Paths:          invalidation.InvalidationBatch.Paths.Items,
	}

	return response, nil
}
