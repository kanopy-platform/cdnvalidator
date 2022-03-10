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
	config     *config.Config
	cloudfront *cloudfront.Client
}

func New(opts ...Option) (*DistributionService, error) {
	var err error

	defaultCloudfront, err := cloudfront.New()
	if err != nil {
		return nil, err
	}

	d := &DistributionService{
		config:     config.New(),
		cloudfront: defaultCloudfront,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d, nil
}

func (d *DistributionService) getDistribution(ctx context.Context, distributionName string) (*config.Distribution, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	distribution := d.config.Distribution(distributionName)
	if distribution == nil {
		return nil, NewInvalidationError(ResourceNotFoundErrorCode, fmt.Errorf("distribution %s not found", distributionName), distributionName)
	}

	// check user is entitled to the distributionName
	entitledDistributions := d.config.DistributionsFromClaims(claims)
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

	distributions := d.config.DistributionsFromClaims(claims)

	ret := make([]string, 0, len(distributions))
	for name := range distributions {
		ret = append(ret, name)
	}

	return ret, nil
}

func (d *DistributionService) CreateInvalidation(ctx context.Context, distributionName string, paths []string) (*InvalidationResponse, error) {
	if len(paths) == 0 {
		return nil, NewInvalidationError(InternalServerError, fmt.Errorf("invalid path"), errors.New("must provide at least one path"))
	}

	distribution, err := d.getDistribution(ctx, distributionName)
	if err != nil {
		return nil, err
	}

	// prepend prefix to all the paths
	absolutePaths := make([]string, 0, len(paths))
	for _, p := range paths {
		// prevent users from doing funny business by going up a directory
		if strings.Contains(p, "../") {
			return nil, NewInvalidationError(InternalServerError, fmt.Errorf("invalid path"), errors.New("path cannot contain ../"))
		}

		absolutePaths = append(absolutePaths, path.Join(distribution.Prefix, p))
	}

	res, err := d.cloudfront.CreateInvalidation(ctx, distribution.ID, absolutePaths)
	if err != nil {
		return nil, NewInvalidationError(InternalServerError, fmt.Errorf("cloudfront CreateInvalidation failed"), err)
	}

	return &InvalidationResponse{
		InvalidationMeta: InvalidationMeta{
			Status: res.Status,
		},
		ID:      res.InvalidationId,
		Created: res.CreateTime,
		Paths:   res.Paths,
	}, nil
}

func (d *DistributionService) GetInvalidationStatus(ctx context.Context, distributionName string, invalidationID string) (*InvalidationResponse, error) {
	distribution, err := d.getDistribution(ctx, distributionName)
	if err != nil {
		return nil, err
	}

	res, err := d.cloudfront.GetInvalidation(ctx, distribution.ID, invalidationID)
	if err != nil {
		return nil, NewInvalidationError(InternalServerError, fmt.Errorf("cloudfront GetInvalidation failed"), err)
	}

	return &InvalidationResponse{
		InvalidationMeta: InvalidationMeta{
			Status: res.Status,
		},
		ID:      res.InvalidationId,
		Created: res.CreateTime,
		Paths:   res.Paths,
	}, nil
}
