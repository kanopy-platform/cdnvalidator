package config

type distributionName = string
type claimName = string
type Distributions map[distributionName]Distribution
type Entitlements map[claimName][]distributionName

type Distribution struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
}

type Config struct {
	distributions Distributions
	entitlements  Entitlements
}
