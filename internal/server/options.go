package server

type Option func(*Server) error

func WithAuthCookieName(name string) Option {
	return func(s *Server) error {
		s.authCookieName = name
		return nil
	}
}

func WithAuthHeaderName(name string) Option {
	return func(s *Server) error {
		s.authHeaderName = name
		return nil
	}
}
