package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	if err := checkErrors(distributionName, ""); err != nil {
		return nil, err
	}

	return &InvalidationResponse{InvalidationMeta: InvalidationMeta{Status: "OK"}}, nil
}

func (f *Fake) GetInvalidationStatus(ctx context.Context, distributionName string, invalidationID string) (*InvalidationResponse, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	if err := checkErrors(distributionName, invalidationID); err != nil {
		return nil, err
	}

	return &InvalidationResponse{
		ID:               "1",
		Created:          time.Unix(0, 0),
		InvalidationMeta: InvalidationMeta{Status: "Complete"},
	}, nil
}

func checkErrors(distributionName, invalidationID string) error {
	if distributionName == "notfound" {
		return NewInvalidationError(ResourceNotFoundErrorCode, fmt.Errorf("distribution %s not found", distributionName), distributionName)
	}

	if distributionName == "unauthorized distribution" {
		return NewInvalidationError(InvalidationUnauthorizedErrorCode, fmt.Errorf("distribution unauthorized"), distributionName)
	}

	if invalidationID == "notfound" {
		err := fmt.Errorf("mocking AWS Error cloudfront.ErrCodeNoSuchInvalidation")
		return NewInvalidationError(ResourceNotFoundErrorCode, err, err)
	}

	return nil
}
