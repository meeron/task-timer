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
	AddWorklog(ctx context.Context, req WorklogRequest) (string, error)
}

// NewWorklogHandler returns the handler for POST WorklogPath.
func NewWorklogHandler(client Worklogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req WorklogRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

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

		id, err := client.AddWorklog(r.Context(), req)
		if err != nil {
			log.Printf("Failed to add Jira worklog to %s: %v", req.IssueKey, err)
			msg := "Could not reach Jira"
			if apiErr, ok := errors.AsType[*APIError](err); ok {
				msg = apiErr.Message
			}
			writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: msg})
			return
		}

		writeJSON(w, http.StatusCreated, WorklogResponse{Id: id})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
