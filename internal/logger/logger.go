// Package logger is a package for configuring various loggers and observability tools
package logger

import (
	"log/slog"
	"os"
)

const DefaultLogLevel = slog.LevelError

type LoggerConfig struct {
	Level slog.Level
}

func Init(config LoggerConfig) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: config.Level}))
	slog.SetDefault(logger)
}
