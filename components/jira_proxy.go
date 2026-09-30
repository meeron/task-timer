package components

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/pkg/jira"
)

// postJiraProxy posts body as JSON to one of the server's Jira proxy
// endpoints (jira.WorklogPath, jira.TestPath) and decodes a successful
// response into out when it is not nil. It blocks, so it must be called
// inside ctx.Async.
func postJiraProxy(path string, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	endpoint := app.Window().URL()
	endpoint.Path = path
	endpoint.RawQuery = ""
	endpoint.Fragment = ""

	resp, err := http.Post(endpoint.String(), "application/json", bytes.NewReader(payload))
	if err != nil {
		return errors.New("Could not reach the server")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if out != nil {
			return json.NewDecoder(resp.Body).Decode(out)
		}
		return nil
	}

	var errResp jira.ErrorResponse
	if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Error != "" {
		return errors.New(errResp.Error)
	}
	return errors.New("Jira request failed (HTTP " + resp.Status + ")")
}
