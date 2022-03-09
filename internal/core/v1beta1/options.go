package v1beta1

import "time"

type Option func(ds *DistributionService)

func WithConfigFile(name string) Option {
	return func(ds *DistributionService) {
		ds.configFile = name
	}
}

func WithAwsRegion(region string) Option {
	return func(ds *DistributionService) {
		ds.awsRegion = region
	}
}

func WithAwsStaticCredentials(key string, secret string) Option {
	return func(ds *DistributionService) {
		ds.awsKey = key
		ds.awsSecret = secret
	}
}

func WithTimeout(t time.Duration) Option {
	return func(ds *DistributionService) {
		ds.timeout = t
	}
}
