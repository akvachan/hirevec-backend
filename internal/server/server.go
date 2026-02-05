// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package server implements the HTTP transport layer, providing RESTful endpoints.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/akvachan/hirevec-backend/internal/store"
	"github.com/akvachan/hirevec-backend/internal/vault"
)

const (
	DefaultReadTimeout  = 2000 * time.Millisecond
	DefaultWriteTimeout = 2000 * time.Millisecond
	DefaultGracePeriod  = 5000 * time.Millisecond
)

type ServerConfig struct {
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	GracePeriod  time.Duration
}

func Run(ctx context.Context, c ServerConfig, s store.Store, v vault.Vault) error {
	server := newServer(ctx, c, s, v)

	ln, err := net.Listen("tcp", c.Host)
	if err != nil {
		return ErrFailedToBindAddress(c.Host, err)
	}

	errCh := startServer(server, ln)

	return waitAndShutdown(ctx, server, errCh, c.GracePeriod)
}

func newServer(ctx context.Context, c ServerConfig, s store.Store, v vault.Vault) *http.Server {
	return &http.Server{
		Addr:         c.Host,
		Handler:      AssembleTree(s, v),
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}

func startServer(server *http.Server, ln net.Listener) chan error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP server starting", "addr", server.Addr)
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	slog.Info("HTTP server ready", "addr", server.Addr)
	return errCh
}

func waitAndShutdown(ctx context.Context, server *http.Server, errCh chan error, gracePeriod time.Duration) error {
	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		return ErrFailedToShutdownServer(err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
	defer cancel()

	slog.Info("starting graceful shutdown", "timeout", gracePeriod)
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed, forcing close", "err", err)
		server.Close()
		return ErrFailedToShutdownServer(err)
	}

	slog.Info("HTTP server shutdown complete")
	return nil
}
