// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"fmt"
	"net/http"
)

type (
	Method string

	PublicRouteConfig struct {
		Mux     *http.ServeMux
		Method  Method
		Route   string
		Handler http.HandlerFunc
	}

	ProtectedRouteConfig struct {
		Mux            *http.ServeMux
		Method         Method
		Route          string
		Handler        http.HandlerFunc
		RequiredScopes []ScopeValueType
	}
)

const (
	MethodGet  Method = http.MethodGet
	MethodPost Method = http.MethodPost

	RouteHealth               = "/v1/health"
	RoutePublicKeys           = "/v1/auth/keys"
	RouteToken                = "/v1/auth/token"
	RouteLogin                = "/v1/auth/login/{provider}"
	RouteCallback             = "/v1/auth/callback/{provider}"
	RouteGetMyRecommendations = "/v1/me/recommendations"
	RouteGetMyReactions       = "/v1/me/reactions"
	RouteGetMyMatches         = "/v1/me/matches"
	RouteCreateMyReaction     = "/v1/me/recommendations/{id}/reaction"
)

func routeKey(method Method, route string) string {
	return fmt.Sprintf("%s %s", method, route)
}

func baseMiddleware(handler http.HandlerFunc) http.Handler {
	return Chain(
		handler,
		Logger,
		ErrorHandler,
		MaxBytesLimiter,
	)
}

func PublicRoute(s Store, v Vault, cfg PublicRouteConfig) {
	handler := baseMiddleware(cfg.Handler)

	cfg.Mux.Handle(
		routeKey(cfg.Method, cfg.Route),
		handler,
	)
}

func ProtectedRoute(s Store, v Vault, cfg ProtectedRouteConfig) {
	handler := Chain(
		cfg.Handler,
		Logger,
		ErrorHandler,
		MaxBytesLimiter,
		Authentication(v, cfg.RequiredScopes),
	)

	cfg.Mux.Handle(
		routeKey(cfg.Method, cfg.Route),
		handler,
	)
}

func GetRootMux(s Store, v Vault) http.Handler {
	mux := http.NewServeMux()

	// Public routes
	PublicRoute(s, v, PublicRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RouteHealth,
		Handler: Health,
	})

	PublicRoute(s, v, PublicRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RoutePublicKeys,
		Handler: PublicKeys(v),
	})

	PublicRoute(s, v, PublicRouteConfig{
		Mux:     mux,
		Method:  MethodPost,
		Route:   RouteToken,
		Handler: CreateAccessToken(s, v),
	})

	PublicRoute(s, v, PublicRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RouteLogin,
		Handler: Login(v),
	})

	PublicRoute(s, v, PublicRouteConfig{
		Mux:     mux,
		Method:  MethodPost,
		Route:   RouteLogin,
		Handler: Login(v),
	})

	PublicRoute(s, v, PublicRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RouteCallback,
		Handler: RedirectProvider(s, v),
	})

	PublicRoute(s, v, PublicRouteConfig{
		Mux:     mux,
		Method:  MethodPost,
		Route:   RouteCallback,
		Handler: RedirectProvider(s, v),
	})

	ProtectedRoute(s, v, ProtectedRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RouteGetMyRecommendations,
		Handler: GetMyRecommendations(s),
		RequiredScopes: []ScopeValueType{
			ScopeValueTypeCandidate, ScopeValueTypeRecruiter,
		},
	})

	ProtectedRoute(s, v, ProtectedRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RouteGetMyReactions,
		Handler: GetMyReactions(s),
		RequiredScopes: []ScopeValueType{
			ScopeValueTypeCandidate, ScopeValueTypeRecruiter,
		},
	})

	ProtectedRoute(s, v, ProtectedRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RouteGetMyMatches,
		Handler: GetMyMatches(s),
		RequiredScopes: []ScopeValueType{
			ScopeValueTypeCandidate, ScopeValueTypeRecruiter,
		},
	})

	ProtectedRoute(s, v, ProtectedRouteConfig{
		Mux:     mux,
		Method:  MethodGet,
		Route:   RouteCreateMyReaction,
		Handler: CreateMyReaction(s),
		RequiredScopes: []ScopeValueType{
			ScopeValueTypeCandidate, ScopeValueTypeRecruiter,
		},
	})

	return mux
}
