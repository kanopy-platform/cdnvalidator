package authorization

type Option func(m *middleware)

func WithEntitlements(e Entitler) Option {
	return func(m *middleware) {
		m.entitlementManager = e
	}
}

func WithCookieName(name string) Option {
	return func(m *middleware) {
		m.authCookieName = name
	}
}

func WithAuthorizationHeader() Option {
	return func(m *middleware) {
		m.authHeaderEnabled = true
	}
}
