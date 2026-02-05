// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package store provides an interface to the storage components
package store

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/akvachan/hirevec-backend/internal/store/db/models"
)

type Store interface {
	GetPosition(uint32) (json.RawMessage, error)
	GetPositions(models.Paginator) (json.RawMessage, error)
	GetCandidate(uint32) (json.RawMessage, error)
	GetCandidates(models.Paginator) (json.RawMessage, error)
	CreateCandidateReaction(models.CandidateReaction) error
	CreateRecruiterReaction(models.RecruiterReaction) error
	CreateMatch(models.Match) error
	ValidateActiveSession(string) (bool, error)
	GetUserByProvider(string, string) (uint32, error)
	CreateUser(models.User) (uint32, error)
	CreateRefreshToken(uint32, time.Time) (string, error)
}

type store struct {
	Postgres *sql.DB
}

type StoreConfig struct {
	DBConnString string
}

func NewStore(c StoreConfig) (*store, error) {
	database, err := sql.Open("pgx", c.DBConnString)
	if err != nil {
		return nil, ErrFailedToConnectToDB(err)
	}
	return &store{Postgres: database}, nil
}
