// Copyright (c) 2026 Arsenii Kvachan. All Rights Reserved. MIT License.

// Package server implements basic routing, middleware, handlers and validation
package server

import (
	"fmt"
	"net/http"
)

var HirevecServer *http.Server

var (
	routePosition           = "/positions/{id}"
	routePositions          = "/positions"
	routeCandidate          = "/candidates/{id}"
	routeCandidates         = "/candidates"
	routeCandidatesReaction = "/candidates/{id}/reactions"
	routeRecruitersReaction = "/recruiters/{id}/reactions"
	routeMatches            = "/matches"
)

var (
	routerAPI       = http.NewServeMux()
	routerV0        = http.NewServeMux()
	routerEndpoints = http.NewServeMux()
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
	registerRoute(routerEndpoints, http.MethodGet, routePositions, handleGetPositions)
	registerRoute(routerEndpoints, http.MethodGet, routePosition, handleGetPosition)
	registerRoute(routerEndpoints, http.MethodGet, routeCandidates, handleGetCandidates)
	registerRoute(routerEndpoints, http.MethodGet, routeCandidate, handleGetCandidate)
	registerRoute(routerEndpoints, http.MethodPost, routeCandidatesReaction, handlePostCandidateReaction)
	registerRoute(routerEndpoints, http.MethodPost, routeRecruitersReaction, handlePostRecruiterReaction)
	registerRoute(routerEndpoints, http.MethodPost, routeMatches, handlePostMatch)

	routerAPI.Handle("/api/", http.StripPrefix("/api", routerV0))
	routerV0.Handle("/v0/", http.StripPrefix("/v0", routerEndpoints))

	return routerAPI
}
