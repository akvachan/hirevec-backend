package utils

import (
	"log/slog"
	"strconv"
	"time"
)

func ParseTimeoutWithDefault(value string, defaultValue time.Duration) time.Duration {
	parsedReadTimeout, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		slog.Warn("failed to parse timeout, using default", "value", value, "default", defaultValue)
		return defaultValue
	}
	return time.Duration(parsedReadTimeout) * time.Millisecond
}

func ParseLogLevelWithDefault(value string, defaultValue slog.Level) slog.Level {
	switch value {
	case "INFO":
		return slog.LevelInfo
	case "ERROR":
		return slog.LevelError
	case "WARN":
		return slog.LevelWarn
	default:
		slog.Warn("failed to parse log level, using default", "value", value, "default", defaultValue)
		return defaultValue
	}
}
