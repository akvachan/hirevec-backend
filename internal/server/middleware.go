// Copyright (c) 2026 Arsenii Kvachan. MIT License.

// Package server implements basic routing, middleware, handlers and validation
package server

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var HirevecLogger *slog.Logger

var (
	middlewareGroupPublic = []middleware{
		middlewareLogging,
		middlewareErrorHandling,
		middlewareMaxBytes,
		middlewareRateLimit,
	}

	middlewareGroupProtected = append(
		middlewareGroupPublic,
		middlewareAuthentication,
		middlewareAuthorization,
	)
)

type middleware func(http.Handler) http.Handler

func chain(handler http.HandlerFunc, middlewares ...middleware) http.Handler {
	wrapped := http.Handler(handler)
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func middlewareErrorHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				HirevecLogger.Error("Error occurred: %v", err)
				writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func middlewareAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// user, pass, ok := r.BasicAuth()
		// if !ok || user != "admin" || pass != "password" {
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }
		next.ServeHTTP(w, r)
	})
}

func middlewareAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// user, pass, ok := r.BasicAuth()
		// if !ok || user != "admin" || pass != "password" {
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }
		next.ServeHTTP(w, r)
	})
}

func middlewareRateLimit(next http.Handler) http.Handler {
	var mu sync.Mutex
	requests := make(map[string]int)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests[r.RemoteAddr]++
		mu.Unlock()

		if requests[r.RemoteAddr] > 5 {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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

func middlewareMaxBytes(next http.Handler) http.Handler {
	const megabyte = 1_000_000

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.MaxBytesHandler(next, megabyte).ServeHTTP(w, r)
	})
}
