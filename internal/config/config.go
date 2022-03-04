package config

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

// validateDistributions checks that the condition that
// two distributions with the same distribution ID MUST NOT share the same prefix.
// or in other terms, every pair of id,prefix (Distribution) must be unique
func validateDistributions(distributions distributionsMap) error {
	uniqueMap := make(map[string]struct{})

	for _, value := range distributions {
		hash := value.stringPropertiesHash()
		if _, ok := uniqueMap[hash]; ok {
			return fmt.Errorf("error parsing configuration: distribution value duplicated id: %s prefix: %s", value.ID, value.Prefix)
		}

		uniqueMap[hash] = struct{}{}
	}

	return nil
}

func (c *Config) parse(data []byte) error {
	config := struct {
		Distributions distributionsMap `json:"distributions"`
		Entitlements  entitlementsMap  `json:"entitlements"`
	}{}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	err := validateDistributions(config.Distributions)
	if err != nil {
		return err
	}

	for name, value := range config.Distributions {
		c.distributions.Set(name, value)
	}
	for _, name := range c.distributions.Names() {
		if _, ok := config.Distributions[name]; !ok {
			c.distributions.Delete(name)
		}
	}

	for name, value := range config.Entitlements {
		c.entitlements.Set(name, value)
	}
	for _, name := range c.entitlements.Names() {
		if _, ok := config.Entitlements[name]; !ok {
			c.entitlements.Delete(name)
		}
	}

	return nil
}

func (c *Config) load(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := c.parse(data); err != nil {
		return err
	}

	return nil
}

func (c *Config) Watch(filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(filePath); err != nil {
		return err
	}

	if err := c.load(filePath); err != nil {
		return fmt.Errorf("error loading configuration: %v", err)
	}

	go c.watcher(filePath, watcher)
	return nil
}

func (c *Config) watcher(filePath string, watcher *fsnotify.Watcher) {
	defer watcher.Close()
	for {
		select {
		case event := <-watcher.Events:
			reload := false

			// Mounted files are symlinks. When the kubelet refreshes the file it is removing
			// and adding a symlink.  Therefore, when we see a remove event we know that a reload
			// needs to take place.
			// https://kubernetes.io/docs/concepts/configuration/secret/#secret-files-permissions
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				if err := watcher.Remove(event.Name); err != nil {
					log.Errorf("error removing watcher from configuration: %v", err)
				}
				if err := watcher.Add(event.Name); err != nil {
					log.Errorf("error re-watching configuration: %v", err)
				}
				reload = true
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				reload = true
			}

			if reload {
				if err := c.load(event.Name); err != nil {
					log.Errorf("error refreshing configuration: %v", err)
				} else {
					log.Info("configuration refreshed")
				}
			}
		case err, ok := <-watcher.Errors:
			log.Errorf("error on reload watcher: %v", err)
			if !ok {
				return
			}
		}
	}
}

// DistributionsFromClaims returns a lookup map of Distribution names
func (c *Config) DistributionsFromClaims(claims []string) map[string]bool {
	lookup := make(map[string]bool)

	for _, claim := range claims {
		if distros, ok := c.entitlements.Get(claim); ok {
			for _, distro := range distros {
				if _, ok := c.distributions.Get(distro); ok {
					lookup[distro] = true
				}
			}
		}
	}

	return lookup
}

// Distribution returns a specific Distribution by name
func (c *Config) Distribution(name string) *Distribution {
	if distro, ok := c.distributions.Get(name); ok {
		return distro
	}

	return nil
}
