package config

import (
	"fmt"
	"os"
	"sort"

	"sigs.k8s.io/yaml"
)

type distributionName = string
type claimName = string
type Distributions map[distributionName]Distribution
type Entitlements map[claimName][]distributionName

type Distribution struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
}

// for marshalling purposes
type config struct {
	Distributions Distributions `json:"distributions"`
	Entitlements  Entitlements  `json:"entitlements"`
}

type Config struct {
	distributions Distributions
	entitlements  Entitlements
}

// two vanity distributions with the same distribution ID MUST NOT share the same prefix.
// or in other terms, every pair of id,prefix must be unique
func (c *Config) validateDistributions() error {

	uniqueMap := make(map[Distribution]struct{})

	for _, value := range c.distributions {
		if _, ok := uniqueMap[value]; ok {
			return fmt.Errorf("error parsing configuration: distribution value duplicated id:%s prefix:%s", value.ID, value.Prefix)
		}

		uniqueMap[value] = struct{}{}
	}

	return nil
}

func (c *Config) parse(data []byte) error {
	config := config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	c.distributions = config.Distributions
	c.entitlements = config.Entitlements

	err := c.validateDistributions()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Load(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := c.parse(data); err != nil {
		return err
	}

	return nil
}

// claimDistributions receives a claim entitlement and
// returns a map of distributions
func (c *Config) claimDistributions(claim string) Distributions {
	distributions := make(Distributions)

	if distros, ok := c.entitlements[claim]; ok {
		for _, distro := range distros {
			if _, ok := c.distributions[distro]; ok {
				distributions[distro] = c.distributions[distro]
			}
		}
	}

	return distributions
}

// claimsDistributions receive a slice of claims and
// returns a map of distributions
func (c *Config) claimsDistributions(claims []string) Distributions {
	allDistributions := make(Distributions)

	for _, claim := range claims {
		entitlementDistros := c.claimDistributions(claim)
		for k := range entitlementDistros {
			allDistributions[k] = entitlementDistros[k]
		}
	}

	return allDistributions
}

// ClaimsDistributions is the public access to claimsDistributions
func (c *Config) ClaimsDistributions(claims []string) Distributions {
	return c.claimsDistributions(claims)
}

// ClaimsDistributionNames returns a slice of distribution names
// after receiving a slice of claims
func (c *Config) ClaimsDistributionNames(claims []string) []string {
	distros := c.claimsDistributions(claims)

	names := make([]string, 0, len(distros))
	for name := range distros {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// ClaimsDistribution gets an specific distribution properties from claims
// if the distributionName is not present it returns an empty Distribution
func (c *Config) ClaimsDistribution(claims []string, distributionName string) Distribution {
	distros := c.claimsDistributions(claims)

	if distro, ok := distros[distributionName]; ok {
		return distro
	}

	return Distribution{}
}
