package config

type Entitlement struct {
	DistributionID string `json:"distributionId"`
	Prefix         string `json:"prefix"`
}

type VanityDistrbution map[string]Entitlement

type VanityDistrbutionName string

type Config struct {
	Entitlements       map[string][]VanityDistrbutionName
	VanityDistrbutions VanityDistrbution
}
