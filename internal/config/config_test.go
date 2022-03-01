package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEntitled(t *testing.T) {
	config := &Config{
		Distributions: map[distributionName]Distribution{
			"vanity0": {
				ID: "12345",
				Prefix: "/",
			},
			"vanity1": {
				ID: "12345",
				Prefix: "/foo",
			},
			"vanity3": {
				ID: "4567",
				Prefix: "/bar",
			},
		},
		Entitlements: map[claimName][]distributionName{
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

	em := NewConfigEntitler(config, "vanity1", "/foo")

	assert.Equal(t, true, em.Entitled(claims))
	assert.Equal(t, true, em.Entitled(claims))
}
