// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type (
	// FailData defines [JSend](https://github.com/omniti-labs/jsend) request failure data.
	FailData map[string]string

	// ResponseStatus defines JSend status codes.
	ResponseStatus string

	// ErrorCode defines JSend error codes.
	ErrorCode uint16

	// RelType defines link relation type, see [RFC5988](https://www.rfc-editor.org/rfc/rfc5988.txt).
	RelType string

	// Link defines a [HAL](https://datatracker.ietf.org/doc/html/draft-kelly-json-hal-11) link object.
	Link struct {
		Href      string `json:"href"`
		Name      string `json:"name,omitempty"`
		Templated bool   `json:"templated,omitempty"`
	}

	Links    map[RelType]Link
	Embedded map[string]any

	// Resource is a flat HAL Resource Object. _links, _embedded, and all
	Resource struct {
		Links    Links          `json:"_links,omitempty"`
		Embedded Embedded       `json:"_embedded,omitempty"`
		Props    map[string]any `json:"-"`
	}

	ErrorResponse struct {
		Status  ResponseStatus `json:"status"`
		Message string         `json:"message"`
		Code    ErrorCode      `json:"code,omitempty"`
	}

	FailResponse struct {
		Status ResponseStatus `json:"status"`
		Data   FailData       `json:"data,omitempty"`
		Links  Links          `json:"_links,omitempty"`
	}
)

var (
	adjectives = []string{
		"fast", "lazy", "clever", "curious", "brave", "mighty", "silent", "noisy", "happy", "grumpy",
	}

	nouns = []string{
		"lion", "tiger", "panda", "fox", "eagle", "shark", "wolf", "dragon", "otter", "koala",
	}
)

const (
	// All went well, and (usually) some data was returned.
	ResponseStatusSuccess = "success"

	// There was a problem with the data submitted, or some pre-condition of the API call wasn't satisfied.
	ResponseStatusFail = "fail"

	// An error occurred in processing the request, i.e. an exception was thrown.
	ResponseStatusError = "error"

	// Conveys an identifier for the link's context.
	RelTypeSelf RelType = "self"

	// Refers to a parent document in a hierarchy of documents.
	RelTypeUp RelType = "up"

	// Refers to the previous resource in an ordered series of resources.
	RelTypePrevious RelType = "previous"

	// Refers to the next resource in a ordered series of resources.
	RelTypeNext RelType = "next"

	// An IRI that refers to the furthest preceding resource in a series of resources.
	RelTypeFirst RelType = "first"

	// An IRI that refers to the furthest following resource in a series of resources.
	RelTypeLast RelType = "last"

	// Refers to an index.
	RelTypeIndex RelType = "index"

	// Refers to a resource offering help (more information, links to other sources information, etc.).
	RelTypeHelp RelType = "help"

	// Refers to a resource that can be used to edit the link's context.
	RelTypeEdit RelType = "edit"
)

func (res Resource) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(res.Props)+2)
	for k, v := range res.Props {
		m[k] = v
	}
	if len(res.Links) > 0 {
		m["_links"] = res.Links
	}
	if len(res.Embedded) > 0 {
		m["_embedded"] = res.Embedded
	}
	return json.Marshal(m)
}

// WriteJSON implements a helper for writing HTTP status and encoding data.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response data", "err", err)
	}
}

func SetDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func Success(w http.ResponseWriter, status int, res Resource) {
	SetDefaultHeaders(w)
	WriteJSON(w, status, res)
}

func Error(w http.ResponseWriter, status int, message string) {
	type ErrorResponse struct {
		Status  ResponseStatus `json:"status"`
		Message string         `json:"message"`
		Code    ErrorCode      `json:"code,omitempty"`
	}
	SetDefaultHeaders(w)
	WriteJSON(w, status, ErrorResponse{Status: ResponseStatusError, Message: message})
}

func Fail(w http.ResponseWriter, status int, data FailData) {
	type FailResponse struct {
		Status ResponseStatus `json:"status"`
		Data   FailData       `json:"data,omitempty"`
		Links  Links          `json:"_links,omitempty"`
	}

	SetDefaultHeaders(w)
	WriteJSON(w, status, FailResponse{Status: ResponseStatusFail, Data: data})
}

func DecodeRequestBody[T any](r *http.Request) (data *T, err error) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(data)
	if err != nil {
		return nil, ErrFailedDecode
	}
	if dec.More() {
		return nil, ErrExtraDataDecoded
	}
	return data, err
}

