package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kanopy-platform/cdnvalidator/internal/config"
	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
)

type DistributionService struct {
	Config     *config.Config
	Cloudfront *cloudfront.Client
}

func New(config *config.Config, cloudfront *cloudfront.Client) *DistributionService {
	return &DistributionService{
		Config:     config,
		Cloudfront: cloudfront,
	}
}

func (d *DistributionService) getDistribution(ctx context.Context, distributionName string) (*config.Distribution, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	distribution := d.Config.Distribution(distributionName)
	if distribution == nil {
		return nil, NewInvalidationError(ResourceNotFoundErrorCode, fmt.Errorf("distribution %s not found", distributionName), distributionName)
	}

	// check user is entitled to the distributionName
	entitledDistributions := d.Config.DistributionsFromClaims(claims)
	if _, ok := entitledDistributions[distributionName]; !ok {
		return nil, NewInvalidationError(InvalidationUnauthorizedErrorCode, fmt.Errorf("distribution unauthorized"), distributionName)
	}

	return distribution, nil
}

func (d *DistributionService) List(ctx context.Context) ([]string, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	distributions := d.Config.DistributionsFromClaims(claims)

	ret := make([]string, 0, len(distributions))
	for name := range distributions {
		ret = append(ret, name)
	}

	return ret, nil
}

func (d *DistributionService) CreateInvalidation(ctx context.Context, distributionName string, paths []string) (*InvalidationResponse, error) {
	distribution, err := d.getDistribution(ctx, distributionName)
	if err != nil {
		return nil, err
	}

	cleanedPaths := make([]string, 0, len(paths))
	invalidPaths := make([]string, 0)

	for _, p := range paths {
		cleanedPath := filepath.Clean(p)

		if strings.HasPrefix(cleanedPath, distribution.Prefix) {
			cleanedPaths = append(cleanedPaths, cleanedPath)
		} else {
			invalidPaths = append(invalidPaths, p)
		}
	}
	if len(invalidPaths) > 0 {
		return nil, NewInvalidationError(BadRequestErrorCode, errors.New("unauthorized paths"), fmt.Sprintf("unauthorized paths: %v", invalidPaths))
	}

	res, err := d.Cloudfront.CreateInvalidation(ctx, distribution.ID, cleanedPaths)
	if err != nil {
		return nil, NewInvalidationError(BadRequestErrorCode, errors.New("cloudfront CreateInvalidation failed"), err)
	}

	return &InvalidationResponse{
		InvalidationMeta: InvalidationMeta{
			Status: res.Status,
		},
		ID:      res.InvalidationID,
		Created: res.CreateTime,
		Paths:   res.Paths,
	}, nil
}

func (d *DistributionService) GetInvalidationStatus(ctx context.Context, distributionName string, invalidationID string) (*InvalidationResponse, error) {
	distribution, err := d.getDistribution(ctx, distributionName)
	if err != nil {
		return nil, err
	}

	res, err := d.Cloudfront.GetInvalidation(ctx, distribution.ID, invalidationID)
	if err != nil {
		return nil, NewInvalidationError(BadRequestErrorCode, fmt.Errorf("cloudfront GetInvalidation failed"), err)
	}

	return &InvalidationResponse{
		InvalidationMeta: InvalidationMeta{
			Status: res.Status,
		},
		ID:      res.InvalidationID,
		Created: res.CreateTime,
		Paths:   res.Paths,
	}, nil
}
