package core

import "context"

type DistributionName string

type Distribution struct {
}

type Distributions interface {
	List(claims []string) map[DistributionName]Distribution
	CreateInvalidation(ctx context.Context, distributionID string, claims []string)
}
