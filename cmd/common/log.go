// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package common implements common helper functions for the scripts
package common

import (
	"log/slog"
	"os"
)

func Exit(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
