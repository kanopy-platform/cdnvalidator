package core

import (
	"context"
	"time"
)

type ClaimsKey string

const (
	ContextBoundaryKey ClaimsKey = "claims"
)

type VanityDistributionName string

type Distribution struct {
	DistributionID string `json:"-"`
	PathPrefix     string `json:"pathPrefix"`
}

type InvalidationStatus struct {
	Created time.Duration `json:"createTime"`
	Status  string        `json:"status"`
	Paths   []string      `json:"paths,omitempty"`
}

type DistributionService interface {
	List(ctx context.Context) (map[VanityDistributionName]Distribution, error)
	CreateInvalidation(ctx context.Context, distributionID string, paths []string) (*InvalidationStatus, error)
	GetInvalidationStatus(ctx context.Context, distributionID string, invalidationID string) (*InvalidationStatus, error)
}

type Fake struct {
}

func NewFake() DistributionService {
	return &Fake{}
}

func (f *Fake) List(ctx context.Context) (map[VanityDistributionName]Distribution, error) {
	return nil, nil
}

func (f *Fake) CreateInvalidation(ctx context.Context, distributionID string, paths []string) (*InvalidationStatus, error) {
	return nil, nil
}

func (f *Fake) GetInvalidationStatus(ctx context.Context, distributionID string, invalidationID string) (*InvalidationStatus, error) {
	return nil, nil
}
