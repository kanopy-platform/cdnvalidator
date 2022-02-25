package authorization

type Option func(m *Middleware)

func WithEntitlements(e EntitlementGetter) Option {
	return func(m *Middleware) {
		m.entitlement = e
	}
}
