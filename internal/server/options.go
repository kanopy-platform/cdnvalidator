package server

type Option func(*Server) error

func WithAuthCookieName(name string) Option {
	return func(s *Server) error {
		s.authCookieName = name
		return nil
	}
}
