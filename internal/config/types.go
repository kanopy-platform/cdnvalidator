package config

import (
	"fmt"
	"sync"
)

type distributionName = string
type claimName = string
type distributionsMap map[distributionName]*Distribution
type entitlementsMap map[claimName][]distributionName

type Distribution struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
}

// StringPropertiesHash concatenates all string properties in Distribution
// to form a unique hash
func (d *Distribution) stringPropertiesHash() string {
	return fmt.Sprintf("%s%s", d.ID, d.Prefix)
}

type Distributions struct {
	mu      sync.RWMutex
	entries distributionsMap
}

func (d *Distributions) Get(key string) (*Distribution, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := &Distribution{}
	_, ok := d.entries[key]

	if ok {
		result.ID = d.entries[key].ID
		result.Prefix = d.entries[key].Prefix
	}

	return result, ok
}

func (d *Distributions) Set(key string, value *Distribution) {
	d.mu.Lock()
	d.entries[key] = value
	d.mu.Unlock()
}

func (d *Distributions) Delete(key string) {
	d.mu.Lock()
	delete(d.entries, key)
	d.mu.Unlock()
}

func (d *Distributions) Names() []string {
	var names []string

	d.mu.RLock()
	for name := range d.entries {
		names = append(names, name)
	}
	d.mu.RUnlock()

	return names
}

type Entitlements struct {
	mu      sync.RWMutex
	entries entitlementsMap
}

func (e *Entitlements) Get(key string) ([]distributionName, bool) {
	e.mu.RLock()
	result, ok := e.entries[key]
	e.mu.RUnlock()

	return result, ok
}

func (e *Entitlements) Set(key string, value []distributionName) {
	e.mu.Lock()
	e.entries[key] = value
	e.mu.Unlock()
}

func (e *Entitlements) Delete(key string) {
	e.mu.Lock()
	delete(e.entries, key)
	e.mu.Unlock()
}

func (e *Entitlements) Names() []string {
	var names []string

	e.mu.RLock()
	for name := range e.entries {
		names = append(names, name)
	}
	e.mu.RUnlock()

	return names
}

type Config struct {
	distributions Distributions
	entitlements  Entitlements
}

func New() *Config {
	config := &Config{}
	config.distributions.entries = make(distributionsMap)
	config.entitlements.entries = make(entitlementsMap)

	return config
}
