// Package app provides a high-level interface to app modules
package app

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/akvachan/hirevec-backend/internal/logger"
	"github.com/akvachan/hirevec-backend/internal/server"
	"github.com/akvachan/hirevec-backend/internal/store"
	"github.com/akvachan/hirevec-backend/internal/utils"
	"github.com/akvachan/hirevec-backend/internal/vault"
)

type AppConfig struct {
	Host               string
	ReadTimeout        string
	WriteTimeout       string
	GracePeriod        string
	DBConnString       string
	LogLevel           string
	SymmetricKeyHex    string
	AsymmetricKeyHex   string
	GoogleClientID     string
	GoogleClientSecret string
	AppleClientID      string
	AppleClientSecret  string
}

func Run(c AppConfig) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	loggerConfig := logger.LoggerConfig{
		Level: utils.ParseLogLevelWithDefault(c.LogLevel, logger.DefaultLogLevel),
	}
	logger.Init(loggerConfig)

	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	vaultConfig := vault.VaultConfig{
		Host:               c.Host,
		GoogleClientID:     c.GoogleClientID,
		GoogleClientSecret: c.GoogleClientSecret,
		AppleClientID:      c.AppleClientID,
		AppleClientSecret:  c.AppleClientSecret,
		SymmetricKeyHex:    c.SymmetricKeyHex,
		AsymmetricKeyHex:   c.AsymmetricKeyHex,
	}
	v, err := vault.NewVault(ctx, vaultConfig)
	if err != nil {
		return fmt.Errorf("vault init failed: %w", err)
	}

	storeConfig := store.StoreConfig{
		DBConnString: c.DBConnString,
	}
	s, err := store.NewStore(storeConfig)
	if err != nil {
		return fmt.Errorf("store init failed: %w", err)
	}

	serverConfig := server.ServerConfig{
		Host:         c.Host,
		ReadTimeout:  utils.ParseTimeoutWithDefault(c.ReadTimeout, server.DefaultReadTimeout),
		WriteTimeout: utils.ParseTimeoutWithDefault(c.WriteTimeout, server.DefaultWriteTimeout),
		GracePeriod:  utils.ParseTimeoutWithDefault(c.GracePeriod, server.DefaultGracePeriod),
	}

	return server.Run(ctx, serverConfig, s, v)
}

func (c *AppConfig) Validate() error {
	var errs []error

	if c.DBConnString == "" {
		errs = append(errs, fmt.Errorf("database connection string (DEV_DB_URL) is required"))
	}
	if c.Host == "" {
		errs = append(errs, fmt.Errorf("host (HOST) is required"))
	}
	c.Host = strings.TrimSuffix(c.Host, "/")

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
