// Copyright (c) 2026 Arsenii Kvachan. MIT License.

// Package server implements basic routing, middleware, handlers and validation
package server

import (
	"fmt"
	"net/http"
	"strings"
)

type router struct {
	mux    *http.ServeMux
	prefix string
}

type apiVersion uint8

const v0 apiVersion = 0

type route struct {
	method      string
	path        string
	apiVersion  apiVersion
	handler     http.HandlerFunc
	middleware  []middleware
	description string
}

func newRouter(rootMux *http.ServeMux, prefix string) *router {
	if strings.HasPrefix(prefix, "/") {
		panic("prefix cannot have a leading / (slash)")
	}
	if strings.HasSuffix(prefix, "/") {
		panic("prefix cannot have a trailing / (slash)")
	}

	r := &router{
		mux:    http.NewServeMux(),
		prefix: prefix,
	}

	rootMux.Handle("/"+prefix+"/", r.mux)

	return r
}

func (r *router) addRoutes(routes ...route) {
	for _, route := range routes {
		if route.handler == nil {
			panic("handler cannot be nil")
		}
		if strings.HasPrefix(route.path, "/") {
			panic("path cannot have a leading / (slash)")
		}
		if strings.HasSuffix(route.path, "/") {
			panic("path cannot have a trailing / (slash)")
		}
		if route.description == "" {
			panic("description cannot be empty")
		}

		pattern := fmt.Sprintf(
			"%s /%s/v%d/%s",
			route.method,
			r.prefix,
			route.apiVersion,
			route.path,
		)

		handler := chain(route.handler, route.middleware...)
		r.mux.Handle(pattern, handler)
	}
}

func GetRootRouter() *http.ServeMux {
	rootMux := http.NewServeMux()

	apiRouter := newRouter(rootMux, "api")
	apiRouter.addRoutes(
		route{
			method:      http.MethodGet,
			path:        "health",
			apiVersion:  v0,
			handler:     handleHealth,
			middleware:  middlewareGroupPublic,
			description: "Health check endpoint",
		},
		route{
			method:      http.MethodGet,
			path:        "positions",
			apiVersion:  v0,
			handler:     handleGetPositions,
			middleware:  middlewareGroupProtected,
			description: "List all positions",
		},
		route{
			method:      http.MethodGet,
			path:        "positions/{id}",
			apiVersion:  v0,
			handler:     handleGetPosition,
			middleware:  middlewareGroupProtected,
			description: "Get position by ID",
		},
		route{
			method:      http.MethodGet,
			path:        "candidates",
			apiVersion:  v0,
			handler:     handleGetCandidates,
			middleware:  middlewareGroupProtected,
			description: "List all candidates",
		},
		route{
			method:      http.MethodGet,
			path:        "candidates/{id}",
			apiVersion:  v0,
			handler:     handleGetCandidate,
			middleware:  middlewareGroupProtected,
			description: "Get candidate by ID",
		},
	)

	return rootMux
}