// GenerateUsername creates username with a cryptographically random suffix
func GenerateUsername() (string, error) {
	randInt := func(n int) int {
		if n <= 0 {
			return 0
		}
		b := make([]byte, 1)
		_, _ = rand.Read(b)
		return int(b[0]) % n
	}
	adj := adjectives[randInt(len(adjectives))]
	noun := nouns[randInt(len(nouns))]

	suffix := make([]byte, 2)
	_, err := rand.Read(suffix)
	if err != nil {
		return "", ErrFailedGenerateUsernameSuffix
	}

	username := fmt.Sprintf("%s_%s%s", adj, noun, hex.EncodeToString(suffix))
	username = strings.ToLower(username)

	return username, nil
}

func Health(w http.ResponseWriter, r *http.Request) {
	Success(w, http.StatusOK, Resource{
		Links: Links{RelTypeSelf: Link{Href: "/v1/health"}},
	})
}

// Returns position recommendations for the authenticated candidate.
func GetMyRecommendations(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r)
		if !ok {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		candidate, err := s.GetCandidateByUserID(userID)
		if err != nil {
			if errors.Is(err, ErrCandidateNotFound) {
				Error(w, http.StatusNotFound, "candidate profile not found")
				return
			}
			Error(w, http.StatusInternalServerError, "failed to fetch candidate profile")
			return
		}

		page := GetPagination(r)

		recommendations, nextCursor, err := s.GetPositionRecommendations(candidate.ID, page.Cursor, page.Limit)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to fetch recommendations")
			return
		}

		page.Count = len(recommendations)
		page.HasNext = nextCursor != ""

		links := Links{
			RelTypeSelf:          Link{Href: "/v1/me/recommendations"},
			RelType("reactions"): Link{Href: "/v1/me/reactions"},
		}
		if nextCursor != "" {
			links[RelTypeNext] = Link{Href: fmt.Sprintf("/v1/me/recommendations?cursor=%s&limit=%d", nextCursor, page.Limit)}
		}

		positions := make([]Resource, len(recommendations))
		for i, rec := range recommendations {
			positions[i] = Resource{
				Links: Links{
					RelTypeSelf:      Link{Href: "/v1/me/recommendations/" + rec.RecommendationID},
					RelType("react"): Link{Href: "/v1/me/recommendations/" + rec.RecommendationID + "/reaction"},
				},
				Props: map[string]any{
					"recommendation_id": rec.RecommendationID,
					"position_id":       rec.PositionID,
					"title":             rec.Title,
					"company":           rec.Company,
					"description":       rec.Description,
				},
			}
		}

		Success(w, http.StatusOK, Resource{
			Links:    links,
			Embedded: Embedded{"positions": positions},
			Props:    map[string]any{"page": page},
		})
	}
}

// Records a candidate's reaction to a position recommendation.
func CreateMyReaction(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r)
		if !ok {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		candidate, err := s.GetCandidateByUserID(userID)
		if err != nil {
			if errors.Is(err, ErrCandidateNotFound) {
				Error(w, http.StatusNotFound, "candidate profile not found")
				return
			}
			Error(w, http.StatusInternalServerError, "failed to fetch candidate profile")
			return
		}

		recommendationID := r.PathValue("id")
		if recommendationID == "" {
			Fail(w, http.StatusBadRequest, FailData{"id": "recommendation id is required"})
			return
		}

		rec, err := s.GetRecommendation(recommendationID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				Error(w, http.StatusNotFound, "recommendation not found")
				return
			}
			Error(w, http.StatusInternalServerError, "failed to fetch recommendation")
			return
		}

		// If recommendation is for a different candidate, send StatusForbidden.
		if rec.CandidateID != candidate.ID {
			Error(w, http.StatusForbidden, "forbidden")
			return
		}

		body, err := DecodeRequestBody[struct {
			Reaction ReactionType `json:"reaction"`
		}](r)
		if err != nil {
			Fail(w, http.StatusBadRequest, FailData{"body": "invalid request body"})
			return
		}

		if !body.Reaction.IsValid() {
			Fail(w, http.StatusBadRequest, FailData{"reaction": "must be one of: positive, negative, neutral"})
			return
		}

		reaction := Reaction{
			RecommendationID: recommendationID,
			ReactorType:      ReactorTypeCandidate,
			ReactorID:        candidate.ID,
			ReactionType:     body.Reaction,
		}
		if err := s.CreateReaction(reaction); err != nil {
			Error(w, http.StatusInternalServerError, "failed to record reaction")
			return
		}

		Success(w, http.StatusCreated, Resource{
			Links: Links{
				RelTypeSelf:          Link{Href: "/v1/me/recommendations/" + recommendationID + "/reaction"},
				RelTypeUp:            Link{Href: "/v1/me/recommendations"},
				RelType("reactions"): Link{Href: "/v1/me/reactions"},
				RelType("matches"):   Link{Href: "/v1/me/matches"},
			},
			Props: map[string]any{
				"recommendation_id": reaction.RecommendationID,
				"reactor_type":      reaction.ReactorType,
				"reactor_id":        reaction.ReactorID,
				"reaction_type":     reaction.ReactionType,
				"reacted_at":        reaction.ReactedAt,
			},
		})
	}
}

