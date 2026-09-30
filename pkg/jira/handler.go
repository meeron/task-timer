package jira

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

// Worklogger is implemented by Client; it exists so the handler can be tested
// without a real Jira instance.
type Worklogger interface {
	AddWorklog(ctx context.Context, cfg Config, req WorklogRequest) (string, error)
}

// ConnectionTester is implemented by Client; it exists so the handler can be
// tested without a real Jira instance.
type ConnectionTester interface {
	Myself(ctx context.Context, cfg Config) (string, error)
}

// NewWorklogHandler returns the handler for POST WorklogPath.
func NewWorklogHandler(client Worklogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body ProxyRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		cfg, err := NormalizeConfig(body.Config)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		req := body.Worklog

		req.IssueKey = strings.ToUpper(strings.TrimSpace(req.IssueKey))
		if !IsValidIssueKey(req.IssueKey) {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid Jira issue key"})
			return
		}
		if req.TimeSpentSeconds < MinTimeSpent {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Time spent must be at least 1 minute"})
			return
		}
		if req.StartedUnix <= 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid start time"})
			return
		}

		id, err := client.AddWorklog(r.Context(), cfg, req)
		if err != nil {
			log.Printf("Failed to add Jira worklog to %s: %v", req.IssueKey, err)
			writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: upstreamErrorMessage(err)})
			return
		}

		writeJSON(w, http.StatusCreated, WorklogResponse{Id: id})
	})
}

// NewTestHandler returns the handler for POST TestPath. It checks a Config by
// fetching the Jira user it authenticates as.
func NewTestHandler(client ConnectionTester) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body Config
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		cfg, err := NormalizeConfig(body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		name, err := client.Myself(r.Context(), cfg)
		if err != nil {
			log.Printf("Jira connection test for %s failed: %v", cfg.BaseURL, err)
			writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: upstreamErrorMessage(err)})
			return
		}

		writeJSON(w, http.StatusOK, TestResponse{DisplayName: name})
	})
}

// upstreamErrorMessage returns the message to show the user for a failed
// Jira call.
func upstreamErrorMessage(err error) string {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.Message
	}
	return "Could not reach Jira"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
