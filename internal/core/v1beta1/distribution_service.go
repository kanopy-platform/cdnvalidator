package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/kanopy-platform/cdnvalidator/internal/config"
	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
)

type DistributionService struct {
	config     *config.Config
	configFile string
	cloudfront *cloudfront.Client
	awsRegion  string
	awsKey     string
	awsSecret  string
	timeout    time.Duration
}

func New(opts ...Option) (*DistributionService, error) {
	var err error

	d := &DistributionService{
		config:  config.New(),
		timeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(d)
	}

	// set up config
	if d.configFile != "" {
		if err := d.config.Watch(d.configFile); err != nil {
			return nil, err
		}
	}

	// set up AWS Cloudfront client
	cfOpts := []cloudfront.Option{
		cloudfront.WithAwsRegion(d.awsRegion),
		cloudfront.WithStaticCredentials(d.awsKey, d.awsSecret),
		cloudfront.WithTimeout(d.timeout),
	}

	d.cloudfront, err = cloudfront.New(cfOpts...)
	if err != nil {
		return nil, err
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

func (d *DistributionService) List(ctx context.Context) (map[VanityDistributionName]Distribution, error) {
	claims := core.GetClaims(ctx)
	if len(claims) == 0 {
		return nil, errors.New("no claims present")
	}

	ret := make(map[VanityDistributionName]Distribution)

	distributions := d.config.DistributionsFromClaims(claims)
	for name := range distributions {
		distribution := d.config.Distribution(name)
		ret[VanityDistributionName(name)] = Distribution{distribution.ID, distribution.Prefix}
	}

	return ret, nil
}

func (d *DistributionService) CreateInvalidation(ctx context.Context, distributionName string, paths []string) (*InvalidationResponse, error) {
	distribution, err := d.getDistribution(ctx, distributionName)
	if err != nil {
		return nil, err
	}

	// prepend prefix to all the paths
	distributionPaths := make([]string, 0)
	for _, p := range paths {
		// prevent users from doing funny business by going up a directory
		if strings.Contains(p, "../") {
			return nil, NewInvalidationError(InternalServerError, fmt.Errorf("invalid path"), errors.New("path cannot contain ../"))
		}

		distributionPaths = append(distributionPaths, path.Join(distribution.Prefix, p))
	}

	res, err := d.cloudfront.CreateInvalidation(ctx, distribution.ID, distributionPaths)
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
