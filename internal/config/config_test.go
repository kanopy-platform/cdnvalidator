package config

import (
	"context"
	"os"
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func emptyConfig() *Config {
	config := &Config{}

	config.distributions = make(distributionsMap)
	config.entitlements = make(entitlementsMap)

	return config
}

func setupConfig() *Config {
	config := emptyConfig()

	config.distributions["dis1"] = &Distribution{ID: "123", Prefix: "/foo"}
	config.distributions["dis2"] = &Distribution{ID: "456", Prefix: "/bar"}
	config.entitlements["grp1"] = []string{"dis1", "dis2"}
	config.entitlements["grp2"] = []string{"dis2"}

	return config
}

func TestValidateDistributions(t *testing.T) {
	distributions := distributionsMap{
		"dis1": {
			ID:     "123",
			Prefix: "/foo",
		},
		"dis2": {
			ID:     "456",
			Prefix: "/bar",
		},
	}

	repeatedDistributions := make(distributionsMap)
	for name, value := range distributions {
		repeatedDistributions[name] = value
	}

	repeatedDistributions["repeated"] = distributions["dis1"]

	tests := []struct {
		distros distributionsMap
		want    error
	}{
		{
			distros: distributions,
			want:    nil,
		},
		{
			distros: repeatedDistributions,
			want:    errors.New("error parsing configuration: distribution value duplicated id: 123 prefix: /foo"),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, validateDistributions(test.distros))
	}
}

func TestValidateEntitlements(t *testing.T) {
	distributions := distributionsMap{
		"dis1": {
			ID:     "123",
			Prefix: "/foo",
		},
		"dis2": {
			ID:     "456",
			Prefix: "/bar",
		},
	}

	tests := []struct {
		distributions distributionsMap
		entitlements  entitlementsMap
		want          error
	}{
		{
			distributions: distributions,
			entitlements: entitlementsMap{
				"grp1": {
					"dis1",
				},
				"grp2": {
					"dis2",
				},
			},
			want: nil,
		},
		{
			distributions: distributions,
			entitlements: entitlementsMap{
				"grp1": {
					"dis1",
					"dis2",
				},
				"grp2": {
					"dis3",
				},
			},
			want: errors.New("error parsing configuration: distribution dis3 in entitlement grp2 is not configured"),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, validateEntitlements(test.entitlements, test.distributions))
	}
}

func TestParse(t *testing.T) {
	config := emptyConfig()

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
    - dis2
`
	err := config.parse([]byte(yamlString))
	assert.NoError(t, err)

	assert.Equal(t, &Distribution{ID: "123", Prefix: "/foo"}, config.distributions["dis1"])
	grp1 := config.entitlements["grp1"]
	assert.Equal(t, []string{"dis1", "dis2"}, grp1)

	// assert concurrent access to config
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		for {
			if ctx.Err() != nil {
				return
			}

			// concurrently access values in distributions and entitlements
			config.DistributionsFromClaims([]string{"grp1"})
			config.Distribution("dis2")
		}
	}()

	newYaml := `---
distributions:
  dis1:
    id: "123"
    prefix: "/foo"
entitlements:
  grp1:
    - dis1
`

	err = config.parse([]byte(newYaml))
	assert.NoError(t, err)
	cancel()
}

func TestLoad(t *testing.T) {
	config := emptyConfig()

	tmpFile, err := os.CreateTemp(os.TempDir(), "cdnvalidator-")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fileConfig := setupConfig()
	data, err := yaml.Marshal(fileConfig)
	assert.NoError(t, err)

	_, err = tmpFile.Write(data)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	assert.NoError(t, config.load(tmpFile.Name()))
}

func TestDistributionsFromClaims(t *testing.T) {
	config := setupConfig()

	tests := []struct {
		claims []string
		want   map[string]bool
	}{
		{
			claims: []string{"grp1"},
			want:   map[string]bool{"dis1": true, "dis2": true},
		},
		{
			claims: []string{"grp2"},
			want:   map[string]bool{"dis2": true},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, config.DistributionsFromClaims(test.claims))
	}
}

func TestDistribution(t *testing.T) {
	config := setupConfig()
	want := &Distribution{ID: "123", Prefix: "/foo"}

	assert.Equal(t, want, config.Distribution("dis1"))
	assert.Nil(t, config.Distribution("no-exists"))
}
