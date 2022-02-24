package cloudfront

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

type Option func(c *Client)

func WithStaticCredentials(id string, secret string, token string) Option {
	return func(c *Client) {
		c.awsCfg.Credentials = credentials.NewStaticCredentials(id, secret, token)
	}
}

func WithTimeout(t time.Duration) Option {
	return func(c *Client) {
		c.timeout = t
	}
}
