package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// startedLayout is the timestamp format Jira expects for worklog "started".
const startedLayout = "2006-01-02T15:04:05.000-0700"

// APIError is returned when Jira responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("jira: %s (HTTP %d)", e.Message, e.StatusCode)
}

// Client talks to the Jira Cloud REST API v3. It holds no credentials; every
// call gets the requesting user's Config.
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 15 * time.Second}}
}

// AddWorklog creates a worklog on req.IssueKey and returns its Jira id.
func (c *Client) AddWorklog(ctx context.Context, cfg Config, req WorklogRequest) (string, error) {
	body := map[string]any{
		"timeSpentSeconds": req.TimeSpentSeconds,
		"started":          time.Unix(req.StartedUnix, 0).Format(startedLayout),
	}
	if comment := strings.TrimSpace(req.Comment); comment != "" {
		body["comment"] = toADF(comment)
	}

	var created struct {
		Id string `json:"id"`
	}
	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog", url.PathEscape(req.IssueKey))
	if err := c.do(ctx, cfg, http.MethodPost, path, body, &created); err != nil {
		return "", err
	}
	return created.Id, nil
}

// Myself returns the display name of the account cfg authenticates as. It is
// used to test the connection settings.
func (c *Client) Myself(ctx context.Context, cfg Config) (string, error) {
	var user struct {
		DisplayName string `json:"displayName"`
	}
	if err := c.do(ctx, cfg, http.MethodGet, "/rest/api/3/myself", nil, &user); err != nil {
		if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.StatusCode == http.StatusNotFound {
			apiErr.Message = "Jira site not found, check the Jira URL"
		}
		return "", err
	}
	return user.DisplayName, nil
}

// do sends a request to the Jira REST API at cfg.BaseURL+path, JSON-encoding
// body when it is not nil, and decodes a 2xx JSON response into out. Non-2xx
// responses are returned as *APIError.
func (c *Client) do(ctx context.Context, cfg Config, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(payload)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, cfg.BaseURL+path, reqBody)
	if err != nil {
		return err
	}
	httpReq.SetBasicAuth(cfg.Email, cfg.APIToken)
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &APIError{StatusCode: resp.StatusCode, Message: errorMessage(resp.StatusCode, respBody)}
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("jira: decoding response: %w", err)
	}
	return nil
}

// toADF converts plain text to an Atlassian Document Format document, one
// paragraph per line. Empty lines become empty paragraphs because ADF text
// nodes must not be empty.
func toADF(text string) map[string]any {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	content := make([]any, 0, len(lines))
	for _, line := range lines {
		paragraph := map[string]any{"type": "paragraph"}
		if line != "" {
			paragraph["content"] = []any{map[string]any{"type": "text", "text": line}}
		}
		content = append(content, paragraph)
	}
	return map[string]any{"type": "doc", "version": 1, "content": content}
}

// errorMessage extracts a human-readable message from a Jira error body of the
// form {"errorMessages": [...], "errors": {"field": "msg"}}.
func errorMessage(status int, body []byte) string {
	var jiraErr struct {
		ErrorMessages []string          `json:"errorMessages"`
		Errors        map[string]string `json:"errors"`
	}
	if json.Unmarshal(body, &jiraErr) == nil {
		msgs := jiraErr.ErrorMessages
		for _, msg := range jiraErr.Errors {
			msgs = append(msgs, msg)
		}
		if len(msgs) > 0 {
			return strings.Join(msgs, "; ")
		}
	}

	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return "Jira rejected the credentials, check your Jira settings"
	case http.StatusNotFound:
		return "Issue does not exist or you do not have permission to see it"
	}
	return http.StatusText(status)
}
