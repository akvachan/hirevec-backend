// Copyright (c) 2026 Arsenii Kvachan. All Rights Reserved. MIT License.

package hirevec

import (
	"fmt"
	"net/http"
)

var (
	positionRoute           = "/api/v0/positions/{id}"
	positionsRoute          = "/api/v0/positions/"
	candidateRoute          = "/api/v0/candidates/{id}"
	candidatesRoute         = "/api/v0/candidates/"
	candidatesReactionRoute = "/api/v0/candidates/{id}/reactions"
	recruitersReactionRoute = "/api/v0/recruiters/{id}/reactions"
	matchesRoute            = "/api/v0/matches/"
)

func registerRoute(
	router *http.ServeMux,
	method string,
	route string,
	handler func(http.ResponseWriter, *http.Request),
) {
	router.HandleFunc(fmt.Sprintf("%v %v", method, route), handler)
}

func registerRoutes() *http.ServeMux {
	r := http.NewServeMux()

	registerRoute(r, http.MethodGet, positionsRoute, handleGetPositions)
	registerRoute(r, http.MethodGet, positionRoute, handleGetPosition)
	registerRoute(r, http.MethodGet, candidatesRoute, handleGetCandidates)
	registerRoute(r, http.MethodGet, candidateRoute, handleGetCandidate)
	registerRoute(r, http.MethodPost, candidatesReactionRoute, handlePostCandidateReaction)
	registerRoute(r, http.MethodPost, recruitersReactionRoute, handlePostRecruiterReaction)
	registerRoute(r, http.MethodPost, matchesRoute, handlePostMatch)

	return r
}
