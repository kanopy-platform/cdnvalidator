package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEntitlements(t *testing.T) {
	c := &Config{
		Entitlements: map[string][]VanityDistrbutionName{
			"grp1": {
				"vn1",
				"vn2",
			},
			"grp2": {
				"vn1",
			},
			"grp3": {
				"vn2",
			},
		},

		VanityDistrbutions: VanityDistrbution{
			"vn1": Entitlement{
				DistributionID: "d1",
				Prefix:         "/",
			},
			"vn2": Entitlement{
				DistributionID: "d1",
				Prefix:         "/",
			},
		},
	}

	tests := []struct {
		boundaries []string
		want       []Entitlement
	}{
		{
			boundaries: []string{"grp1", "grp2"},
			want: []Entitlement{
				c.VanityDistrbutions["vn1"],
				c.VanityDistrbutions["vn2"],
			},
		},
		{
			boundaries: []string{"grp2"},
			want: []Entitlement{
				c.VanityDistrbutions["vn1"],
			},
		},
		{
			boundaries: []string{"grp1"},
			want: []Entitlement{
				c.VanityDistrbutions["vn1"],
				c.VanityDistrbutions["vn2"],
			},
		},

		{
			boundaries: []string{"grp3"},
			want: []Entitlement{
				c.VanityDistrbutions["vn2"],
			},
		},

		{
			boundaries: []string{"not-exists"},
			want:       []Entitlement{},
		},
	}

	for _, test := range tests {
		assert.Equal(t, c.GetEntitlements(test.boundaries...), test.want)
	}
}
