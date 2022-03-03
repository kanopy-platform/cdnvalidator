package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type configChange func(config)

func addRepeatedDistribution(originalDistribution string) configChange {
	return func(c config) {
		c.Distributions["repeated"] = c.Distributions[originalDistribution]
	}
}

func setupConfig(changes ...configChange) *Config {
	config := config{
		Distributions: Distributions{
			"dis1": {
				ID:     "123",
				Prefix: "/foo",
			},
			"dis2": {
				ID:     "456",
				Prefix: "/bar",
			},
			"dis3": {
				ID:     "789",
				Prefix: "/",
			},
			"dis4": {
				ID:     "135",
				Prefix: "/yay",
			},
		},
		Entitlements: Entitlements{
			"grp1": {
				"dis1",
				"dis2",
			},
			"grp2": {
				"dis2",
			},
			"grp3": {
				"dis2",
				"dis3",
			},
			"grp4": {
				"dis4",
			},
		},
	}

	for _, change := range changes {
		change(config)
	}

	return &Config{
		distributions: config.Distributions,
		entitlements:  config.Entitlements,
	}
}

func TestValidateDistributions(t *testing.T) {
	config := setupConfig(addRepeatedDistribution("dis1"))

	assert.Error(t, config.validateDistributions())
}

func TestParse(t *testing.T) {
	config := &Config{}

	yamlString := `---
distributions:
  dis1:
    id: "123"
    prefix: "/foo"
  dis2:
    id: "456"
    prefix: "/bar"
entitlements:
  grp1:
    - dis1
    - dis2
  grp2:
    - dis3
`
	err := config.parse([]byte(yamlString))
	assert.NoError(t, err)
}

func TestClaimDistributions(t *testing.T) {
	config := setupConfig()

	tests := []struct {
		claim string
		want  Distributions
	}{
		{
			claim: "grp1",
			want: Distributions{
				"dis1": {
					ID:     "123",
					Prefix: "/foo",
				},
				"dis2": {
					ID:     "456",
					Prefix: "/bar",
				},
			},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, config.claimDistributions(test.claim))
	}
}

func TestClaimsDistributions(t *testing.T) {
	config := setupConfig()

	tests := []struct {
		claims []string
		want   Distributions
	}{
		{
			claims: []string{"grp1", "grp3"},
			want: Distributions{
				"dis1": {
					ID:     "123",
					Prefix: "/foo",
				},
				"dis2": {
					ID:     "456",
					Prefix: "/bar",
				},
				"dis3": {
					ID:     "789",
					Prefix: "/",
				},
			},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, config.ClaimsDistributions(test.claims))
	}
}

func TestClaimsDistributionNames(t *testing.T) {
	config := setupConfig()

	tests := []struct {
		claims []string
		want   []string
	}{
		{
			claims: []string{"grp1", "grp3"},
			want:   []string{"dis1", "dis2", "dis3"},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, config.ClaimsDistributionNames(test.claims))
	}
}

func TestClaimsDistribution(t *testing.T) {
	config := setupConfig()

	claims := []string{"grp1"}

	want := Distribution{
		ID:     "123",
		Prefix: "/foo",
	}

	assert.Equal(t, want, config.ClaimsDistribution(claims, "dis1"))
	assert.Equal(t, Distribution{}, config.ClaimsDistribution(claims, "no-exists"))
}
