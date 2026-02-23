// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package hirevec

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ParamLocation string

const (
	ParamLocationBody   ParamLocation = "body"
	ParamLocationQuery  ParamLocation = "query"
	ParamLocationHeader ParamLocation = "header"
	ParamLocationPath   ParamLocation = "path"
)

type ParamType string

const (
	ParamTypeInteger ParamType = "integer"
	ParamTypeFloat   ParamType = "float"
	ParamTypeString  ParamType = "string"
)

type ActionParam struct {
	Name     string        `json:"name"` // must be in kebab-case
	Location ParamLocation `json:"location"`
	Type     ParamType     `json:"type"`
	Required bool          `json:"required"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type Action struct {
	Rel    string        `json:"rel"`
	Name   string        `json:"name"`
	Method string        `json:"method"`
	Href   string        `json:"href"`
	Params []ActionParam `json:"params,omitempty"`
}

type ResponseStatus string

const (
	ResponseStatusSuccess = "success"
	ResponseStatusError   = "error"
	ResponseStatusFail    = "fail"
)

type SuccessResponse struct {
	Status  ResponseStatus `json:"status"`
	Data    any            `json:"data,omitempty"`
	Actions []Action       `json:"actions,omitempty"`
	Links   []Link         `json:"links,omitempty"`
}

type ErrorResponse struct {
	Status  ResponseStatus `json:"status"`
	Message string         `json:"message"`
}

type FailResponse struct {
	Status  ResponseStatus `json:"status"`
	Data    any            `json:"data"`
	Actions []Action       `json:"actions,omitempty"`
	Links   []Link         `json:"links,omitempty"`
}

type AuthErrorCode string

const (
	AuthInvalidRequest       AuthErrorCode = "invalid_request"
	AuthInvalidGrant         AuthErrorCode = "invalid_grant"
	AuthInvalidClient        AuthErrorCode = "invalid_client"
	AuthUnsupportedGrantType AuthErrorCode = "unsupported_grant_type"
)

type AuthErrorResponse struct {
	Error            AuthErrorCode `json:"error"`
	ErrorDescription string        `json:"error_description,omitempty"`
	ErrorURI         string        `json:"error_uri,omitempty"`
	Actions          []Action      `json:"actions,omitempty"`
	Links            []Link        `json:"links,omitempty"`
}

type ResponseContext struct {
	Links   []Link
	Actions []Action
}

func UnpackContext(rctx []ResponseContext) ([]Link, []Action) {
	if len(rctx) > 0 {
		return rctx[0].Links, rctx[0].Actions
	}
	return nil, nil
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("could not encode response data", "err", err)
	}
}

func SetDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func SetAuthHeaders(w http.ResponseWriter) {
	SetDefaultHeaders(w)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}

func Success(w http.ResponseWriter, status int, data any, rctx ...ResponseContext) {
	links, actions := UnpackContext(rctx)
	SetDefaultHeaders(w)
	WriteJSON(w, status, SuccessResponse{ResponseStatusSuccess, data, actions, links})
}

func Error(w http.ResponseWriter, status int, message string) {
	SetDefaultHeaders(w)
	WriteJSON(w, status, ErrorResponse{ResponseStatusError, message})
}

func Fail(w http.ResponseWriter, status int, data any, rctx ...ResponseContext) {
	links, actions := UnpackContext(rctx)
	SetDefaultHeaders(w)
	WriteJSON(w, status, FailResponse{ResponseStatusFail, data, actions, links})
}

func AuthSuccess(w http.ResponseWriter, data any) {
	SetAuthHeaders(w)
	WriteJSON(w, http.StatusOK, data)
}

func AuthError(w http.ResponseWriter, code AuthErrorCode, description string, rctx ...ResponseContext) {
	links, actions := UnpackContext(rctx)
	SetAuthHeaders(w)
	WriteJSON(w, http.StatusBadRequest, AuthErrorResponse{Error: code, ErrorDescription: description, Actions: actions, Links: links})
}

func Unauthorized(w http.ResponseWriter, code AuthErrorCode, description string, rctx ...ResponseContext) {
	links, actions := UnpackContext(rctx)
	SetAuthHeaders(w)
	w.Header().Set("WWW-Authenticate", "Bearer")
	WriteJSON(w, http.StatusUnauthorized, AuthErrorResponse{Error: code, ErrorDescription: description, Actions: actions, Links: links})
}
