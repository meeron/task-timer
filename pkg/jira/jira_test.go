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

func TestNormalizeConfig(t *testing.T) {
	valid := Config{BaseURL: " https://Your-Org.atlassian.net/jira/ ", Email: " me@example.com ", APIToken: " secret "}
	cfg, err := NormalizeConfig(valid)
	if err != nil {
		t.Fatalf("valid config: unexpected error %v", err)
	}
	expected := Config{BaseURL: "https://your-org.atlassian.net", Email: "me@example.com", APIToken: "secret"}
	if cfg != expected {
		t.Errorf("NormalizeConfig = %+v, expected %+v", cfg, expected)
	}

	for _, baseURL := range []string{
		"",
		"your-org.atlassian.net",
		"http://your-org.atlassian.net",
		"https://atlassian.net",
		"https://a.b.atlassian.net",
		"https://your-org.atlassian.net.evil.com",
		"https://your-org.atlassian.net:8443",
		"https://user@your-org.atlassian.net",
		"https://localhost",
	} {
		if _, err := NormalizeConfig(Config{BaseURL: baseURL, Email: "e", APIToken: "t"}); err == nil {
			t.Errorf("NormalizeConfig(BaseURL %q): expected error", baseURL)
		}
	}

	if _, err := NormalizeConfig(Config{BaseURL: "https://x.jira.com", Email: "e", APIToken: "t"}); err != nil {
		t.Errorf("jira.com host: unexpected error %v", err)
	}
	if _, err := NormalizeConfig(Config{BaseURL: "https://x.atlassian.net", Email: " ", APIToken: "t"}); err == nil {
		t.Error("missing email: expected error")
	}
	if _, err := NormalizeConfig(Config{BaseURL: "https://x.atlassian.net", Email: "e"}); err == nil {
		t.Error("missing token: expected error")
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

	client := NewClient()
	cfg := Config{BaseURL: srv.URL, Email: "me@example.com", APIToken: "secret"}
	started := time.Date(2026, 9, 29, 9, 0, 0, 0, time.UTC)
	id, err := client.AddWorklog(context.Background(), cfg, WorklogRequest{
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

	client := NewClient()
	_, err := client.AddWorklog(context.Background(), Config{BaseURL: srv.URL, Email: "e", APIToken: "t"}, WorklogRequest{IssueKey: "ABC-1", StartedUnix: 1, TimeSpentSeconds: 60})

	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Message != "Issue does not exist" {
		t.Errorf("got %+v", apiErr)
	}
}

type fakeWorklogger struct {
	gotCfg Config
	got    WorklogRequest
	err    error
}

func (f *fakeWorklogger) AddWorklog(_ context.Context, cfg Config, req WorklogRequest) (string, error) {
	f.gotCfg = cfg
	f.got = req
	return "42", f.err
}

func TestWorklogHandler(t *testing.T) {
	const cfg = `"config":{"baseUrl":"https://x.atlassian.net/","email":"me@example.com","apiToken":"secret"}`
	body := func(worklog string) string { return `{` + cfg + `,"worklog":` + worklog + `}` }

	tests := []struct {
		name       string
		body       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"ok, key normalized", body(`{"issueKey":" abc-1 ","startedUnix":1,"timeSpentSeconds":60}`), nil, 201, `{"id":"42"}`},
		{"bad json", `{`, nil, 400, `{"error":"Invalid request body"}`},
		{"missing config", `{"worklog":{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":60}}`, nil, 400, `{"error":"Jira URL must look like https://your-org.atlassian.net"}`},
		{"non-jira host", `{"config":{"baseUrl":"https://example.com","email":"e","apiToken":"t"},"worklog":{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":60}}`, nil, 400, `{"error":"Jira URL must look like https://your-org.atlassian.net"}`},
		{"bad key", body(`{"issueKey":"nope","startedUnix":1,"timeSpentSeconds":60}`), nil, 400, `{"error":"Invalid Jira issue key"}`},
		{"too short", body(`{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":59}`), nil, 400, `{"error":"Time spent must be at least 1 minute"}`},
		{"jira error", body(`{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":60}`), &APIError{404, "Issue does not exist"}, 502, `{"error":"Issue does not exist"}`},
		{"network error", body(`{"issueKey":"ABC-1","startedUnix":1,"timeSpentSeconds":60}`), errors.New("dial tcp"), 502, `{"error":"Could not reach Jira"}`},
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
			if tc.wantStatus == 201 {
				if fake.got.IssueKey != "ABC-1" {
					t.Errorf("issue key = %q, expected normalized ABC-1", fake.got.IssueKey)
				}
				if fake.gotCfg.BaseURL != "https://x.atlassian.net" || fake.gotCfg.APIToken != "secret" {
					t.Errorf("config = %+v, expected normalized request config", fake.gotCfg)
				}
			}
		})
	}
}

