package components

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/pkg/jira"
)

type WorklogDialog struct {
	app.Compo

	TaskName        string
	TaskDescription string
	Duration        time.Duration
	LoggedUnix      int64
	OnLogged        func(ctx app.Context)
	OnCancel        func(ctx app.Context)

	issueKey    string
	keyErrorMsg string
	description string
	input       string
	errorMsg    string
	submitErr   string
	submitting  bool
}

func (d *WorklogDialog) OnMount(ctx app.Context) {
	d.issueKey = strings.ToUpper(strings.TrimSpace(d.TaskName))
	d.keyErrorMsg = ""
	if !jira.IsValidIssueKey(d.issueKey) {
		d.keyErrorMsg = "Task name is not a Jira issue key (e.g. ABC-123)"
	}
	d.description = d.TaskDescription
	d.input = formatDurationTemplate(d.Duration)
	d.errorMsg = ""
	d.submitErr = ""
	d.submitting = false
}

func (d *WorklogDialog) Render() app.UI {
	submitText := "Log work"
	if d.submitting {
		submitText = "Logging…"
	}

	return app.Dialog().
		Attr("open", true).
		Class("fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-xs border-0 outline-none w-full h-full max-w-none max-h-none m-0").
		OnKeyDown(d.onKeyDown).
		Body(
			// Backdrop overlay (clicking outside closes dialog)
			app.Div().
				Class("fixed inset-0 -z-10").
				OnClick(d.handleCancel),

			// Dialog Card
			app.Div().
				Class("relative bg-white rounded-2xl shadow-2xl border border-slate-200/80 max-w-sm w-full p-5 space-y-4 text-left").
				Body(
					// Header
					app.Div().Class("flex items-start justify-between gap-3").Body(
						app.Div().Class("flex items-center gap-2.5 min-w-0").Body(
							app.Div().Class("h-9 w-9 shrink-0 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center text-lg").
								Body(app.Span().Text("📝")),
							app.Div().Class("min-w-0").Body(
								app.H3().Class("text-sm font-bold text-slate-900").Text("Add Jira worklog"),
								app.P().Class("text-xs text-slate-500 truncate").Text(d.TaskName),
							),
						),
						app.Button().
							Type("button").
							Class("text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer").
							Text("✕").
							OnClick(d.handleCancel),
					),

					// Already-logged notice
					app.If(d.LoggedUnix != 0, func() app.UI {
						return app.P().Class("text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded-xl px-3 py-2").
							Text("Already logged on " + time.Unix(d.LoggedUnix, 0).Format("Jan 2, 15:04") + ". Logging again adds another worklog.")
					}),

					// Issue key field
					app.Div().Class("space-y-2").Body(
						app.Label().Class("block text-xs font-semibold text-slate-700").Text("Jira issue"),
						app.Input().
							Class("w-full px-3.5 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-sm font-mono text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:bg-white transition-all").
							Type("text").
							Placeholder("e.g. ABC-123").
							Value(d.issueKey).
							OnInput(d.ValueTo(&d.issueKey)).
							OnChange(d.ValueTo(&d.issueKey)).
							OnKeyDown(d.onKeyDown),
						app.If(d.keyErrorMsg != "", func() app.UI {
							return app.P().Class("text-xs text-red-600 font-medium").Text(d.keyErrorMsg)
						}),
					),

					// Description field
					app.Div().Class("space-y-2").Body(
						app.Label().Class("block text-xs font-semibold text-slate-700").Text("Work description (optional)"),
						app.Textarea().
							Class("w-full px-3.5 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-sm text-slate-800 placeholder-slate-400 resize-y focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:bg-white transition-all").
							Rows(3).
							Placeholder("What did you work on?").
							Attr("value", d.description).
							OnInput(d.ValueTo(&d.description)).
							OnChange(d.ValueTo(&d.description)),
					),

					// Time spent field
					app.Div().Class("space-y-2").Body(
						app.Label().Class("block text-xs font-semibold text-slate-700").Text("Time spent (template: 1h 15m)"),
						app.Input().
							Class("w-full px-3.5 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-sm font-mono text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:bg-white transition-all").
							Type("text").
							Placeholder("e.g. 1h 15m").
							Value(d.input).
							OnInput(d.ValueTo(&d.input)).
							OnChange(d.ValueTo(&d.input)).
							OnKeyDown(d.onKeyDown),
						app.If(d.errorMsg != "", func() app.UI {
							return app.P().Class("text-xs text-red-600 font-medium").Text(d.errorMsg)
						}),
					),

					app.If(d.submitErr != "", func() app.UI {
						return app.P().Class("text-xs text-red-700 bg-red-50 border border-red-200 rounded-xl px-3 py-2").Text(d.submitErr)
					}),

					// Action buttons
					app.Div().Class("flex items-center justify-end gap-2 pt-2 border-t border-slate-100").Body(
						app.Button().
							Type("button").
							Class("px-3.5 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:text-slate-800 hover:bg-slate-100 active:scale-95 transition-all cursor-pointer disabled:opacity-50").
							Disabled(d.submitting).
							Text("Cancel").
							OnClick(d.handleCancel),
						app.Button().
							Type("button").
							Class("px-4 py-2 rounded-xl text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-700 active:scale-95 transition-all shadow-xs cursor-pointer disabled:opacity-50").
							Disabled(d.submitting).
							Text(submitText).
							OnClick(d.handleSubmit),
					),
				),
		)
}

