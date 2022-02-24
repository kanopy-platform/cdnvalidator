package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cf "github.com/aws/aws-sdk-go/service/cloudfront"
)

func NewClient(opts ...Option) (*Client, error) {
	client := &Client{
		awsCfg:  &aws.Config{},
		timeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(client)
	}

	// By default, if no StaticCredentials are provided, NewSession will use environment variables
	// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
	session, err := session.NewSession(client.awsCfg)
	if err != nil {
		return nil, err
	}

	client.cfApi = cf.New(session)

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
func checkInvalidationStruct(invalidation *cf.Invalidation) error {
	if invalidation == nil {
		return fmt.Errorf("cloudfront Invalidation struct set to nil")
	}

	if invalidation.CreateTime == nil {
		return fmt.Errorf("cloudfront Invalidation.CreateTime set to nil")
	}

	if invalidation.Id == nil {
		return fmt.Errorf("cloudfront Invalidation.Id set to nil")
	}

	if invalidation.InvalidationBatch == nil {
		return fmt.Errorf("cloudfront Invalidation.InvalidationBatch set to nil")
	}

	if invalidation.InvalidationBatch.CallerReference == nil {
		return fmt.Errorf("cloudfront Invalidation.InvalidationBatch.CallerReference set to nil")
	}

	if invalidation.InvalidationBatch.Paths == nil {
		return fmt.Errorf("cloudfront Invalidation.InvalidationBatch.Paths set to nil")
	}

	if invalidation.InvalidationBatch.Paths.Items == nil {
		return fmt.Errorf("cloudfront Invalidation.InvalidationBatch.Paths.Items set to nil")
	}

	if invalidation.Status == nil {
		return fmt.Errorf("cloudfront Invalidation.Status set to nil")
	}

	return nil
}

// Creates an Invalidation request
func (c *Client) CreateInvalidation(distributionId string, paths []string) (*CreateInvalidationOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	output, err := c.cfApi.CreateInvalidationWithContext(ctx, &cf.CreateInvalidationInput{
		DistributionId: aws.String(distributionId),
		InvalidationBatch: &cf.InvalidationBatch{
			// Using CallerReference as unique identifier for Invalidation request.
			// Effectively rate limits CreateInvalidation requests for the same (Distribution_Id, Paths)
			// to once per second.
			CallerReference: aws.String(currentTimeSeconds()),
			Paths: &cf.Paths{
				Items:    aws.StringSlice(paths),
				Quantity: aws.Int64(int64(len(paths))),
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
		InvalidationId: aws.StringValue(invalidation.Id),
		Status:         aws.StringValue(invalidation.Status),
	}

	return response, nil
}

// Retrieves information about the given invalidation request
func (c *Client) GetInvalidation(distributionId string, invalidationId string) (*GetInvalidationOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	output, err := c.cfApi.GetInvalidationWithContext(ctx, &cf.GetInvalidationInput{
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
		CreateTime: aws.TimeValue(invalidation.CreateTime),
		Status:     aws.StringValue(invalidation.Status),
		Paths:      aws.StringValueSlice(invalidation.InvalidationBatch.Paths.Items),
	}

	return response, nil
}
