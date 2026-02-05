// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package vault deals with authentication and authorization.
package vault

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"aidanwoods.dev/go-paseto"
)

var stateStore = &StateStore{
	states: make(map[string]time.Time),
}

type IssuedTokenType string

const (
	refreshToken           IssuedTokenType = "urn:ietf:params:oauth:token-type:refresh_token"
	accessToken            IssuedTokenType = "urn:ietf:params:oauth:token-type:access_token"
	RefreshTokenExpiration                 = 30 * 24 * time.Hour
	AccessTokenExpiration                  = 30 * time.Minute
)

type StateStore struct {
	mu     sync.RWMutex
	states map[string]time.Time
}

// PasetoKey defines the public key structure within PublicPasetoKeys.
type PasetoKey struct {
	Version uint8  `json:"version"`
	Kid     uint32 `json:"kid"`
	Key     []byte `json:"key"`
}

// PublicPasetoKeys defines the API response from the endpoint that serves public keys.
type PublicPasetoKeys struct {
	Keys []PasetoKey `json:"keys"`
}

type RefreshTokenClaims struct {
	UserID   uint32
	Provider string
	JTI      string
}

type AccessTokenClaims struct {
	UserID   uint32
	Provider string
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    uint32 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (v vault) ParseAccessToken(tokenString string) (*AccessTokenClaims, error) {
	parsedToken, err := v.AccessTokenParser.ParseV4Public(v.V4AsymetricPublicKey, tokenString, nil)
	if err != nil {
		return nil, errors.New("invalid access token")
	}

	userID, err := parsedToken.GetSubject()
	if err != nil {
		return nil, errors.New("invalid subject")
	}
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, errors.New("could not parse user ID")
	}

	provider, err := parsedToken.GetString("provider")
	if err != nil {
		return nil, errors.New("could not parse provider")
	}
	if provider != "apple" && provider != "google" {
		return nil, errors.New("invalid provider")
	}

	return &AccessTokenClaims{
		UserID:   uint32(id),
		Provider: provider,
	}, nil
}

func (v vault) ParseRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	parsedToken, err := v.RefreshTokenParser.ParseV4Local(v.V4SymmetricKey, tokenString, nil)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	userID, err := parsedToken.GetSubject()
	if err != nil || userID == "" {
		return nil, errors.New("invalid subject")
	}
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, errors.New("could not parse user ID")
	}

	provider, err := parsedToken.GetString("provider")
	if err != nil {
		return nil, errors.New("could not parse provider")
	}
	if provider != "apple" && provider != "google" {
		return nil, errors.New("invalid provider")
	}

	tokenType, err := parsedToken.GetString("type")
	if err != nil {
		return nil, errors.New("could not parse type")
	}
	if tokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	jti, err := parsedToken.GetJti()
	if err != nil {
		return nil, errors.New("could not parse jti")
	}
	if jti == "" {
		return nil, errors.New("invalid refresh token")
	}

	return &RefreshTokenClaims{
		UserID:   uint32(id),
		Provider: provider,
		JTI:      jti,
	}, nil
}

func (v vault) GetPublicKey() []byte {
	return v.V4AsymetricPublicKey.ExportBytes()
}

func (v vault) CreateAccessToken(userID uint32, provider string, scope string) (string, error) {
	now := time.Now().UTC()

	token := paseto.NewToken()
	token.SetAudience("hirevec-api")
	token.SetIssuer("hirevec")
	token.SetSubject(fmt.Sprintf("%d", userID))
	token.SetExpiration(now.Add(AccessTokenExpiration))
	token.SetNotBefore(now)
	token.SetIssuedAt(now)

	if err := token.Set("token_type", accessToken); err != nil {
		return "", errors.New("could not set token type")
	}

	if err := token.Set("provider", provider); err != nil {
		return "", errors.New("could not set provider")
	}

	token.SetString("scope", scope)

	return token.V4Sign(v.V4AsymmetricSecretKey, nil), nil
}

func (v vault) CreateRefreshToken(userID uint32, provider string, jti string) (string, error) {
	now := time.Now().UTC()

	token := paseto.NewToken()
	token.SetAudience("hirevec-api")
	token.SetIssuer("hirevec")
	token.SetSubject(fmt.Sprintf("%d", userID))
	token.SetExpiration(now.Add(RefreshTokenExpiration))
	token.SetNotBefore(now)
	token.SetIssuedAt(now)
	token.SetJti(jti)

	if err := token.Set("token_type", refreshToken); err != nil {
		return "", errors.New("could not set token type")
	}

	if err := token.Set("provider", provider); err != nil {
		return "", errors.New("could not set provider")
	}

	return token.V4Encrypt(v.V4SymmetricKey, nil), nil
}

func (v vault) CreateTokenPair(userID uint32, provider string, jti string, scope string) (TokenPair, error) {
	accessToken, err := v.CreateAccessToken(userID, provider, scope)
	if err != nil {
		return TokenPair{}, errors.New("could not create an access token")
	}

	refreshToken, err := v.CreateRefreshToken(userID, provider, jti)
	if err != nil {
		return TokenPair{}, errors.New("could not create a refresh token")
	}

	return TokenPair{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    uint32(AccessTokenExpiration.Abs().Seconds()),
		RefreshToken: refreshToken,
		Scope:        scope,
	}, nil
}
