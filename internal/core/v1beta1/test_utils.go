package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/kanopy-platform/cdnvalidator/internal/core"
)

type Fake struct {
}

func NewFake() *Fake {
	return &Fake{}

}

func (f *Fake) List(ctx context.Context) (map[VanityDistributionName]Distribution, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	if claims[0] == "gr1" {
		return map[VanityDistributionName]Distribution{
			"f1": {
				DistributionID: "d1",
				PathPrefix:     "/",
			},
		}, nil
	}

	return nil, nil
}

func (f *Fake) CreateInvalidation(ctx context.Context, distributionName string, paths []string) (*InvalidationResponse, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	if distributionName == "notfound" {
		return nil, NewInvalidationError(DistributionNotFoundErrorCode, fmt.Errorf("distribution %s not found", distributionName))
	}

	if distributionName == "unauthorized distribution" {
		return nil, NewInvalidationError(InvalidationUnAuthorizedErrorCode, fmt.Errorf("distribution unauthorized"), distributionName)
	}

	return &InvalidationResponse{InvalidationMeta: InvalidationMeta{Status: "OK"}}, nil
}

func (f *Fake) GetInvalidationStatus(ctx context.Context, distributionName string, invalidationID string) (*InvalidationResponse, error) {
	return nil, nil
}
