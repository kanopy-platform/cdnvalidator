package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kanopy-platform/cdnvalidator/internal/core"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
	"github.com/spf13/viper"
	"honnef.co/go/tools/config" // TODO replace with Ricardo's
)

type DistributionService struct {
	config     *config.Config
	configFile string
	cloudfront *cloudfront.Client
	awsRegion  string
	awsKey     string
	awsSecret  string
	awsTimeout time.Duration
}

func New(opts ...Option) (*DistributionService, error) {
	var err error

	d := &DistributionService{
		config:     config.New(),
		configFile: viper.GetString("config-file"),
		awsRegion:  viper.GetString("aws-region"),
		awsKey:     viper.GetString("aws-key"),
		awsSecret:  viper.GetString("aws-secret"),
		awsTimeout: viper.GetDuration("aws-timeout"),
	}

	for _, opt := range opts {
		opt(d)
	}

	// set up config
	if err := d.config.Watch(d.configFile); err != nil {
		return nil, err
	}

	// set up AWS Cloudfront client
	cfOpts := []cloudfront.Option{
		cloudfront.WithAwsRegion(d.awsRegion),
		cloudfront.WithStaticCredentials(d.awsKey, d.awsSecret),
		cloudfront.WithTimeout(d.awsTimeout),
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

	// check user is entitled to the distributionName
	entitlements := d.config.DistributionsFromClaims(claims)
	distribution := d.config.Distribution(distributionName)

	if _, ok := entitlements[distributionName]; !ok {
		if distribution != nil {
			// distribution exists but user is not entitled to it
			return nil, NewInvalidationError(InvalidationUnauthorizedErrorCode, fmt.Errorf("distribution unauthorized"), distributionName)
		} else {
			return nil, NewInvalidationError(ResourceNotFoundErrorCode, fmt.Errorf("distribution %s not found", distributionName), distributionName)
		}
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
	for _, name := range distributions {
		distribution := d.config.Distribution(name)
		ret[name] = Distribution{distribution.ID, distribution.Prefix}
	}

	return ret, nil
}

func (d *DistributionService) CreateInvalidation(ctx context.Context, distributionName string, paths []string) (*InvalidationResponse, error) {
	distribution, err := d.getDistribution(ctx, distributionName)
	if err != nil {
		return nil, err
	}

	// check paths are valid
	invalidPaths := make([]string, 0)

	for _, path := range paths {
		if !strings.HasPrefix(path, distribution.Prefix) {
			invalidPaths = append(invalidPaths, path)
		}
	}
	if len(invalidPaths) > 0 {
		return nil, NewInvalidationError(InvalidationUnauthorizedErrorCode, fmt.Errorf("paths unauthorized"), invalidPaths)
	}

	res, err := d.cloudfront.CreateInvalidation(ctx, distribution.ID, paths)
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
