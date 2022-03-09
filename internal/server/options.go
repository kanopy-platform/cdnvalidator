package server

import "time"

type Option func(*Server) error

func WithAuthCookieName(name string) Option {
	return func(s *Server) error {
		s.authCookieName = name
		return nil
	}
}

func WithConfigFile(name string) Option {
	return func(s *Server) error {
		s.configFile = name
		return nil
	}
}

func WithAwsRegion(region string) Option {
	return func(s *Server) error {
		s.awsRegion = region
		return nil
	}
}

func WithAwsStaticCredentials(key string, secret string) Option {
	return func(s *Server) error {
		s.awsKey = key
		s.awsSecret = secret
		return nil
	}
}

func WithTimeout(t time.Duration) Option {
	return func(s *Server) error {
		s.timeout = t
		return nil
	}
}
