package server

import (
	"github.com/kanopy-platform/cdnvalidator/internal/config"
	"github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
)

type Option func(*Server) error

func WithConfig(c *config.Config) Option {
	return func(s *Server) error {
		s.apiOptions = append(s.apiOptions, v1beta1.WithConfig(c))
		return nil
	}
}

func WithCloudfrontClient(c *cloudfront.Client) Option {
	return func(s *Server) error {
		s.apiOptions = append(s.apiOptions, v1beta1.WithCloudfrontClient(c))
		return nil
	}
}
