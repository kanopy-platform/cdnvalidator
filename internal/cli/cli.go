package cli

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kanopy-platform/cdnvalidator/internal/config"
	"github.com/kanopy-platform/cdnvalidator/internal/server"
	"github.com/kanopy-platform/cdnvalidator/pkg/aws/cloudfront"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RootCommand struct{}

func NewRootCommand() *cobra.Command {
	root := &RootCommand{}

	cmd := &cobra.Command{
		Use:               "cdnvalidator",
		PersistentPreRunE: root.persistentPreRunE,
		RunE:              root.runE,
	}

	cmd.PersistentFlags().String("log-level", "info", "Configure log level")
	cmd.PersistentFlags().String("listen-address", ":8080", "Server listen address")
	cmd.PersistentFlags().String("auth-cookie", "auth_token", "Auth cookie name")
	cmd.PersistentFlags().String("auth-header", "", "Header name for the auth token, takes precedence over auth-cookie when set.")
	cmd.PersistentFlags().String("config-file", "", "Configuration file name")
	cmd.PersistentFlags().String("aws-region", "us-east-1", "AWS region for Cloudfront")
	cmd.PersistentFlags().String("aws-key", "", "AWS static credential key for Cloudfront")
	cmd.PersistentFlags().String("aws-secret", "", "AWS static credential secret for Cloudfront")
	cmd.PersistentFlags().String("timeout", "30s", "Timeout")

	return cmd
}

func (c *RootCommand) persistentPreRunE(cmd *cobra.Command, args []string) error {
	// bind flags to viper
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// set log level
	logLevel, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		return err
	}

	log.SetLevel(logLevel)

	return nil
}

func (c *RootCommand) runE(cmd *cobra.Command, args []string) error {
	addr := viper.GetString("listen-address")

	log.Printf("Starting server on %s\n", addr)

	// build config
	config := config.New()
	if viper.GetString("config-file") == "" {
		return errors.New("no config file specified")
	}
	if err := config.Watch(viper.GetString("config-file")); err != nil {
		return err
	}

	// build cloudfront client
	cloudfrontClient, err := cloudfront.New(
		cloudfront.WithAWSRegion(viper.GetString("aws-region")),
		cloudfront.WithStaticCredentials(viper.GetString("aws-key"), viper.GetString("aws-secret")),
		cloudfront.WithTimeout(viper.GetDuration("timeout")),
	)
	if err != nil {
		return err
	}

	s, err := server.New(
		config,
		cloudfrontClient,
		server.WithAuthCookieName(viper.GetString("auth-cookie")),
		server.WithAuthHeaderName(viper.GetString("auth-header")),
	)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Infof("listen: %s", err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Server started")

	<-done

	log.Info("Server stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server Shutdown failed: %+v", err)
		return err
	}

	log.Info("Server shutdown")

	return nil
}
