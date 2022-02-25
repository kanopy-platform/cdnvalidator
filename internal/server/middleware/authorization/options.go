package authorization

type Option func(m *Middleware)

func WithEntitlements(e Entitler) Option {
	return func(m *Middleware) {
		m.entitlementManager = e
	}
}
