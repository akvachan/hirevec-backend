// Copyright (c) 2026 Arsenii Kvachan. MIT License.

// Package server implements basic routing, middleware, handlers and validation
package server

import (
	"fmt"
	"net/http"
	"strings"
)

type router struct {
	mux *http.ServeMux
}

type apiVersion int

const v0 apiVersion = 0

type route struct {
	method          string
	path            string
	apiVersion      apiVersion
	handler         http.HandlerFunc
	middlewareGroup []middleware
}

func newRouter() *router {
	return &router{
		mux: http.NewServeMux(),
	}
}

func (r *router) addRoutes(routes ...route) {
	for _, route := range routes {
		pattern := fmt.Sprintf(
			"%s /api/v%d/%s",
			route.method,
			route.apiVersion,
			strings.TrimPrefix(route.path, "/"),
		)
		handler := chain(route.handler, route.middlewareGroup...)
		r.mux.Handle(pattern, handler)
	}
}

func GetRootRouter() *http.ServeMux {
	rootMux := http.NewServeMux()
	apiRouter := newRouter()
	apiRouter.addRoutes(
		route{
			method:          http.MethodGet,
			path:            "health",
			apiVersion:      v0,
			handler:         handleHealth,
			middlewareGroup: middlewareGroupPublic,
		},
		route{
			method:          http.MethodGet,
			path:            "positions",
			apiVersion:      v0,
			handler:         handleGetPositions,
			middlewareGroup: middlewareGroupProtected,
		},
		route{
			method:          http.MethodGet,
			path:            "positions/{id}",
			apiVersion:      v0,
			handler:         handleGetPosition,
			middlewareGroup: middlewareGroupProtected,
		},
		route{
			method:          http.MethodGet,
			path:            "candidates",
			apiVersion:      v0,
			handler:         handleGetCandidates,
			middlewareGroup: middlewareGroupProtected,
		},
		route{
			method:          http.MethodGet,
			path:            "candidates/{id}",
			apiVersion:      v0,
			handler:         handleGetCandidate,
			middlewareGroup: middlewareGroupProtected,
		},
	)
	rootMux.Handle("/api/", apiRouter.mux)

	return rootMux
}