func TestClientMyself(t *testing.T) {
	var gotMethod, gotPath, gotUser string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotUser, _, _ = r.BasicAuth()
		io.WriteString(w, `{"accountId":"abc","displayName":"Jane Doe"}`)
	}))
	defer srv.Close()

	name, err := NewClient().Myself(context.Background(), Config{BaseURL: srv.URL, Email: "me@example.com", APIToken: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Jane Doe" {
		t.Errorf("name = %q, expected Jane Doe", name)
	}
	if gotMethod != http.MethodGet || gotPath != "/rest/api/3/myself" || gotUser != "me@example.com" {
		t.Errorf("request = %s %s as %q", gotMethod, gotPath, gotUser)
	}
}

func TestClientMyselfErrors(t *testing.T) {
	tests := []struct {
		status  int
		body    string
		wantMsg string
	}{
		{http.StatusUnauthorized, ``, "Jira rejected the credentials, check your Jira settings"},
		{http.StatusNotFound, `<html>not found</html>`, "Jira site not found, check the Jira URL"},
	}

	for _, tc := range tests {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			io.WriteString(w, tc.body)
		}))

		_, err := NewClient().Myself(context.Background(), Config{BaseURL: srv.URL, Email: "e", APIToken: "t"})
		srv.Close()

		apiErr, ok := errors.AsType[*APIError](err)
		if !ok {
			t.Fatalf("HTTP %d: expected *APIError, got %v", tc.status, err)
		}
		if apiErr.Message != tc.wantMsg {
			t.Errorf("HTTP %d: message = %q, expected %q", tc.status, apiErr.Message, tc.wantMsg)
		}
	}
}

type fakeConnectionTester struct {
	gotCfg Config
	err    error
}

func (f *fakeConnectionTester) Myself(_ context.Context, cfg Config) (string, error) {
	f.gotCfg = cfg
	return "Jane Doe", f.err
}

func TestTestHandler(t *testing.T) {
	const cfg = `{"baseUrl":"https://x.atlassian.net/","email":"me@example.com","apiToken":"secret"}`

	tests := []struct {
		name       string
		body       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"ok", cfg, nil, 200, `{"displayName":"Jane Doe"}`},
		{"bad json", `{`, nil, 400, `{"error":"Invalid request body"}`},
		{"non-jira host", `{"baseUrl":"https://example.com","email":"e","apiToken":"t"}`, nil, 400, `{"error":"Jira URL must look like https://your-org.atlassian.net"}`},
		{"jira error", cfg, &APIError{401, "Jira rejected the credentials, check your Jira settings"}, 502, `{"error":"Jira rejected the credentials, check your Jira settings"}`},
		{"network error", cfg, errors.New("dial tcp"), 502, `{"error":"Could not reach Jira"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeConnectionTester{err: tc.err}
			rec := httptest.NewRecorder()
			NewTestHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, TestPath, strings.NewReader(tc.body)))

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, expected %d", rec.Code, tc.wantStatus)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tc.wantBody {
				t.Errorf("body = %s, expected %s", got, tc.wantBody)
			}
			if tc.wantStatus == 200 && fake.gotCfg.BaseURL != "https://x.atlassian.net" {
				t.Errorf("config = %+v, expected normalized request config", fake.gotCfg)
			}
		})
	}
}
