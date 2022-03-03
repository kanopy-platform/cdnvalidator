package prometheus

type Option func(m *middleware)

func WithHTTPDurationBuckets(buckets ...float64) Option {
	return func(m *middleware) {
		m.buckets = buckets
	}
}
