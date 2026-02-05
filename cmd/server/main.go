// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package main

import (
	"log/slog"
	"os"

	"github.com/akvachan/hirevec-backend/internal"
	"github.com/akvachan/hirevec-backend/internal/utils"
)

func loadConfig() app.AppConfig {
	return app.AppConfig{
		Host:         os.Getenv("HOST"),
		ReadTimeout:  os.Getenv("REQUEST_READ_TIMEOUT"),
		WriteTimeout: os.Getenv("REQUEST_WRITE_TIMEOUT"),
		GracePeriod:  os.Getenv("GRACE_PERIOD"),
		DBConnString: os.Getenv("DEV_DB_URL"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
		SymmetricKeyHex: os.Getenv("SYMMETRIC_KEY"),
		AsymmetricKeyHex: os.Getenv("ASYMMETRIC_KEY"),
	}
}

func main() {
	err := utils.Loadenv(".dev.env")
	if err != nil {
		slog.Warn(
			"could not load .env, using system environment",
			"err", err,
		)
	}
	if err := app.Run(loadConfig()); err != nil {
		slog.Error(
			"app crashed",
			"err", err,
		)
		os.Exit(1)
	}
}
