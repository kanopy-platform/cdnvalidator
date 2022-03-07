package config

import (
	"io/ioutil"
	"os"
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func setupConfig() *Config {
	config := New()

	config.distributions.Set("dis1", &Distribution{ID: "123", Prefix: "/foo"})
	config.distributions.Set("dis2", &Distribution{ID: "456", Prefix: "/bar"})
	config.entitlements.Set("grp1", []string{"dis1", "dis2"})
	config.entitlements.Set("grp2", []string{"dis2"})

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

func TestParse(t *testing.T) {
	config := New()

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

	// assert Set
	assert.Equal(t, &Distribution{ID: "123", Prefix: "/foo"}, config.distributions.Get("dis1"))
	grp1, _ := config.entitlements.Get("grp1")
	assert.Equal(t, []string{"dis1", "dis2"}, grp1)

	// assert Delete
	reducedYaml := `---
distributions:
  dis1:
    id: "123"
    prefix: "/foo"
entitlements:
  grp1:
    - dis1
    - dis2
`
	err = config.parse([]byte(reducedYaml))
	assert.NoError(t, err)

	assert.Nil(t, config.distributions.Get("dis2"))
	grp1, ok := config.entitlements.Get("grp2")
	assert.False(t, ok)
}

func TestLoad(t *testing.T) {
	config := New()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "cdnvalidator-")
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
