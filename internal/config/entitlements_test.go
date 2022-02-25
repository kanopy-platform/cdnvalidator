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
				DistributionID: "d2",
				Prefix:         "/",
			},
		},
	}

	tests := []struct {
		distribution string
		boundaries   []string
		want         []Entitlement
	}{
		{
			distribution: "vn1",
			boundaries:   []string{"grp1", "grp2"},
			want: []Entitlement{
				c.VanityDistrbutions["vn1"],
			},
		},
		{
			distribution: "vn1",
			boundaries:   []string{"grp2"},
			want: []Entitlement{
				c.VanityDistrbutions["vn1"],
			},
		},

		{
			distribution: "vn1",
			boundaries:   []string{"grp1"},
			want: []Entitlement{
				c.VanityDistrbutions["vn1"],
			},
		},

		{
			distribution: "vn2",
			boundaries:   []string{"grp3"},
			want: []Entitlement{
				c.VanityDistrbutions["vn2"],
			},
		},

		{
			distribution: "vn2",
			boundaries:   []string{"not-exists"},
			want:         []Entitlement{},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, c.GetEntitlements(test.distribution, test.boundaries...))
	}
}
