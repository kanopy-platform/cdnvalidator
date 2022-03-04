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
	config := &Config{
		distributions: Distributions{
			"dis1": {
				ID:     "123",
				Prefix: "/foo",
			},
			"dis2": {
				ID:     "456",
				Prefix: "/bar",
			},
		},
		entitlements: Entitlements{
			"grp1": {
				"dis1",
				"dis2",
			},
			"grp2": {
				"dis2",
			},
		},
	}

	return config
}

func TestValidateDistributions(t *testing.T) {
	tests := []struct {
		config *Config
		want   error
	}{
		{
			config: setupConfig(),
			want:   nil,
		},
		{
			config: &Config{
				distributions: Distributions{
					"dis1": {
						ID:     "123",
						Prefix: "/foo",
					},
					"dis2": {
						ID:     "456",
						Prefix: "/bar",
					},
					"repeated": {
						ID:     "123",
						Prefix: "/foo",
					},
				},
				entitlements: Entitlements{},
			},
			want: errors.New("error parsing configuration: distribution value duplicated id: 123 prefix: /foo"),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, test.config.validateDistributions())
	}
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

func TestLoad(t *testing.T) {
	config := &Config{}

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

	assert.NoError(t, config.Load(tmpFile.Name()))
}
