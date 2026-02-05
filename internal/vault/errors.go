// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package vault deals with authentication and authorization.
package vault

import (
	"fmt"
)

var (
	ErrInvalidProvider                  = fmt.Errorf("invalid provider")
	ErrInvalidIDToken                   = fmt.Errorf("invalid id_token")
	ErrFailedToParseClaims              = fmt.Errorf("failed to parse claims")
	ErrEmailNotVerified                 = fmt.Errorf("email not verified")
	ErrTokenExchangeFailed              = func(err error) error { return fmt.Errorf("token exchange failed: %v", err) }
	ErrMissingIDToken                   = fmt.Errorf("no id_token field in oauth2 token")
	ErrFailedToLoadSymmetricKey         = fmt.Errorf("failed to load a symmetric key")
	ErrFailedToLoadAsymmetricKey        = fmt.Errorf("failed to load an asymmetric key")
	ErrFailedToCreateGoogleOIDCProvider = func(err error) error { return fmt.Errorf("failed to create Google OIDC provider: %w", err) }
	ErrFailedToCreateAppleOIDCProvider  = func(err error) error { return fmt.Errorf("failed to create Apple OIDC provider: %w", err) }
)