func (d *WorklogDialog) handleCancel(ctx app.Context, e app.Event) {
	if d.submitting {
		return
	}
	if d.OnCancel != nil {
		d.OnCancel(ctx)
	}
}

func (d *WorklogDialog) handleSubmit(ctx app.Context, e app.Event) {
	if d.submitting {
		return
	}

	d.keyErrorMsg = ""
	d.errorMsg = ""
	d.submitErr = ""

	issueKey := strings.ToUpper(strings.TrimSpace(d.issueKey))
	if !jira.IsValidIssueKey(issueKey) {
		d.keyErrorMsg = "Enter a Jira issue key (e.g. ABC-123)"
	}

	var spent int64
	input := strings.TrimSpace(d.input)
	if input == "" {
		d.errorMsg = "Please enter the time spent (e.g. 1h 15m)"
	} else if parsed, err := parseDurationInput(input); err != nil {
		d.errorMsg = "Invalid time format. Example: 1h 15m"
	} else if spent = worklogSeconds(parsed); spent < jira.MinTimeSpent {
		d.errorMsg = "Jira requires at least 1 minute"
	}

	if d.keyErrorMsg != "" || d.errorMsg != "" {
		return
	}

	req := jira.WorklogRequest{
		IssueKey:         issueKey,
		Comment:          strings.TrimSpace(d.description),
		StartedUnix:      time.Now().Unix() - spent,
		TimeSpentSeconds: spent,
	}

	d.submitting = true
	ctx.Async(func() {
		err := postWorklog(req)
		ctx.Dispatch(func(c app.Context) {
			d.submitting = false
			if err != nil {
				d.submitErr = err.Error()
				return
			}
			if d.OnLogged != nil {
				d.OnLogged(c)
			}
		})
	})
}

func (d *WorklogDialog) onKeyDown(ctx app.Context, e app.Event) {
	key := e.Get("key").String()
	// Enter inserts a newline in the description textarea instead of submitting.
	if key == "Enter" && e.Get("target").Get("tagName").String() != "TEXTAREA" {
		d.handleSubmit(ctx, e)
	} else if key == "Escape" {
		d.handleCancel(ctx, e)
	}
}

// worklogSeconds rounds a duration to whole minutes, as Jira tracks worklogs
// with minute precision.
func worklogSeconds(d time.Duration) int64 {
	return int64(d.Round(time.Minute).Seconds())
}

// postWorklog sends the worklog to the server's Jira proxy. It blocks, so it
// must be called inside ctx.Async.
func postWorklog(req jira.WorklogRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	endpoint := app.Window().URL()
	endpoint.Path = jira.WorklogPath
	endpoint.RawQuery = ""
	endpoint.Fragment = ""

	resp, err := http.Post(endpoint.String(), "application/json", bytes.NewReader(body))
	if err != nil {
		return errors.New("Could not reach the server")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	var errResp jira.ErrorResponse
	if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Error != "" {
		return errors.New(errResp.Error)
	}
	return errors.New("Failed to add worklog (HTTP " + resp.Status + ")")
}
