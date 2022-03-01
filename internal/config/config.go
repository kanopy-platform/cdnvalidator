package config

import (
	"os"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Distributions distributions `json:"distributions"`
	Entitlements entitlements `json:"entitlements"`
}

func (c *Config) Load(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	config := &Config{}
	if err = yaml.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}

type distributionList = []string
type entitlements map[string]distributionList

type distributionProperties struct {
	id string
	prefix string
}

type distributions map[string]distributionProperties

type EntitlementManager struct {
	entitlements entitlements
	distributions distributions
}

func NewEntitlementManager(config *Config) *EntitlementManager {
	return &EntitlementManager{
		entitlements: config.Entitlements,
		distributions: config.Distributions,
	}
}


// TODO This will most likely receive *http.Request instead of "path" in the future
func (e *EntitlementManager) Entitled(path string, claims []string) bool {

	var distributions []string
	for _, claim := range claims {
		if distros, ok := e.entitlements[claim]; ok {
			distributions = append(distributions, distros...)
		}
	}

	for _, distro := range distributions {
		if distroProps, ok := e.distributions[distro]; ok {
			if path == distroProps.prefix {
				return true
			}
		}
	}

	return false
}
