// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	ContextKey string

	Middleware func(http.HandlerFunc) http.HandlerFunc

	ResponseWriter struct {
		http.ResponseWriter
		status int
	}
)

const (
	DefaultPageSizeLimit            = 50
	PageSizeMaxLimit                = 100
	ContextKeyUserID     ContextKey = "user_id"
	ContextKeyClaims     ContextKey = "claims"
)

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Chain wraps handler into a sequence of middlewares, each middleware is applied in the same order it is provided.
func Chain(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

func PanicHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error(
					"panic recovered",
					"err", err,
				)
				Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	}
}

func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(ContextKeyUserID).(string)
	return userID, ok
}

func GetClaims(r *http.Request) (*AccessTokenClaims, bool) {
	claims, ok := r.Context().Value(ContextKeyClaims).(*AccessTokenClaims)
	return claims, ok
}

func GetPagination(r *http.Request) Page {
	p := Page{
		Cursor: r.URL.Query().Get("cursor"),
		Limit:  DefaultPageSizeLimit,
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			p.Limit = min(parsed, PageSizeMaxLimit)
		}
	}
	return p
}

func GetRecommendationsQueryParams(r *http.Request) (RecommendationsQueryParams, error) {
	hideReacted := r.URL.Query().Get("hide_reacted")
	var params RecommendationsQueryParams
	switch hideReacted {
	case "true":
		params.HideReacted = true
	case "false":
		params.HideReacted = false
	case "":
		params.HideReacted = false
	default:
		return params, ErrInvalidHideReactedParamValue
	}
	return params, nil
}

func Authentication(v Vault, allowedScopes []ScopeValueType) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var bearer string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				var found bool
				bearer, found = strings.CutPrefix(authHeader, "Bearer ")
				if !found || bearer == "" {
					Unauthorized(w, AuthInvalidClient, "Bearer token is required")
					return
				}
			}

			claims, err := v.ParseAccessToken(bearer)
			if err != nil || claims == nil {
				AuthError(w, AuthInvalidGrant, "invalid access token")
				return
			}

			allowed := make(map[ScopeValueType]bool, len(allowedScopes))
			for _, s := range allowedScopes {
				allowed[s] = true
			}

			for _, s := range claims.Scope {
				if _, ok := allowed[s]; !ok {
					AuthError(w, AuthInvalidGrant, "unauthorized")
					return
				}
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyClaims, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &ResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info(
			"request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	}
}

func MaxBytesLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1_000_000)
		next.ServeHTTP(w, r)
	}
}
