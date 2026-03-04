// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package main

import (
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/akvachan/hirevec-backend/internal"
)

func main() {
}

func CreateRefreshToken(userID string, provider string, jti string, symmetricKeyHex string) (*hirevec.RefreshToken, error) {
	now := time.Now().UTC()

	token := paseto.NewToken()
	token.SetAudience("hirevec-api")
	token.SetIssuer("hirevec")
	token.SetSubject(userID)
	token.SetExpiration(now.Add(hirevec.RefreshTokenExpiration))
	token.SetNotBefore(now)
	token.SetIssuedAt(now)
	token.SetJti(jti)

	if err := token.Set("token_type", hirevec.IssuedTokenTypeRefreshToken); err != nil {
		return nil, hirevec.ErrFailedToSetTokenType
	}

	if err := token.Set("provider", provider); err != nil {
		return nil, hirevec.ErrFailedToSetProvider
	}

	symmetricKey, _ := paseto.V4SymmetricKeyFromHex(symmetricKeyHex)

	return &hirevec.RefreshToken{
		RefreshToken: token.V4Encrypt(symmetricKey, nil),
		ExpiresIn:    uint32(hirevec.RefreshTokenExpiration.Abs().Seconds()),
		UserID:       userID,
	}, nil
}
