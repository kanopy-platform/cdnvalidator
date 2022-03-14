package cloudfront

import (
	"time"
)

type Option func(c *Client)

func WithAWSRegion(region string) Option {
	return func(c *Client) {
		if region != "" {
			c.region = region
		}
	}
}

func WithStaticCredentials(key string, secret string) Option {
	return func(c *Client) {
		c.staticCredentials.key = key
		c.staticCredentials.secret = secret
	}
}

func WithTimeout(t time.Duration) Option {
	return func(c *Client) {
		c.timeout = t
	}
}
