// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package server implements the HTTP transport layer, providing RESTful endpoints.
package server

import (
	"fmt"
)

var (
	ErrExtraDataDecoded       = fmt.Errorf("extra data decoded")
	ErrFailedToBindAddress    = func(host string, err error) error { return fmt.Errorf("failed to bind to %s: %w", host, err) }
	ErrFailedToDecode         = fmt.Errorf("could not decode")
	ErrFailedToParseLimit     = fmt.Errorf("limit must be zero or a positive integer")
	ErrFailedToParseOffset    = fmt.Errorf("offset must be zero or a positive integer")
	ErrFailedToParseSerialID  = fmt.Errorf("id must be an integer")
	ErrFailedToShutdownServer = func(err error) error { return fmt.Errorf("failed to shutdown server: %w", err) }
	ErrNameHasForbiddenChars  = fmt.Errorf("name contains forbidden characters")
	ErrNameTooLong            = fmt.Errorf("name length must be smaller than 128 characters")
	ErrNameTooShort           = fmt.Errorf("name length must be bigger than 1 character")
	ErrNotPositiveSerialID    = fmt.Errorf("id must be a positive integer")
)
