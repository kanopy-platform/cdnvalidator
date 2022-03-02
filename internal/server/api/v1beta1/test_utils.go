package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/internal/core/v1beta1"
)

type Fake struct {
}

func NewFake() DistributionService {
	return &Fake{}

}

func (f *Fake) List(ctx context.Context) (map[v1beta1.VanityDistributionName]v1beta1.Distribution, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	if claims[0] == "gr1" {
		return map[v1beta1.VanityDistributionName]v1beta1.Distribution{
			"f1": {
				DistributionID: "d1",
				PathPrefix:     "/",
			},
		}, nil
	}

	return nil, nil
}

func (f *Fake) CreateInvalidation(ctx context.Context, distributionName string, paths []string) (*v1beta1.InvalidationResponse, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	if distributionName == "notfound" {
		return nil, v1beta1.NewInvalidationError(v1beta1.DistributionNotFoundErrorCode, fmt.Errorf("distribution %s not found", distributionName))
	}

	if distributionName == "unauthorized path" {
		return nil, v1beta1.NewInvalidationError(v1beta1.InvalidationUnAuthorizedErrorCode, fmt.Errorf("path unauthorized"), paths[0])
	}

	return &v1beta1.InvalidationResponse{InvalidationMeta: v1beta1.InvalidationMeta{Status: "OK"}}, nil
}

func (f *Fake) GetInvalidationStatus(ctx context.Context, distributionName string, invalidationID string) (*v1beta1.InvalidationResponse, error) {
	return nil, nil
}
