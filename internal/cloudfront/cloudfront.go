package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	cf "github.com/aws/aws-sdk-go/service/cloudfront"
)

func NewClient(id string, secret string) (*Client, error) {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	})
	if err != nil {
		return nil, err
	}

	client := &Client{
		cfApi:   cf.New(session),
		timeout: 5 * time.Second,
	}

	return client, nil
}

// currentTimeSeconds returns the current UTC time formatted using the format string
//	"20060102150405"
// which translates to: "2006-01-02 15:04:05"
func currentTimeSeconds() string {
	t := time.Now().UTC()
	return t.Format("20060102150405")
}

// checkInvalidation verifies the Invalidation struct's pointers are non-nil
func checkInvalidation(invalidation *cf.Invalidation) error {
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

// CreateInvalidation calls the aws-sdk-go API for creating a cloudfront invalidation
//
// Invalidations to the same distributionId and paths combination can be created every second.
// Any additional requests within that time interval is a no-op.
func (c *Client) CreateInvalidation(distributionId string, paths []string) (*CreateInvalidationOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	output, err := c.cfApi.CreateInvalidationWithContext(ctx, &cf.CreateInvalidationInput{
		DistributionId: aws.String(distributionId),
		InvalidationBatch: &cf.InvalidationBatch{
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

	if err := checkInvalidation(output.Invalidation); err != nil {
		return nil, err
	}
	invalidation := *output.Invalidation

	response := &CreateInvalidationOutput{
		InvalidationId: aws.StringValue(invalidation.Id),
		Status:         aws.StringValue(invalidation.Status),
	}

	return response, nil
}

// GetInvalidation calls the aws-sdk-go API for getting a cloudfront invalidation
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

	if err := checkInvalidation(output.Invalidation); err != nil {
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
