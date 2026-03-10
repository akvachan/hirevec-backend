// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"strings"
	"sync"
	"time"

	"aidanwoods.dev/go-paseto"
)

type (
	ScopeType string

	ClaimType string

	IssuedTokenType string

	StateStore struct {
		mu     sync.RWMutex
		states map[string]time.Time
	}

	PasetoKey struct {
		Version uint8  `json:"version"`
		Kid     uint32 `json:"kid"`
		Key     []byte `json:"key"`
	}

	PublicPasetoKeys struct {
		Keys []PasetoKey `json:"keys"`
	}

	RefreshTokenClaims struct {
		UserID   string
		Provider string
		JTI      string
	}

	AccessTokenClaims struct {
		UserID   string
		Provider string
		Scope    ScopeType
	}

	AccessToken struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   uint32 `json:"expires_in"`
		Scope       string `json:"scope"`
		UserID      string `json:"user_id"`
	}

	RefreshToken struct {
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    uint32 `json:"expires_in"`
		UserID       string `json:"user_id"`
	}

	TokenPair struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    uint32 `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		UserID       string `json:"user_id"`
	}
)

var stateStore = &StateStore{
	states: make(map[string]time.Time),
}

const (
	IssuedTokenTypeRefreshToken   IssuedTokenType = "urn:ietf:params:oauth:token-type:refresh_token"
	IssuedTokenTypeAccessToken    IssuedTokenType = "urn:ietf:params:oauth:token-type:access_token"
	DefaultRefreshTokenExpiration                 = 30 * 24 * time.Hour
	DefaultAccessTokenExpiration                  = 30 * time.Minute
	ScopeTypeCandidate                            = "role:candidate"
	ScopeTypeRecruiter                            = "role:recruiter"
	ScopeTypeAdmin                                = "role:admin"
	ScopeTypeOnboarding                           = "role:onboarding"
	TokenAudience                                 = "api.hirevec.com"
	TokenIssuer                                   = "api.hirevec.com"
)

func NewScope(scope string) (ScopeType, error) {
	switch scope {
	case ScopeTypeAdmin, ScopeTypeOnboarding, ScopeTypeCandidate, ScopeTypeRecruiter:
		return ScopeType(scope), nil
	default:
		return "", ErrInvalidScopeType
	}
}

func (v PasetoVault) ParseAccessToken(tokenString string) (*AccessTokenClaims, error) {
	parsedToken, err := v.AccessTokenParser.ParseV4Public(v.V4AsymetricPublicKey, tokenString, nil)
	if err != nil {
		return nil, ErrInvalidAccessToken
	}

	userID, err := parsedToken.GetSubject()
	if err != nil {
		return nil, ErrInvalidSubject
	}

	provider, err := parsedToken.GetString("provider")
	if err != nil {
		return nil, ErrFailedParseProvider
	}
	if provider != "apple" && provider != "google" {
		return nil, ErrInvalidProvider
	}

	scope, err := parsedToken.GetString("scope")
	if err != nil {
		return nil, ErrFailedParseScope
	}

	validScope, err := NewScope(scope)
	if err != nil {
		return nil, ErrInvalidScopeType
	}

	return &AccessTokenClaims{
		UserID:   userID,
		Provider: provider,
		Scope:    validScope,
	}, nil
}

func (v PasetoVault) ParseRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	parsedToken, err := v.RefreshTokenParser.ParseV4Local(v.V4SymmetricKey, tokenString, nil)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	userID, err := parsedToken.GetSubject()
	if err != nil || userID == "" {
		return nil, ErrInvalidSubject
	}

	provider, err := parsedToken.GetString("provider")
	if err != nil {
		return nil, ErrFailedParseProvider
	}
	if provider != "apple" && provider != "google" {
		return nil, ErrInvalidProvider
	}

	tokenType, err := parsedToken.GetString("type")
	if err != nil {
		return nil, ErrFailedParseTokenType
	}
	if tokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	jti, err := parsedToken.GetJti()
	if err != nil || jti == "" {
		return nil, ErrFailedParseJTI
	}

	return &RefreshTokenClaims{
		UserID:   userID,
		Provider: provider,
		JTI:      jti,
	}, nil
}

func (v PasetoVault) GetPublicKey() []byte {
	return v.V4AsymetricPublicKey.ExportBytes()
}

func (v PasetoVault) CreateAccessToken(userID string, provider string, scope string) (*AccessToken, error) {
	now := time.Now().UTC()

	var expiration time.Duration
	switch {
	case scope == ScopeTypeOnboarding:
		expiration = 24 * time.Hour
	case v.AccessTokenExpiration != 0:
		expiration = v.AccessTokenExpiration
	default:
		expiration = DefaultAccessTokenExpiration
	}

	token := paseto.NewToken()
	token.SetAudience(TokenAudience)
	token.SetIssuer(TokenIssuer)
	token.SetSubject(userID)
	token.SetExpiration(now.Add(expiration))
	token.SetNotBefore(now)
	token.SetIssuedAt(now)

	if err := token.Set("token_type", IssuedTokenTypeAccessToken); err != nil {
		return nil, ErrFailedSetTokenType
	}

	if err := token.Set("provider", provider); err != nil {
		return nil, ErrFailedSetProvider
	}

	token.SetString("scope", scope)

	return &AccessToken{
		AccessToken: token.V4Sign(v.V4AsymmetricSecretKey, nil),
		TokenType:   "Bearer",
		ExpiresIn:   uint32(expiration.Abs().Seconds()),
		Scope:       scope,
		UserID:      userID,
	}, nil
}

func (v PasetoVault) CreateRefreshToken(userID string, provider string, jti string) (*RefreshToken, error) {
	now := time.Now().UTC()

	token := paseto.NewToken()
	token.SetAudience(TokenAudience)
	token.SetIssuer(TokenIssuer)
	token.SetSubject(userID)
	token.SetExpiration(now.Add(DefaultRefreshTokenExpiration))
	token.SetNotBefore(now)
	token.SetIssuedAt(now)
	token.SetJti(jti)

	if err := token.Set("token_type", IssuedTokenTypeRefreshToken); err != nil {
		return nil, ErrFailedSetTokenType
	}

	if err := token.Set("provider", provider); err != nil {
		return nil, ErrFailedSetProvider
	}

	var expiresIn uint32
	if v.RefreshTokenExpiration != 0 {
		expiresIn = uint32(v.RefreshTokenExpiration.Abs().Seconds())
	} else {
		expiresIn = uint32(DefaultRefreshTokenExpiration.Abs().Seconds())
	}

	return &RefreshToken{
		RefreshToken: token.V4Encrypt(v.V4SymmetricKey, nil),
		ExpiresIn:    expiresIn,
		UserID:       userID,
	}, nil
}

func (v PasetoVault) CreateTokenPair(userID string, provider string, jti string, scope string) (*TokenPair, error) {
	accessToken, err := v.CreateAccessToken(userID, provider, scope)
	if err != nil {
		return nil, ErrFailedCreateAccessToken
	}

	refreshToken, err := v.CreateRefreshToken(userID, provider, jti)
	if err != nil {
		return nil, ErrFailedCreateRefreshToken
	}

	return &TokenPair{
		AccessToken:  accessToken.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    uint32(DefaultAccessTokenExpiration.Abs().Seconds()),
		RefreshToken: refreshToken.RefreshToken,
		Scope:        scope,
		UserID:       userID,
	}, nil
}

func (v PasetoVault) GetScopeForRoles(roles []string) (string, error) {
	scopes := make([]string, 0, len(roles))

	for _, r := range roles {
		switch r {
		case "candidate", "recruiter", "admin":
			scopes = append(scopes, "role:"+r)
		default:
			return "", ErrInvalidRole
		}
	}

	return strings.Join(scopes, " "), nil
}
