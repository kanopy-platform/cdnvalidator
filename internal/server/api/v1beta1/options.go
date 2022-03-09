package v1beta1

import (
	"time"

	"github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
)

type Options struct {
	distributionServiceOptions []v1beta1.Option
}

type Option func(o *Options)

func WithConfigFile(name string) Option {
	return func(o *Options) {
		o.distributionServiceOptions = append(o.distributionServiceOptions, v1beta1.WithConfigFile(name))
	}
}

func WithAwsRegion(region string) Option {
	return func(o *Options) {
		o.distributionServiceOptions = append(o.distributionServiceOptions, v1beta1.WithAwsRegion(region))
	}
}

func WithAwsStaticCredentials(key string, secret string) Option {
	return func(o *Options) {
		o.distributionServiceOptions = append(o.distributionServiceOptions, v1beta1.WithAwsStaticCredentials(key, secret))
	}
}

func WithTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.distributionServiceOptions = append(o.distributionServiceOptions, v1beta1.WithTimeout(t))
	}
}
