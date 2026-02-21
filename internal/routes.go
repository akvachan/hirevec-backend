// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"net/http"
)

func GetRootMux(localStore Store, localVault Vault) http.Handler {
	rootMux := http.NewServeMux()

	// public endpoints
	// WARNING: PublicEndpoint wraps handlers with a basic middleware stack WITHOUT AUTHENTICATION AND AUTHORIZATION
	getPublicKeys := PublicEndpoint(GetPublicKeys(localVault))
	createAccessToken := PublicEndpoint(CreateAccessToken(localStore, localVault))
	login := PublicEndpoint(Login(localVault))
	callback := PublicEndpoint(RedirectProvider(localStore, localVault))

	// protected endpoint
	getCandidates := ProtectedEndpoint(GetCandidates(localStore))
	getCandidate := ProtectedEndpoint(GetCandidate(localStore))
	createCandidate := ProtectedEndpoint(CreateCandidate(localStore, localVault))
	getPositions := ProtectedEndpoint(GetPositions(localStore))
	getPosition := ProtectedEndpoint(GetPosition(localStore))

	rootMux.Handle("GET 	/api/v1/auth/keys", getPublicKeys)
	rootMux.Handle("GET 	/api/v1/auth/token", createAccessToken)
	rootMux.Handle("GET 	/api/v1/auth/login/{provider}", login)
	rootMux.Handle("POST 	/api/v1/auth/login/{provider}", login)
	rootMux.Handle("GET 	/api/v1/auth/callback/{provider}", callback)
	rootMux.Handle("POST 	/api/v1/auth/callback/{provider}", callback)
	rootMux.Handle("GET 	/api/v1/candidates", getCandidates)
	rootMux.Handle("GET 	/api/v1/candidates/{id}", getCandidate)
	rootMux.Handle("POST 	/api/v1/candidates", createCandidate)
	rootMux.Handle("GET 	/api/v1/positions", getPositions)
	rootMux.Handle("GET 	/api/v1/positions/{id}", getPosition)

	return rootMux
}
