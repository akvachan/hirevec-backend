// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"log/slog"
	"os"
)

type LoggerConfig struct {
	Level slog.Level
}

const DefaultLogLevel = slog.LevelError

func InitLogger(config LoggerConfig) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: config.Level}))
	slog.SetDefault(logger)
}
