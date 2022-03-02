package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntitled(t *testing.T) {
	config := &Config{
		Distributions: map[distributionName]Distribution{
			"dis1": {
				ID:     "12345",
				Prefix: "/foo",
			},
			"dis2": {
				ID:     "12345",
				Prefix: "/bar",
			},
			"dis3": {
				ID:     "4567",
				Prefix: "/",
			},
		},
		Entitlements: map[claimName][]distributionName{
			"grp1": {
				"dis1",
				"dis2",
			},
			"grp2": {
				"dis2",
			},
		},
	}

	tests := []struct {
		name         string
		distribution string
		claims       []string
		prefix       string
		want         bool
	}{
		{
			name:         "multiple claims in entitlement",
			distribution: "dis1",
			claims:       []string{"grp1", "grp2"},
			prefix:       "/foo",
			want:         true,
		},
		{
			name:         "one claim in entitlement",
			distribution: "dis2",
			claims:       []string{"grp2"},
			prefix:       "/bar",
			want:         true,
		},
		{
			name:         "invalid claim in config",
			distribution: "dis1",
			claims:       []string{"grp3"},
			prefix:       "/",
			want:         false,
		},
		{
			name:         "invalid prefix",
			distribution: "dis1",
			claims:       []string{"grp1"},
			prefix:       "/no-exists",
			want:         false,
		},
	}

	for _, test := range tests {
		em := NewConfigEntitler(config, test.distribution, test.prefix)
		t.Logf("Running test %s", test.name)
		assert.Equal(t, test.want, em.Entitled(test.claims))
	}
}
