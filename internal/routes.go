// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"net/http"
)

func GetRootMux(localStore Store, localVault Vault) http.Handler {
	rootMux := http.NewServeMux()

	// !!! WARNING: PublicEndpoint wraps handlers with a basic middleware stack WITHOUT AUTHENTICATION AND AUTHORIZATION !!!
	health := PublicEndpoint(Health)
	getPublicKeys := PublicEndpoint(GetPublicKeys(localVault))
	createAccessToken := PublicEndpoint(CreateAccessToken(localStore, localVault))
	login := PublicEndpoint(Login(localVault))
	callback := PublicEndpoint(RedirectProvider(localStore, localVault))

	getCandidates := ProtectedEndpoint(GetCandidates(localStore))
	getCandidate := ProtectedEndpoint(GetCandidate(localStore))
	createCandidate := ProtectedEndpoint(CreateCandidate(localStore, localVault))
	getPositions := ProtectedEndpoint(GetPositions(localStore))
	getPosition := ProtectedEndpoint(GetPosition(localStore))

	rootMux.HandleFunc("GET /api/v1/health", health)
	rootMux.HandleFunc("GET /api/v1/auth/keys", getPublicKeys)
	rootMux.HandleFunc("GET /api/v1/auth/token", createAccessToken)
	rootMux.HandleFunc("GET /api/v1/auth/login/{provider}", login)
	rootMux.HandleFunc("POST /api/v1/auth/login/{provider}", login)
	rootMux.HandleFunc("GET /api/v1/auth/callback/{provider}", callback)
	rootMux.HandleFunc("POST /api/v1/auth/callback/{provider}", callback)
	rootMux.HandleFunc("GET /api/v1/candidates", getCandidates)
	rootMux.HandleFunc("GET /api/v1/candidates/{id}", getCandidate)
	rootMux.HandleFunc("POST /api/v1/candidates", createCandidate)
	rootMux.HandleFunc("GET /api/v1/positions", getPositions)
	rootMux.HandleFunc("GET /api/v1/positions/{id}", getPosition)

	return rootMux
}
