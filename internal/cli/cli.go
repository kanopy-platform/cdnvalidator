package cli

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kanopy-platform/cdnvalidator/internal/server"
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
	cmd.PersistentFlags().String("auth-cookie", "", "Auth cookie name")
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

	opts := []server.Option{
		server.WithAuthCookieName(viper.GetString("auth-cookie")),
		server.WithConfigFile(viper.GetString("config-file")),
		server.WithAwsRegion(viper.GetString("aws-region")),
		server.WithAwsStaticCredentials(viper.GetString("aws-key"), viper.GetString("aws-secret")),
		server.WithTimeout(viper.GetDuration("timeout")),
	}

	s, err := server.New(opts...)
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
