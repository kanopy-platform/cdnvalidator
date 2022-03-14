package cloudfront

import (
	"context"
	"time"

	cf "github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

// Defines the set of APIs from aws-sdk-go-v2/service/cloudfront required
// This abstraction allows mocking these methods in _test.go
type cfClientAPI interface {
	CreateInvalidation(ctx context.Context, params *cf.CreateInvalidationInput, optFns ...func(*cf.Options)) (*cf.CreateInvalidationOutput, error)
	GetInvalidation(ctx context.Context, params *cf.GetInvalidationInput, optFns ...func(*cf.Options)) (*cf.GetInvalidationOutput, error)
}

type Client struct {
	cfClient          cfClientAPI
	region            string
	staticCredentials awsStaticCredentials
	timeout           time.Duration
}

type awsStaticCredentials struct {
	key    string
	secret string
}

type CreateInvalidationOutput struct {
	InvalidationID string
	Status         string
	CreateTime     time.Time
	Paths          []string
}

type GetInvalidationOutput struct {
	InvalidationID string
	Status         string
	CreateTime     time.Time
	Paths          []string
}
