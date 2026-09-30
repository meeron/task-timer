// Package jira implements the optional Jira Cloud worklog integration.
//
// The browser cannot call Jira directly (Jira Cloud rejects cross-origin
// requests authenticated with API tokens), so the server exposes
// WorklogPath as a proxy and keeps the credentials. The types and helpers in
// this file are shared by the server and the WASM client.
package jira

import "regexp"

const (
	// EnabledEnvKey is passed to the PWA through app.Handler.Env and is set to
	// "true" only when the server has a complete Jira configuration.
	EnabledEnvKey = "JIRA_ENABLED"

	// WorklogPath is the server endpoint that proxies worklogs to Jira.
	WorklogPath = "/api/jira/worklog"

	// MinTimeSpent is the smallest worklog Jira accepts, in seconds.
	MinTimeSpent = 60
)

var issueKeyRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]+-[1-9][0-9]*$`)

// IsValidIssueKey reports whether key looks like a Jira issue key (e.g. ABC-123).
func IsValidIssueKey(key string) bool {
	return issueKeyRe.MatchString(key)
}

// WorklogRequest is the JSON body sent from the client to WorklogPath.
type WorklogRequest struct {
	IssueKey         string `json:"issueKey"`
	Comment          string `json:"comment"`
	StartedUnix      int64  `json:"startedUnix"`
	TimeSpentSeconds int64  `json:"timeSpentSeconds"`
}

// WorklogResponse is returned by WorklogPath on success.
type WorklogResponse struct {
	Id string `json:"id"`
}

// ErrorResponse is returned by WorklogPath on failure.
type ErrorResponse struct {
	Error string `json:"error"`
}
