package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEntitled(t *testing.T) {
	config := Config{
		Distributions: distributions{
			"vanity0": {
				id: "12345",
				prefix: "/",
			},
			"vanity1": {
				id: "12345",
				prefix: "/foo",
			},
			"vanity3": {
				id: "4567",
				prefix: "/bar",
			},
		},
		Entitlements: entitlements{
			"one": {
				"vanity1",
				"vanity2",
			},
			"two": {
				"vanity3",
			},
			"three": {
				"vanity4",
			},
		},
	}

	claims := []string{"one", "two"}

	em := NewEntitlementManager(&config)

	assert.Equal(t, true, em.Entitled("/foo", claims))
	assert.Equal(t, true, em.Entitled("/bar", claims))
	assert.Equal(t, false, em.Entitled("/some", claims))
}