// Returns all reactions made by the authenticated candidate.
func GetMyReactions(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r)
		if !ok {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		candidate, err := s.GetCandidateByUserID(userID)
		if err != nil {
			if errors.Is(err, ErrCandidateNotFound) {
				Error(w, http.StatusNotFound, "candidate profile not found")
				return
			}
			Error(w, http.StatusInternalServerError, "failed to fetch candidate profile")
			return
		}

		page := GetPagination(r)

		reactions, nextCursor, err := s.GetReactionsByCandidateID(candidate.ID, page.Cursor, page.Limit)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to fetch reactions")
			return
		}

		page.Count = len(reactions)
		page.HasNext = nextCursor != ""

		links := Links{
			RelTypeSelf: Link{Href: "/v1/me/reactions"},
		}
		if nextCursor != "" {
			links[RelTypeNext] = Link{Href: "/v1/me/reactions?cursor=" + nextCursor}
		}

		// Each embedded reaction links back to the recommendation it was made on.
		embedded := make([]Resource, len(reactions))
		for i, rx := range reactions {
			embedded[i] = Resource{
				Links: Links{
					RelTypeSelf: Link{Href: "/v1/me/recommendations/" + rx.RecommendationID + "/reaction"},
				},
				Props: map[string]any{
					"recommendation_id": rx.RecommendationID,
					"reactor_type":      rx.ReactorType,
					"reactor_id":        rx.ReactorID,
					"reaction_type":     rx.ReactionType,
					"reacted_at":        rx.ReactedAt,
				},
			}
		}

		Success(w, http.StatusOK, Resource{
			Links:    links,
			Embedded: Embedded{"reactions": embedded},
			Props:    map[string]any{"page": page},
		})
	}
}

// Returns all mutual matches for the authenticated candidate.
func GetMyMatches(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r)
		if !ok {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		candidate, err := s.GetCandidateByUserID(userID)
		if err != nil {
			if errors.Is(err, ErrCandidateNotFound) {
				Error(w, http.StatusNotFound, "candidate profile not found")
				return
			}
			Error(w, http.StatusInternalServerError, "failed to fetch candidate profile")
			return
		}

		page := GetPagination(r)

		matches, nextCursor, err := s.GetMatchesByCandidateID(candidate.ID, page.Cursor, page.Limit)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to fetch matches")
			return
		}

		page.Count = len(matches)
		page.HasNext = nextCursor != ""

		links := Links{
			RelTypeSelf: Link{Href: "/v1/me/matches"},
		}
		if nextCursor != "" {
			links[RelTypeNext] = Link{Href: fmt.Sprintf("/v1/me/matches?cursor=%s&limit=%d", nextCursor, page.Limit)}
		}

		embedded := make([]Resource, len(matches))
		for i, m := range matches {
			embedded[i] = Resource{
				Links: Links{
					RelTypeSelf: Link{Href: "/v1/positions/" + m.PositionID},
				},
				Props: map[string]any{
					"position_id": m.PositionID,
					"title":       m.Title,
					"description": m.Description,
					"company":     m.Company,
					"matched_at":  m.MatchedAt,
				},
			}
		}

		Success(w, http.StatusOK, Resource{
			Links:    links,
			Embedded: Embedded{"matches": embedded},
			Props:    map[string]any{"page": page},
		})
	}
}
