package config

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

// two vanity distributions with the same distribution ID MUST NOT share the same prefix.
// or in other terms, every pair of id,prefix (Distribution) must be unique
func (c *Config) validateDistributions() error {
	uniqueMap := make(map[Distribution]struct{})

	for _, value := range c.distributions {
		if _, ok := uniqueMap[value]; ok {
			return fmt.Errorf("error parsing configuration: distribution value duplicated id: %s prefix: %s", value.ID, value.Prefix)
		}

		uniqueMap[value] = struct{}{}
	}

	return nil
}

func (c *Config) parse(data []byte) error {
	config := struct {
		Distributions Distributions `json:"distributions"`
		Entitlements  Entitlements  `json:"entitlements"`
	}{}
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

func (c *Config) Watch(filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(filePath); err != nil {
		return err
	}

	if err := c.Load(filePath); err != nil {
		return fmt.Errorf("error loading configuration: %v", err)
	}

	go c.watch(filePath, watcher)
	return nil
}

func (c *Config) watch(filePath string, watcher *fsnotify.Watcher) {
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
				watcher.Remove(event.Name)
				if err := watcher.Add(event.Name); err != nil {
					log.Errorf("Error re-watching revoked token list: %v", err)
				}
				reload = true
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				reload = true
			}

			if reload {
				if err := c.Load(event.Name); err != nil {
					log.Errorf("Error refreshing revoked token list: %v", err)
				} else {
					log.Info("revoked tokens list refreshed")
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
