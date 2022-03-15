package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"path"
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

	// prepend prefix to all the paths
	absolutePaths := make([]string, 0, len(paths))
	for _, p := range paths {
		// prevent users from doing funny business by going up a directory
		if strings.Contains(p, "../") {
			return nil, NewInvalidationError(BadRequestErrorCode, fmt.Errorf("invalid path"), errors.New("path cannot contain ../"))
		}

		absolutePaths = append(absolutePaths, path.Join(distribution.Prefix, p))
	}

	res, err := d.Cloudfront.CreateInvalidation(ctx, distribution.ID, absolutePaths)
	if err != nil {
		return nil, NewInvalidationError(BadRequestErrorCode, fmt.Errorf("cloudfront CreateInvalidation failed"), err)
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
