package authorization

type Option func(m *Middleware)

func WithEntitlements(e Entitler) Option {
	return func(m *Middleware) {
		m.entitlementManager = e
	}
}

func WithCookieName(name string) Option {
	return func(m *Middleware) {
		m.authCookieName = name
	}
}

func WithAuthorizationHeader() Option {
	return func(m *Middleware) {
		m.authHeaderEnabled = true
	}
}
