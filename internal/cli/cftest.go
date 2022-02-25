package cli

import (
	"context"
	"os"
	"time"

	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type cfTestCommand struct {
	*cobra.Command
	distributionId string
	paths          []string
	region         string
	credKey        string
	credSecret     string
	timeout        uint
}

func newCfTestCommand() *cobra.Command {
	cmd := cfTestCommand{Command: &cobra.Command{}}

	cmd.Use = "cftest"
	cmd.Short = "Creates an invalidation against the specified distribution_id and paths"
	cmd.RunE = cmd.testRunE

	cmd.Flags().StringVarP(&cmd.distributionId, "distribution", "d", "", "Cloudfront distribution ID")
	if err := cmd.MarkFlagRequired("distribution"); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	cmd.Flags().StringSliceVarP(&cmd.paths, "paths", "p", []string{}, "Comma separated paths to invalidate")
	if err := cmd.MarkFlagRequired("paths"); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	cmd.Flags().StringVarP(&cmd.region, "region", "r", "", "AWS Region")
	cmd.Flags().StringVarP(&cmd.credKey, "key", "k", "", "AWS static credentials Key")
	cmd.Flags().StringVarP(&cmd.credSecret, "secret", "s", "", "AWS static credentials Secret")
	cmd.Flags().UintVarP(&cmd.timeout, "timeout", "t", 30, "Timeout for cloudfront commands (seconds)")

	return cmd.Command
}

func (cmd *cfTestCommand) testRunE(command *cobra.Command, args []string) error {
	opts := []cloudfront.Option{
		cloudfront.WithTimeout(time.Duration(cmd.timeout) * time.Second),
	}

	if cmd.region != "" {
		opts = append(opts, cloudfront.WithAwsRegion(cmd.region))
	}

	if cmd.credKey != "" && cmd.credSecret != "" {
		opts = append(opts, cloudfront.WithStaticCredentials(cmd.credKey, cmd.credSecret))
	}

	c, err := cloudfront.New(opts...)
	if err != nil {
		log.Errorf("cloudfront.New error: %v", err)
		return err
	}

	log.Info("Creating Invalidation...")

	create, err := c.CreateInvalidation(context.Background(), cmd.distributionId, cmd.paths)
	if err != nil {
		log.Errorf("CreateInvalidation error: %v", err)
		return err
	}
	log.Infof("Created Invalidation: Id=%v, Status=%v", create.InvalidationId, create.Status)

	get, err := c.GetInvalidation(context.Background(), cmd.distributionId, create.InvalidationId)
	if err != nil {
		log.Errorf("GetInvalidation error: %v", err)
		return err
	}
	log.Infof("Got Invalidation %v: CreateTime=%v, Status=%v, Paths=%v", create.InvalidationId, get.CreateTime, get.Status, get.Paths)

	return nil
}
