// Copyright (c) 2026 Arsenii Kvachan. All Rights Reserved. MIT License.

package hirevec

import "encoding/json"

func selectPositionByID(outJSON *json.RawMessage, id int) error {
	return HirevecDatabase.QueryRow(
		`
		SELECT row_to_json(t) 
		FROM general.positions t
		WHERE t.id = $1
		`,
		id,
	).Scan(outJSON)
}

func selectPositions(outJSON *json.RawMessage, p paginator) error {
	return HirevecDatabase.QueryRow(
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
	).Scan(outJSON)
}

func selectCandidateByID(outJSON *json.RawMessage, id int) error {
	return HirevecDatabase.QueryRow(
		`
		SELECT row_to_json(t) 
		FROM general.candidates t
		WHERE t.id = $1
		`,
		id,
	).Scan(outJSON)
}

func selectCandidates(outJSON *json.RawMessage, p paginator) error {
	return HirevecDatabase.QueryRow(
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
	).Scan(outJSON)
}

func insertCandidateReaction(r candidateReaction) error {
	_, err := HirevecDatabase.Exec(
		`
		INSERT INTO general.candidates_reactions (
			candidate_id,
			position_id,
			reaction_type
		)
		VALUES ($1, $2, $3);
		`,
		r.CandidateID,
		r.PositionID,
		r.ReactionType,
	)
	return err
}

func insertRecruiterReaction(r recruiterReaction) error {
	_, err := HirevecDatabase.Exec(
		`
		INSERT INTO general.recruiters_reactions (
			recruiter_id,
			position_id,
			candidate_id,
			reaction_type
		)
		VALUES ($1, $2, $3);
		`,
		r.RecruiterID,
		r.PositionID,
		r.CandidateID,
		r.ReactionType,
	)
	return err
}

func insertMatch(m match) error {
	_, err := HirevecDatabase.Exec(
		`
		INSERT INTO matches (
			candidate_id,
			position_id
		)
		VALUES ($1, $2);
		`,
		m.CandidateID,
		m.PositionID,
	)
	return err
}
