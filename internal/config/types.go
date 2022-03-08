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
func (d *Distribution) hashKey() string {
	return fmt.Sprintf("%s%s", d.ID, d.Prefix)
}

type distributions struct {
	mu      sync.RWMutex
	entries distributionsMap
}

func (d *distributions) Get(key string) *Distribution {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if entry, ok := d.entries[key]; ok {
		return &Distribution{
			ID:     entry.ID,
			Prefix: entry.Prefix,
		}
	}

	return nil
}

func (d *distributions) Set(key string, value *Distribution) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.entries[key] = value
}

func (d *distributions) Delete(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.entries, key)
}

func (d *distributions) Names() []string {
	var names []string

	d.mu.RLock()
	defer d.mu.RUnlock()

	for name := range d.entries {
		names = append(names, name)
	}

	return names
}

type entitlements struct {
	mu      sync.RWMutex
	entries entitlementsMap
}

func (e *entitlements) Get(key string) ([]distributionName, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result, ok := e.entries[key]

	return result, ok
}

func (e *entitlements) Set(key string, value []distributionName) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.entries[key] = value
}

func (e *entitlements) Delete(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.entries, key)
}

func (e *entitlements) Names() []string {
	var names []string

	e.mu.RLock()
	defer e.mu.RUnlock()

	for name := range e.entries {
		names = append(names, name)
	}

	return names
}

type Config struct {
	mu            sync.Mutex
	distributions distributions
	entitlements  entitlements
}

func New() *Config {
	config := &Config{}
	config.distributions.entries = make(distributionsMap)
	config.entitlements.entries = make(entitlementsMap)

	return config
}
