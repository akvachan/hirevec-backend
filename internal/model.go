// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import "time"

type (
	Provider string

	ReactionType string

	ReactorType string

	// User represents a system user
	User struct {
		ID             string   `json:"id,omitempty"`
		Provider       Provider `json:"provider,omitempty"`
		ProviderUserID string   `json:"provider_user_id,omitempty"`
		Email          string   `json:"email,omitempty"`
		FullName       string   `json:"full_name,omitempty"`
		UserName       string   `json:"user_name"`
	}

	// Candidate represents a candidate profile
	Candidate struct {
		ID     string `json:"id"`
		UserID string `json:"user_id,omitempty"`
		About  string `json:"about"`
	}

	// Recruiter represents a recruiter profile
	Recruiter struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
	}

	Recommendation struct {
		ID          string `json:"id"`
		PositionID  string `json:"position_id"`
		CandidateID string `json:"candidate_id"`
	}

	// Position represents a job position
	Position struct {
		ID          string `json:"id"`
		RecruiterID string `json:"recruiter_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Company     string `json:"company"`
	}

	Match struct {
		PositionID  string    `json:"position_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Company     string    `json:"company"`
		MatchedAt   time.Time `json:"matched_at"`
	}

	// Reaction represents either a candidate or recruiter reaction to a recommendation
	Reaction struct {
		RecommendationID string       `json:"recommendation_id"`
		ReactorType      ReactorType  `json:"reactor_type"`
		ReactorID        string       `json:"reactor_id"`
		ReactionType     ReactionType `json:"reaction_type"`
		ReactedAt        time.Time    `json:"reacted_at"`
	}

	Page struct {
		Cursor  string `json:"cursor,omitempty"`
		Limit   int    `json:"limit"`
		Count   int    `json:"count"`
		HasNext bool   `json:"has_next"`
	}

	PositionRecommendation struct {
		RecommendationID string `json:"recommendation_id"`
		PositionID       string `json:"position_id"`
		Title            string `json:"title"`
		Company          string `json:"company"`
		Description      string `json:"description"`
	}

	CandidateRecommendation struct {
		RecommendationID string `json:"recommendation_id"`
		CandidateID      string `json:"candidate_id"`
		UserName         string `json:"username"`
		FullName         string `json:"full_name,omitempty"`
		About            string `json:"about"`
	}
)

const (
	ProviderApple  Provider = "apple"
	ProviderGoogle Provider = "google"

	ReactionTypePositive ReactionType = "positive"
	ReactionTypeNegative ReactionType = "negative"
	ReactionTypeNeutral  ReactionType = "neutral"

	ReactorTypeCandidate ReactorType = "candidate"
	ReactorTypeRecruiter ReactorType = "recruiter"
)

// Raw returns the string representation of the provider
func (p Provider) Raw() string {
	switch p {
	case ProviderApple:
		return "apple"
	case ProviderGoogle:
		return "google"
	default:
		return ""
	}
}

// IsValid checks if the reaction type is valid
func (r ReactionType) IsValid() bool {
	return r == ReactionTypePositive || r == ReactionTypeNegative || r == ReactionTypeNeutral
}

// IsValid checks if the reactor type is valid
func (r ReactorType) IsValid() bool {
	return r == ReactorTypeCandidate || r == ReactorTypeRecruiter
}
