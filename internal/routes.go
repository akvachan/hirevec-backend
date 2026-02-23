// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"fmt"
	"net/http"
)

type Route struct {
	Method string
	Href   string
}

func (r Route) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Href)
}

var (
	RouteGetHealth         = Route{http.MethodGet, "/api/v1/health"}
	RouteGetPublicKeys     = Route{http.MethodGet, "/api/v1/auth/keys"}
	RouteCreateAccessToken = Route{http.MethodPost, "/api/v1/auth/token"}
	RouteGetLogin          = Route{http.MethodGet, "/api/v1/auth/login/{provider}"}
	RouteCreateLogin       = Route{http.MethodPost, "/api/v1/auth/login/{provider}"}
	RouteGetCallback       = Route{http.MethodGet, "/api/v1/auth/callback/{provider}"}
	RouteCreateCallback    = Route{http.MethodPost, "/api/v1/auth/callback/{provider}"}

	// DEPRECATED:
	RouteGetPositions     = Route{http.MethodGet, "/api/v1/positions/{id}"}
	RouteGetCandidates    = Route{http.MethodGet, "/api/v1/candidates/{id}"}
	RouteCreateCandidates = Route{http.MethodPost, "/api/v1/candidates"}
)

func GetRootMux(s Store, v Vault) http.Handler {
	mux := http.NewServeMux()

	var (
		health            = Public(Health)
		publicKeys        = Public(GetPublicKeys(v))
		createAccessToken = Public(CreateAccessToken(s, v))
		login             = Public(Login(v))
		callback          = Public(RedirectProvider(s, v))
	)

	mux.Handle(RouteGetHealth.String(), health)
	mux.Handle(RouteGetPublicKeys.String(), publicKeys)
	mux.Handle(RouteCreateAccessToken.String(), createAccessToken)
	mux.Handle(RouteGetLogin.String(), login)
	mux.Handle(RouteCreateLogin.String(), login)
	mux.Handle(RouteGetCallback.String(), callback)
	mux.Handle(RouteCreateCallback.String(), callback)

	return mux
}
