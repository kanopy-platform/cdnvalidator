package v1beta1

import (
	"context"

	"github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
)

type DistributionService interface {
	List(ctx context.Context) ([]string, error)
	CreateInvalidation(ctx context.Context, distributionName string, paths []string) (*v1beta1.InvalidationResponse, error)
	GetInvalidationStatus(ctx context.Context, distributionName string, invalidationID string) (*v1beta1.InvalidationResponse, error)
}
