package config

// implements the authorization EntitlementGetter interface
func (c *Config) GetEntitlements(distrbution string, boundaries ...string) []Entitlement {
	es := []Entitlement{}

	used := make(map[string]int)

	for _, g := range boundaries {
		if vanityNames, ok := c.Entitlements[g]; ok {
			for _, v := range vanityNames {
				if string(v) == distrbution {
					if _, ok := used[string(v)]; !ok {
						es = append(es, c.VanityDistrbutions[v])
						used[string(v)] = 1
					}
				}
			}
		}
	}
	return es
}
