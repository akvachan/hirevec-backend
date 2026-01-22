// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package server implements the HTTP transport layer, providing RESTful endpoints.
package server

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// HirevecLogger is the global structured logger for the server package.
var HirevecLogger *slog.Logger

var (
	// middlewareGroupPublic defines the standard stack for all endpoints, including logging, safety, and rate limiting.
	middlewareGroupPublic = []middleware{
		middlewareLogging,
		middlewareErrorHandling,
		middlewareMaxBytes,
	}

	// middlewareGroupProtected adds authentication and authorization layers to the public middleware stack for restricted endpoints.
	middlewareGroupProtected = append(
		middlewareGroupPublic,
		middlewareAuthentication,
		middlewareAuthorization,
	)
)

// middleware represents a function that wraps an http.Handler to provide pre-processing or post-processing logic.
type middleware func(http.Handler) http.Handler

// chain takes a base handler and applies a slice of middlewares in order.
//
// Middlewares are wrapped such that the first middleware in the slice
// is the outermost layer of the onion.
func chain(handler http.HandlerFunc, middlewares ...middleware) http.Handler {
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

// middlewareErrorHandling recovers from panics within the request lifecycle and returns a 500 Internal Server Error to the client.
func middlewareErrorHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				HirevecLogger.Error("error occurred: %v", err)
				writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// middlewareAuthentication verifies the identity of the user making the request.
func middlewareAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
		next.ServeHTTP(w, r)
	})
}

// middlewareAuthorization ensures the authenticated user has permission to access the requested resource.
func middlewareAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
		next.ServeHTTP(w, r)
	})
}

// middlewareRateLimit implements a simple in-memory request throttler based on the remote IP address.
func middlewareRateLimit(maxRequests int, window time.Duration) middleware {
	return func(next http.Handler) http.Handler {
		type client struct {
			count int
			reset time.Time
		}
		var (
			mu      sync.RWMutex
			clients = make(map[string]*client)
		)

		stopCleanup := make(chan struct{})
		go func() {
			ticker := time.NewTicker(window)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					mu.Lock()
					now := time.Now()
					for ip, c := range clients {
						if now.After(c.reset) {
							delete(clients, ip)
						}
					}
					mu.Unlock()
				case <-stopCleanup:
					return
				}
			}
		}()

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if ip == "" {
				writeErrorResponse(w, http.StatusBadRequest, "invalid remote address")
				return
			}

			now := time.Now()

			mu.Lock()
			c, exists := clients[ip]
			if !exists || now.After(c.reset) {
				c = &client{
					count: 1,
					reset: now.Add(window),
				}
				clients[ip] = c
			} else {
				c.count++
			}
			count := c.count
			resetAt := c.reset
			mu.Unlock()

			remaining := maxRequests - count

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, remaining)))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if remaining < 0 {
				retryAfter := max(int(time.Until(resetAt).Seconds()), 0)
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				writeErrorResponse(w, http.StatusTooManyRequests, "too many requests")
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

// middlewareLogging records structured information about the HTTP request, including method, path, response status, and processing time.
func middlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		HirevecLogger.Info(
			"request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}

// middlewareMaxBytes limits the maximum size of the request body to 1MB to prevent denial-of-service attacks via large payloads.
func middlewareMaxBytes(next http.Handler) http.Handler {
	const megabyte = 1_000_000

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.MaxBytesHandler(next, megabyte).ServeHTTP(w, r)
	})
}
