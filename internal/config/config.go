package config

import (
	"os"
	"sigs.k8s.io/yaml"
)

type distributionName = string
type claimName = string

type Config struct {
	Distributions map[distributionName]Distribution `json:"distributions"`
	Entitlements  map[claimName][]distributionName  `json:"entitlements"`
}

type Distribution struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
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

type Entitler interface {
	Entitled(claims []string) bool
}

type BaseEntitler struct {
	distribution string
	prefix       string
}

type ConfigEntitler struct {
	config *Config
	BaseEntitler
}

func NewConfigEntitler(config *Config, distribution, prefix string) *ConfigEntitler {
	return &ConfigEntitler{
		config: config,
		BaseEntitler: BaseEntitler{
			distribution: distribution,
			prefix:       prefix,
		},
	}
}

func (ce *ConfigEntitler) Entitled(claims []string) bool {
	for _, claim := range claims {
		if distros, ok := ce.config.Entitlements[claim]; ok {
			for _, distro := range distros {
				if distro == ce.distribution {
					if distroInfo, ok := ce.config.Distributions[distro]; ok {
						if distroInfo.Prefix == ce.prefix {
							return true
						}
					}

				}
			}
		}
	}

	return false
}
