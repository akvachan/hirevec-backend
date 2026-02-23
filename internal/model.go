// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

type Provider string

const (
	ProviderApple  Provider = "apple"
	ProviderGoogle Provider = "google"
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

type User struct {
	Provider       Provider
	ProviderUserID string
	Email          string
	FirstName      string
	LastName       string
	FullName       string
}

type Candidate struct {
	UserID string
	About  string
}

type Recruiter struct {
	UserID string
}

// Paginator defines parameters for paginating database queries and API responses.
type Paginator struct {
	// Limit is the maximum number of records to return.
	Limit uint8 `json:"limit"`

	// Offset is the number of records to skip before starting to return results.
	Offset uint8 `json:"offset"`
}

// Match represents a successful connection between a candidate and a specific job position.
type Match struct {
	CandidateID string
	PositionID  string
}

// ReactionType defines a restricted set of strings representing user sentiment.
type ReactionType string

const (
	// ReactionTypePositive indicates interest or approval.
	ReactionTypePositive ReactionType = "positive"

	// ReactionTypeNegative indicates a lack of interest or rejection.
	ReactionTypeNegative ReactionType = "negative"
)

func (r ReactionType) IsValid() bool {
	return r == ReactionTypePositive || r == ReactionTypeNegative
}

// CandidateReaction represents the internal model for a candidate's response to a job posting.
type CandidateReaction struct {
	CandidateID  string
	PositionID   string
	ReactionType ReactionType
}

// RecruiterReaction represents the internal model for a recruiter's response to a specific candidate.
type RecruiterReaction struct {
	RecruiterID  string
	CandidateID  string
	PositionID   string
	ReactionType ReactionType
}
