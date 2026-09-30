// Package jira implements the optional Jira Cloud worklog integration.
//
// The browser cannot call Jira directly (Jira Cloud rejects cross-origin
// requests authenticated with API tokens), so the server exposes
// WorklogPath as a stateless proxy. Each user keeps their own Jira connection
// settings in the browser (IndexedDB) and sends them with every request; the
// server never stores them. The types and helpers in this file are shared by
// the server and the WASM client.
package jira

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

const (
	// WorklogPath is the server endpoint that proxies worklogs to Jira.
	WorklogPath = "/api/jira/worklog"

	// TestPath is the server endpoint that checks connection settings by
	// fetching the authenticated Jira user.
	TestPath = "/api/jira/test"

	// MinTimeSpent is the smallest worklog Jira accepts, in seconds.
	MinTimeSpent = 60
)

var issueKeyRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]+-[1-9][0-9]*$`)

// IsValidIssueKey reports whether key looks like a Jira issue key (e.g. ABC-123).
func IsValidIssueKey(key string) bool {
	return issueKeyRe.MatchString(key)
}

// Config holds a user's Jira Cloud connection settings.
type Config struct {
	BaseURL  string `json:"baseUrl"` // e.g. https://your-org.atlassian.net
	Email    string `json:"email"`
	APIToken string `json:"apiToken"`
}

// IsSet reports whether all settings are filled in.
func (c Config) IsSet() bool {
	return c.BaseURL != "" && c.Email != "" && c.APIToken != ""
}

// NormalizeConfig trims the settings, reduces BaseURL to its scheme and host,
// and validates them. Only https Jira Cloud sites (*.atlassian.net,
// *.jira.com) are accepted so the proxy cannot be used to reach arbitrary
// hosts.
func NormalizeConfig(c Config) (Config, error) {
	c.Email = strings.TrimSpace(c.Email)
	c.APIToken = strings.TrimSpace(c.APIToken)

	u, err := url.Parse(strings.TrimSpace(c.BaseURL))
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || !isJiraCloudHost(u.Hostname()) {
		return Config{}, errors.New("Jira URL must look like https://your-org.atlassian.net")
	}
	c.BaseURL = "https://" + strings.ToLower(u.Hostname())

	if c.Email == "" {
		return Config{}, errors.New("Enter the email of your Atlassian account")
	}
	if c.APIToken == "" {
		return Config{}, errors.New("Enter an Atlassian API token")
	}
	return c, nil
}

// isJiraCloudHost reports whether host is a single-label subdomain of a Jira
// Cloud domain, e.g. your-org.atlassian.net.
func isJiraCloudHost(host string) bool {
	host = strings.ToLower(host)
	for _, suffix := range []string{".atlassian.net", ".jira.com"} {
		if name, ok := strings.CutSuffix(host, suffix); ok && name != "" && !strings.Contains(name, ".") {
			return true
		}
	}
	return false
}

// ProxyRequest is the JSON body sent from the client to WorklogPath.
type ProxyRequest struct {
	Config  Config         `json:"config"`
	Worklog WorklogRequest `json:"worklog"`
}

// WorklogRequest describes a worklog to add to a Jira issue.
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

// TestResponse is returned by TestPath on success. The request body is a
// Config.
type TestResponse struct {
	DisplayName string `json:"displayName"`
}

// ErrorResponse is returned by WorklogPath and TestPath on failure.
type ErrorResponse struct {
	Error string `json:"error"`
}
