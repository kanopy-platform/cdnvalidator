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

type Config struct {
	mu            sync.Mutex
	distributions distributionsMap
	entitlements  entitlementsMap
}
