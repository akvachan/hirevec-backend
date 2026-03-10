// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akvachan/hirevec-backend/cmd/common"
	"github.com/akvachan/hirevec-backend/internal"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := common.Loadenv(".env"); err != nil {
		slog.Warn("could not load .env, using system environment", "err", err)
	}
	pgHost := common.Getenv("POSTGRES_HOST", "localhost")
	pgPort := common.Getenv("POSTGRES_PORT", "5432")
	pgDB := common.Getenv("POSTGRES_DB", "hirevec")
	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgPortParsed := hirevec.ParseUint16WithDefault(pgPort, 5432)

	storeCfg := hirevec.StoreConfig{
		PostgresHost:     pgHost,
		PostgresPort:     pgPortParsed,
		PostgresDB:       pgDB,
		PostgresUser:     pgUser,
		PostgresPassword: pgPassword,
	}
	store, err := hirevec.NewPostgresStore(storeCfg)
	if err != nil {
		die("could not create a new store", "err", err)
	}

	userID, _, err := store.GetUserByProvider(hirevec.ProviderGoogle, "admin")
	switch {
	case errors.Is(err, hirevec.ErrUserNotFound):
		slog.Info("provisioning an admin")

		admin := hirevec.User{
			Provider:       hirevec.ProviderGoogle,
			ProviderUserID: "admin",
			Email:          "admin@admin.com",
			FirstName:      "admin",
			LastName:       "admin",
			FullName:       "admin",
			UserName:       "admin",
		}

		userID, err = store.CreateUser(admin)
		if err != nil {
			die("could not create an admin", "err", err)
		}
	case errors.Is(err, hirevec.ErrUserNoRole):
	default:
		die("could not get an admin", "err", err)
	}

	vaultCfg := hirevec.VaultConfig{
		SymmetricKeyHex:       os.Getenv("SYMMETRIC_KEY"),
		AsymmetricKeyHex:      os.Getenv("ASYMMETRIC_KEY"),
		AccessTokenExpiration: 365 * 24 * time.Hour, // 1 year
	}
	vault, err := hirevec.NewPasetoVault(ctx, vaultCfg)
	if err != nil {
		die("could not create a new vault", "err", err)
	}

	token, err := vault.CreateAccessToken(userID, hirevec.ProviderGoogle.Raw(), hirevec.ScopeTypeAdmin)
	if err != nil {
		die("could not create a refresh token", "err", err)
	}

	fmt.Printf("%s\n", token.AccessToken)
}

func die(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
