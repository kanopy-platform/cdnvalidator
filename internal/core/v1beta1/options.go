package v1beta1

import (
	"github.com/kanopy-platform/cdnvalidator/internal/config"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
)

type Option func(ds *DistributionService)

func WithConfig(c *config.Config) Option {
	return func(ds *DistributionService) {
		ds.config = c
	}
}

func WithCloudfrontClient(c *cloudfront.Client) Option {
	return func(ds *DistributionService) {
		ds.cloudfront = c
	}
}
