// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

type (
	Provider     string
	ReactionType string

	Position struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Company     string `json:"company"`
	}

	User struct {
		Provider       Provider `json:"provider,omitempty"`
		ProviderUserID string   `json:"provider_user_id,omitempty"`
		Email          string   `json:"email,omitempty"`
		FirstName      string   `json:"first_name,omitempty"`
		LastName       string   `json:"last_name,omitempty"`
		FullName       string   `json:"full_name,omitempty"`
		UserName       string   `json:"user_name"`
	}

	Candidate struct {
		ID     string `json:"id"`
		UserID string `json:"user_id,omitempty"`
		About  string `json:"about"`
	}

	Recruiter struct {
		UserID string `json:"user_id"`
	}

	Match struct {
		CandidateID string `json:"candidate_id"`
		PositionID  string `json:"position_id"`
	}

	CandidateReaction struct {
		CandidateID  string       `json:"candidate_id"`
		PositionID   string       `json:"position_id"`
		ReactionType ReactionType `json:"reaction_type"`
	}

	RecruiterReaction struct {
		RecruiterID  string       `json:"recruiter_id"`
		CandidateID  string       `json:"candidate_id"`
		PositionID   string       `json:"position_id"`
		ReactionType ReactionType `json:"reaction_type"`
	}
)

const (
	ProviderApple        Provider     = "apple"
	ProviderGoogle       Provider     = "google"
	ReactionTypePositive ReactionType = "positive"
	ReactionTypeNegative ReactionType = "negative"
)

func (p Provider) Raw() string {
	if p == ProviderApple {
		return "apple"
	}
	if p == ProviderGoogle {
		return "google"
	}
	return ""
}

func (r ReactionType) IsValid() bool {
	return r == ReactionTypePositive || r == ReactionTypeNegative
}
