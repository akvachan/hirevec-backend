// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package server implements the HTTP transport layer, providing RESTful endpoints.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akvachan/hirevec-backend/internal/vault"
)

type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyClaims contextKey = "claims"
)

type RateLimiter struct {
	MaxRequests uint
	Window      time.Duration
}

// PublicMiddleware defines the standard stack for all endpoints, including logging, safety, and rate limiting.
func PublicMiddleware(r RateLimiter) []Middleware {
	if r.MaxRequests == 0 || r.Window == 0 {
		slog.Warn("rate limiter max requests not set, defaulting to 100 per minute")
		r.MaxRequests = 100
		r.Window = time.Minute
	}
	return []Middleware{
		ErrorHandling,
		Logging,
		RateLimit(r.MaxRequests, r.Window),
		MaxBytes,
	}
}

// ProtectedMiddleware adds authentication and authorization layers to the public middleware stack for restricted endpoints.
func ProtectedMiddleware(r RateLimiter, v vault.Vault) []Middleware {
	if r.MaxRequests == 0 || r.Window == 0 {
		slog.Warn("rate limiter max requests not set, defaulting to 100 per minute")
		r.MaxRequests = 100
		r.Window = time.Minute
	}
	return []Middleware{
		ErrorHandling,
		Logging,
		RateLimit(r.MaxRequests, r.Window),
		MaxBytes,
		Auth(v),
	}
}

// Middleware represents a function that wraps an http.Handler to provide pre-processing or post-processing logic.
type Middleware func(http.Handler) http.Handler

// Chain takes a base handler and applies a slice of middlewares in order.
//
// Middlewares are wrapped such that the first middleware in the slice
// is the outermost layer of the onion.
func Chain(handler http.HandlerFunc, middlewares ...Middleware) http.Handler {
	wrapped := http.Handler(handler)
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

// responseWriter is a wrapper around http.ResponseWriter that captures the HTTP status code for logging purposes.
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code before sending it to the underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// ErrorHandling recovers from panics within the request lifecycle.
func ErrorHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("error occurred: %v", err)
				WriteErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// GetUserID retrieves userID from context.
func GetUserID(ctx context.Context) (uint32, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(uint32)
	return userID, ok
}

// GetClaims retrieves claims from context.
func GetClaims(ctx context.Context) (*vault.AccessTokenClaims, bool) {
	claims, ok := ctx.Value(contextKeyClaims).(*vault.AccessTokenClaims)
	return claims, ok
}

// Auth verifies the identity and permissions of the user making the request.
func Auth(v vault.Vault) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			bearer, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found || bearer == "" {
				WriteUnauthorizedResponse(w, AuthInvalidClient, "Bearer token is required")
				return
			}

			claims, err := v.ParseAccessToken(bearer)
			if err != nil {
				WriteAuthErrorResponse(w, AuthInvalidRequest, "invalid access token")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, contextKeyClaims, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RateLimit implements a simple in-memory request throttler based on the remote IP address.
func RateLimit(maxRequests uint, window time.Duration) func(http.Handler) http.Handler {
	type client struct {
		count uint
		reset time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			now := time.Now().UTC()

			mu.Lock()
			c, exists := clients[ip]

			if !exists || now.After(c.reset) {
				c = &client{count: 0, reset: now.Add(window)}
				clients[ip] = c
			}

			c.count++
			currentCount := c.count
			resetAt := c.reset
			mu.Unlock()

			remaining := maxRequests - currentCount
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, remaining)))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if remaining == 0 {
				retryAfter := int(time.Until(resetAt).Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(max(0, retryAfter)))
				WriteErrorResponse(w, http.StatusTooManyRequests, "too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP, considering proxies
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := parseFirstIP(xff); ip != "" {
			return ip
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return xri
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}
	return ip
}

// parseFirstIP extracts the first valid IP from a comma-separated list
func parseFirstIP(xff string) string {
	for i := 0; i < len(xff); i++ {
		if xff[i] == ',' {
			if ip := net.ParseIP(xff[:i]); ip != nil {
				return xff[:i]
			}
			break
		}
	}
	if ip := net.ParseIP(xff); ip != nil {
		return xff
	}
	return ""
}

// Logging records structured information about the HTTP request, including method, path, response status, and processing time.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		slog.Info(
			"request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}

// MaxBytes limits the maximum size of the request body to 1MB to prevent denial-of-service attacks via large payloads.
func MaxBytes(next http.Handler) http.Handler {
	const megabyte = 1_000_000
	return http.MaxBytesHandler(next, megabyte)
}
