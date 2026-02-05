// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/akvachan/hirevec-backend/internal/store/db/models"
)

// GetPosition retrieves a single position from the database by its unique identifier.
func (s store) GetPosition(positionID uint32) (json.RawMessage, error) {
	var j json.RawMessage
	err := s.Postgres.QueryRow(
		`
		SELECT row_to_json(t) 
		FROM general.positions t
		WHERE t.id = $1
		`,
		positionID,
	).Scan(&j)
	return j, err
}

// GetPositions retrieves a paginated list of all positions, ordered by ID.
func (s store) GetPositions(p models.Paginator) (json.RawMessage, error) {
	var j json.RawMessage
	err := s.Postgres.QueryRow(
		`
		SELECT COALESCE(json_agg(t), '[]'::json)
		FROM (
			SELECT *
			FROM general.positions
			ORDER BY id
			LIMIT $1 OFFSET $2
		) t
		`,
		p.Limit,
		p.Offset,
	).Scan(&j)
	return j, err
}

// GetCandidate retrieves a single candidate's details by their ID.
func (s store) GetCandidate(candidateID uint32) (json.RawMessage, error) {
	var j json.RawMessage
	err := s.Postgres.QueryRow(
		`
		SELECT row_to_json(t) 
		FROM general.candidates t
		WHERE t.id = $1
		`,
		candidateID,
	).Scan(&j)
	return j, err
}

// GetCandidates retrieves a paginated list of candidates, ordered by ID.
func (s store) GetCandidates(p models.Paginator) (json.RawMessage, error) {
	var j json.RawMessage
	err := s.Postgres.QueryRow(
		`
		SELECT COALESCE(json_agg(t), '[]'::json)
		FROM (
			SELECT *
			FROM general.candidates
			ORDER BY id 
			LIMIT $1 OFFSET $2
		) t
		`,
		p.Limit,
		p.Offset,
	).Scan(&j)
	return j, err
}

// GetUserByProvider retrieves an existing user based on their provider details.
func (s store) GetUserByProvider(provider string, providerUserID string) (userID uint32, err error) {
	err = s.Postgres.QueryRow(
		`
		SELECT id 
		FROM general.users 
		WHERE provider = $1 
		AND provider_user_id = $2
		`,
		provider,
		providerUserID,
	).Scan(&userID)
	return userID, err
}

// CreateUser generates a unique username and inserts a new user record.
func (s store) CreateUser(user models.User) (userID uint32, err error) {
	if user.FirstName == "" || user.LastName == "" || user.FullName == "" {
		return 0, errors.New("empty names provided")
	}

	suffix := make([]byte, 2)
	_, err = rand.Read(suffix)
	if err != nil {
		return 0, errors.New("could not generate a random suffix")
	}

	userName := fmt.Sprintf("%s_%s_%s",
		strings.ToLower(user.FirstName),
		strings.ToLower(user.LastName),
		hex.EncodeToString(suffix),
	)

	err = s.Postgres.QueryRow(
		`
		INSERT INTO general.users (
			provider,
			provider_user_id, 
			email,
			full_name,
			user_name
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`,
		user.Provider,
		user.ProviderUserID,
		user.Email,
		user.FullName,
		userName,
	).Scan(&userID)
	return userID, err
}

// CreateCandidateReaction records a candidate's interest or reaction to a specific job position.
func (s store) CreateCandidateReaction(r models.CandidateReaction) error {
	_, err := s.Postgres.Exec(
		`
		INSERT INTO general.candidates_reactions (
			candidate_id,
			position_id,
			reaction_type
		)
		VALUES ($1, $2, $3)
		`,
		r.CandidateID,
		r.PositionID,
		r.ReactionType,
	)
	return err
}

// CreateRecruiterReaction records a recruiter's reaction to a specific candidate for a position.
func (s store) CreateRecruiterReaction(r models.RecruiterReaction) error {
	_, err := s.Postgres.Exec(
		`
		INSERT INTO general.recruiters_reactions (
			recruiter_id,
			position_id,
			candidate_id,
			reaction_type
		)
		VALUES ($1, $2, $3, $4)
		`,
		r.RecruiterID,
		r.PositionID,
		r.CandidateID,
		r.ReactionType,
	)
	return err
}

// CreateMatch creates a new match record between a candidate and a position when mutual interest is established.
func (s store) CreateMatch(m models.Match) error {
	_, err := s.Postgres.Exec(
		`
		INSERT INTO general.matches (
			candidate_id,
			position_id
		)
		VALUES ($1, $2)
		`,
		m.CandidateID,
		m.PositionID,
	)
	return err
}

// ValidateActiveSession checks if the JTI exists and is not expired.
func (s store) ValidateActiveSession(jti string) (isSessionRevoked bool, err error) {
	err = s.Postgres.QueryRow(
		`
		SELECT revoked 
	 	FROM general.refresh_tokens 
	 	WHERE jti = $1 
		AND expires_at > NOW()
		`,
		jti,
	).Scan(&isSessionRevoked)
	return isSessionRevoked, err
}

// CreateRefreshToken creates a new refresh token record.
func (s store) CreateRefreshToken(userID uint32, expiresAt time.Time) (jti string, err error) {
	err = s.Postgres.QueryRow(
		`
		INSERT INTO general.refresh_tokens (
			user_id,
			expires_at
		)
		VALUES ($1, $2) 
		RETURNING jti
		`,
		userID,
		expiresAt,
	).Scan(&jti)
	return jti, err
}
