package cloudfront

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type Option func(c *Client)

func WithAwsRegion(region string) Option {
	return func(c *Client) {
		c.awsCfgOptions = append(c.awsCfgOptions, config.WithRegion(region))
	}
}

func WithStaticCredentials(key string, secret string) Option {
	return func(c *Client) {
		c.awsCfgOptions = append(c.awsCfgOptions, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")))
	}
}

func WithTimeout(t time.Duration) Option {
	return func(c *Client) {
		c.timeout = t
	}
}
