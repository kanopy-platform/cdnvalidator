package authorization

type Option func(m *middleware)

func WithCookieName(name string) Option {
	return func(m *middleware) {
		m.authCookieName = name
	}
}

func WithHeaderName(name string) Option {
	return func(m *middleware) {
		m.authHeaderName = name
	}
}

func WithAuthorizationHeader() Option {
	return func(m *middleware) {
		m.authHeaderEnabled = true
	}
}
