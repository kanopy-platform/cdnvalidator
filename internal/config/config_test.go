package config

import (
	"fmt"
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
	config.entitlements.Set("grp1", []string{"dis1", "dis2"})

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

	repeatedDistributions["repeated"] = &Distribution{ID: "123", Prefix: "/foo"}

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
		fmt.Println(test.distros)
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

	assert.NoError(t, config.Load(tmpFile.Name()))
}
