package cloudfront

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	cf "github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

// Defines the set of APIs from aws-sdk-go-v2/service/cloudfront required
// This abstraction allows mocking these methods in _test.go
type cfClientAPI interface {
	CreateInvalidation(ctx context.Context, params *cf.CreateInvalidationInput, optFns ...func(*cf.Options)) (*cf.CreateInvalidationOutput, error)
	GetInvalidation(ctx context.Context, params *cf.GetInvalidationInput, optFns ...func(*cf.Options)) (*cf.GetInvalidationOutput, error)
}

type Client struct {
	cfClient      cfClientAPI
	awsCfgOptions []func(*config.LoadOptions) error
	timeout       time.Duration
}

type CreateInvalidationOutput struct {
	InvalidationId string
	Status         string
}

type GetInvalidationOutput struct {
	CreateTime time.Time
	Status     string
	Paths      []string
}
