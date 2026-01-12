// Copyright (c) 2026 Arsenii Kvachan. All Rights Reserved. MIT License.

package hirevec

type ReactionType string

const (
	like    ReactionType = "like"
	dislike ReactionType = "dislike"
)

type paginator struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type candidateReaction struct {
	CandidateID  int
	PositionID   int
	ReactionType string
}

type match struct {
	CandidateID int
	PositionID  int
}

type recruiterReaction struct {
	RecruiterID  int
	CandidateID  int
	PositionID   int
	ReactionType string
}

type CandidatesReactionRequest struct {
	PositionID   string       `json:"position_id"`
	ReactionType ReactionType `json:"reaction_type"`
}

type RecruitersReactionRequest struct {
	PositionID   string       `json:"position_id"`
	CandidateID  string       `json:"candidate_id"`
	ReactionType ReactionType `json:"reaction_type"`
}

type MatchRequest struct {
	PositionID  string `json:"position_id"`
	CandidateID string `json:"candidate_id"`
}
