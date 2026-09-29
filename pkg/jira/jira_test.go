package jira

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIsValidIssueKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"ABC-123", true},
		{"AB-1", true},
		{"A1_B-42", true},
		{"abc-123", false},
		{"ABC", false},
		{"ABC-", false},
		{"ABC-0", false},
		{"1AB-5", false},
		{"A-1", false},
		{"ABC-12 ", false},
		{"Fix login bug", false},
	}

	for _, tc := range tests {
		if got := IsValidIssueKey(tc.key); got != tc.expected {
			t.Errorf("IsValidIssueKey(%q) = %v, expected %v", tc.key, got, tc.expected)
		}
	}
}

func TestConfigFromEnv(t *testing.T) {
	env := func(vars map[string]string) func(string) string {
		return func(k string) string { return vars[k] }
	}

	if _, err := ConfigFromEnv(env(nil)); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("empty env: expected ErrNotConfigured, got %v", err)
	}

	_, err := ConfigFromEnv(env(map[string]string{"JIRA_BASE_URL": "https://x.atlassian.net"}))
	if err == nil || errors.Is(err, ErrNotConfigured) || !strings.Contains(err.Error(), "JIRA_EMAIL, JIRA_API_TOKEN") {
		t.Errorf("partial env: expected missing-vars error, got %v", err)
	}

	cfg, err := ConfigFromEnv(env(map[string]string{
		"JIRA_BASE_URL":  "https://x.atlassian.net/",
		"JIRA_EMAIL":     "me@example.com",
		"JIRA_API_TOKEN": "secret",
	}))
	if err != nil {
		t.Fatalf("full env: unexpected error %v", err)
	}
	if cfg.BaseURL != "https://x.atlassian.net" {
		t.Errorf("BaseURL = %q, expected trailing slash trimmed", cfg.BaseURL)
	}
}

func TestToADF(t *testing.T) {
	got, _ := json.Marshal(toADF("first\n\nsecond"))
	expected := `{"content":[` +
		`{"content":[{"text":"first","type":"text"}],"type":"paragraph"},` +
		`{"type":"paragraph"},` +
		`{"content":[{"text":"second","type":"text"}],"type":"paragraph"}` +
		`],"type":"doc","version":1}`
	if string(got) != expected {
		t.Errorf("toADF =\n%s\nexpected\n%s", got, expected)
	}
}

func TestClientAddWorklog(t *testing.T) {
	var gotPath, gotUser, gotPass string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotUser, gotPass, _ = r.BasicAuth()
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"id":"10001"}`)
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL, Email: "me@example.com", APIToken: "secret"})
	started := time.Date(2026, 9, 29, 9, 0, 0, 0, time.UTC)
	id, err := client.AddWorklog(context.Background(), WorklogRequest{
		IssueKey:         "ABC-123",
		Comment:          "Did things",
		StartedUnix:      started.Unix(),
		TimeSpentSeconds: 4500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "10001" {
		t.Errorf("id = %q, expected 10001", id)
	}
	if gotPath != "/rest/api/3/issue/ABC-123/worklog" {
		t.Errorf("path = %q", gotPath)
	}
	if gotUser != "me@example.com" || gotPass != "secret" {
		t.Errorf("basic auth = %q:%q", gotUser, gotPass)
	}
	if gotBody["timeSpentSeconds"] != float64(4500) {
		t.Errorf("timeSpentSeconds = %v", gotBody["timeSpentSeconds"])
	}
	if parsed, err := time.Parse(startedLayout, gotBody["started"].(string)); err != nil || !parsed.Equal(started) {
		t.Errorf("started = %v (err %v), expected %v", gotBody["started"], err, started)
	}
	if gotBody["comment"] == nil {
		t.Error("comment missing")
	}
}

func TestClientAddWorklogError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"errorMessages":["Issue does not exist"],"errors":{}}`)
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL, Email: "e", APIToken: "t"})
	_, err := client.AddWorklog(context.Background(), WorklogRequest{IssueKey: "ABC-1", StartedUnix: 1, TimeSpentSeconds: 60})

	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Message != "Issue does not exist" {
		t.Errorf("got %+v", apiErr)
	}
}

type fakeWorklogger struct {
	got WorklogRequest
	err error
}

func (f *fakeWorklogger) AddWorklog(_ context.Context, req WorklogRequest) (string, error) {
	f.got = req
	return "42", f.err
}

func TestWorklogHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"ok, key normalized", `{"issueKey":" abc-1 ","startedUnix":1,"timeSpentSeconds":60}`, nil, 201, `{"id":"42"}`},
		{"bad json", `{`, nil, 400, `{"error":"Invalid request body"}`},
		{"bad key", `{"issueKey":"nope","startedUnix":1,"timeSpentSeconds":60}`, nil, 400, `{"error":"Invalid Jira issue key"}`},
		{"too short", `{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":59}`, nil, 400, `{"error":"Time spent must be at least 1 minute"}`},
		{"jira error", `{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":60}`, &APIError{404, "Issue does not exist"}, 502, `{"error":"Issue does not exist"}`},
		{"network error", `{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":60}`, errors.New("dial tcp"), 502, `{"error":"Could not reach Jira"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeWorklogger{err: tc.err}
			rec := httptest.NewRecorder()
			NewWorklogHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, WorklogPath, strings.NewReader(tc.body)))

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, expected %d", rec.Code, tc.wantStatus)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tc.wantBody {
				t.Errorf("body = %s, expected %s", got, tc.wantBody)
			}
			if tc.wantStatus == 201 && fake.got.IssueKey != "ABC-1" {
				t.Errorf("issue key = %q, expected normalized ABC-1", fake.got.IssueKey)
			}
		})
	}
}
