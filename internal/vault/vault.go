// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package vault deals with authentication and authorization.
package vault

import (
	"context"
	"fmt"
	"net/http"

	"aidanwoods.dev/go-paseto"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/akvachan/hirevec-backend/internal/store/db/models"
)

type Vault interface {
	ParseAccessToken(token string) (*AccessTokenClaims, error)
	ParseRefreshToken(token string) (*RefreshTokenClaims, error)
	GetPublicKey() []byte
	CreateAccessToken(userID uint32, provider string, scope string) (string, error)
	CreateRefreshToken(userID uint32, provider string, jti string) (string, error)
	CreateTokenPair(userID uint32, provider string, jti string, scope string) (TokenPair, error)
	CreateStateToken() (string, error)
	CreateAuthCodeURL(state string, verifier string, provider string) (string, error)
	ValidateAndDeleteStateToken(state string) bool
	CleanupExpiredStateTokens()
	ExchangeGoogleCodeForIDToken(ctx context.Context, code string, verifier *http.Cookie) (string, error)
	ExchangeAppleCodeForIDToken(ctx context.Context, code string, verifier *http.Cookie) (string, error)
	VerifyAndParseGoogleIDToken(ctx context.Context, idToken string) (*models.User, error)
	VerifyAndParseAppleIDToken(ctx context.Context, idToken string, userJSON string) (*models.User, error)
}

type vault struct {
	AccessTokenParser     paseto.Parser
	RefreshTokenParser    paseto.Parser
	V4AsymetricPublicKey  paseto.V4AsymmetricPublicKey
	V4AsymmetricSecretKey paseto.V4AsymmetricSecretKey
	V4SymmetricKey        paseto.V4SymmetricKey
	GoogleOIDCConfig      OIDCConfig
	AppleOIDCConfig       OIDCConfig
}

type VaultConfig struct {
	Host               string
	GoogleClientID     string
	GoogleClientSecret string
	AppleClientID      string
	AppleClientSecret  string
	SymmetricKeyHex    string
	AsymmetricKeyHex   string
}

func NewVault(ctx context.Context, c VaultConfig) (*vault, error) {
	var accessTokenParser paseto.Parser
	var refreshTokenParser paseto.Parser

	symmetricKey, err := paseto.V4SymmetricKeyFromHex(c.SymmetricKeyHex)
	if err != nil {
		return nil, ErrFailedToLoadSymmetricKey
	}

	asymmetricKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(c.AsymmetricKeyHex)
	if err != nil {
		return nil, ErrFailedToLoadAsymmetricKey
	}

	accessTokenParser = paseto.NewParser()
	accessTokenParser.AddRule(paseto.ForAudience("hirevec-api"))
	accessTokenParser.AddRule(paseto.IssuedBy("hirevec"))
	accessTokenParser.AddRule(paseto.NotExpired())
	accessTokenParser.AddRule(paseto.NotBeforeNbf())

	refreshTokenParser = paseto.NewParser()
	refreshTokenParser.AddRule(paseto.ForAudience("hirevec-api"))
	refreshTokenParser.AddRule(paseto.IssuedBy("hirevec"))
	refreshTokenParser.AddRule(paseto.NotExpired())
	refreshTokenParser.AddRule(paseto.NotBeforeNbf())

	googleProvider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, ErrFailedToCreateGoogleOIDCProvider(err)
	}

	googleOIDC := OIDCConfig{
		OAuth2Config: &oauth2.Config{
			ClientID:     c.GoogleClientID,
			ClientSecret: c.GoogleClientSecret,
			RedirectURL:  fmt.Sprintf("%s/oauth2/callback/google", c.Host),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			Endpoint:     googleProvider.Endpoint(),
		},
		Verifier: googleProvider.Verifier(&oidc.Config{ClientID: c.GoogleClientID}),
	}

	appleProvider, err := oidc.NewProvider(ctx, "https://appleid.apple.com")
	if err != nil {
		return nil, ErrFailedToCreateAppleOIDCProvider(err)
	}

	appleOIDC := OIDCConfig{
		OAuth2Config: &oauth2.Config{
			ClientID:     c.AppleClientID,
			ClientSecret: c.AppleClientSecret,
			RedirectURL:  fmt.Sprintf("%s/oauth2/callback/apple", c.Host),
			Scopes:       []string{oidc.ScopeOpenID, "name", "email"},
			Endpoint:     appleProvider.Endpoint(),
		},
		Verifier: appleProvider.Verifier(&oidc.Config{ClientID: c.AppleClientID}),
	}

	return &vault{
		AccessTokenParser:     accessTokenParser,
		RefreshTokenParser:    refreshTokenParser,
		V4AsymmetricSecretKey: asymmetricKey,
		V4AsymetricPublicKey:  asymmetricKey.Public(),
		V4SymmetricKey:        symmetricKey,
		GoogleOIDCConfig:      googleOIDC,
		AppleOIDCConfig:       appleOIDC,
	}, nil
}
