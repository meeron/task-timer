package components

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/pkg/jira"
)

// JiraSettingsDialog edits the user's Jira connection settings. OnSave gets
// the normalized config, or an empty config when the user disconnects.
type JiraSettingsDialog struct {
	app.Compo

	Config   jira.Config
	OnSave   func(ctx app.Context, cfg jira.Config)
	OnCancel func(ctx app.Context)

	baseURL  string
	email    string
	apiToken string
	errorMsg string

	testing    bool
	testResult string
}

func (d *JiraSettingsDialog) OnMount(ctx app.Context) {
	d.baseURL = d.Config.BaseURL
	d.email = d.Config.Email
	d.apiToken = d.Config.APIToken
	d.errorMsg = ""
	d.testing = false
	d.testResult = ""
}

func (d *JiraSettingsDialog) Render() app.UI {
	testText := "Test connection"
	if d.testing {
		testText = "Testing…"
	}
	inputClass := "w-full px-3.5 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-sm text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:bg-white transition-all"

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
								Body(app.Span().Text("⚙️")),
							app.Div().Class("min-w-0").Body(
								app.H3().Class("text-sm font-bold text-slate-900").Text("Jira settings"),
								app.P().Class("text-xs text-slate-500").Text("Log task time as Jira Cloud worklogs"),
							),
						),
						app.Button().
							Type("button").
							Class("text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer").
							Text("✕").
							OnClick(d.handleCancel),
					),

					// Base URL field
					app.Div().Class("space-y-2").Body(
						app.Label().Class("block text-xs font-semibold text-slate-700").Text("Jira URL"),
						app.Input().
							Class(inputClass).
							Type("url").
							Placeholder("https://your-org.atlassian.net").
							Value(d.baseURL).
							OnInput(d.onFieldInput(&d.baseURL)).
							OnChange(d.onFieldInput(&d.baseURL)),
					),

					// Email field
					app.Div().Class("space-y-2").Body(
						app.Label().Class("block text-xs font-semibold text-slate-700").Text("Atlassian account email"),
						app.Input().
							Class(inputClass).
							Type("email").
							Placeholder("you@example.com").
							Value(d.email).
							OnInput(d.onFieldInput(&d.email)).
							OnChange(d.onFieldInput(&d.email)),
					),

					// API token field
					app.Div().Class("space-y-2").Body(
						app.Label().Class("block text-xs font-semibold text-slate-700").Text("API token"),
						app.Input().
							Class(inputClass+" font-mono").
							Type("password").
							AutoComplete(false).
							Placeholder("Paste your API token").
							Value(d.apiToken).
							OnInput(d.onFieldInput(&d.apiToken)).
							OnChange(d.onFieldInput(&d.apiToken)),
						app.P().Class("text-xs text-slate-500").Body(
							app.A().
								Class("text-indigo-600 hover:underline").
								Href("https://id.atlassian.com/manage-profile/security/api-tokens").
								Target("_blank").
								Rel("noopener noreferrer").
								Text("Create an API token"),
							app.Text(". Stored only in this browser."),
						),
					),

					app.If(d.errorMsg != "", func() app.UI {
						return app.P().Class("text-xs text-red-700 bg-red-50 border border-red-200 rounded-xl px-3 py-2").Text(d.errorMsg)
					}),
					app.If(d.testResult != "", func() app.UI {
						return app.P().Class("text-xs text-emerald-700 bg-emerald-50 border border-emerald-200 rounded-xl px-3 py-2").Text(d.testResult)
					}),

					// Connection test
					app.Button().
						Type("button").
						Class("w-full px-3.5 py-2 rounded-xl text-xs font-semibold text-indigo-700 bg-indigo-50 hover:bg-indigo-100 active:scale-95 transition-all cursor-pointer disabled:opacity-50").
						Disabled(d.testing).
						Text(testText).
						OnClick(d.handleTest),

					// Action buttons
					app.Div().Class("flex items-center gap-2 pt-2 border-t border-slate-100").Body(
						app.If(d.Config.IsSet(), func() app.UI {
							return app.Button().
								Type("button").
								Class("px-3.5 py-2 rounded-xl text-xs font-semibold text-red-600 hover:text-red-700 hover:bg-red-50 active:scale-95 transition-all cursor-pointer").
								Text("Disconnect").
								OnClick(d.handleDisconnect)
						}),
						app.Div().Class("flex-1"),
						app.Button().
							Type("button").
							Class("px-3.5 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:text-slate-800 hover:bg-slate-100 active:scale-95 transition-all cursor-pointer").
							Text("Cancel").
							OnClick(d.handleCancel),
						app.Button().
							Type("button").
							Class("px-4 py-2 rounded-xl text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-700 active:scale-95 transition-all shadow-xs cursor-pointer").
							Text("Save").
							OnClick(d.handleSave),
					),
				),
		)
}

func (d *JiraSettingsDialog) handleCancel(ctx app.Context, e app.Event) {
	if d.OnCancel != nil {
		d.OnCancel(ctx)
	}
}

func (d *JiraSettingsDialog) handleSave(ctx app.Context, e app.Event) {
	cfg, err := jira.NormalizeConfig(jira.Config{BaseURL: d.baseURL, Email: d.email, APIToken: d.apiToken})
	if err != nil {
		d.errorMsg = err.Error()
		return
	}
	if d.OnSave != nil {
		d.OnSave(ctx, cfg)
	}
}

// handleTest checks the entered settings (saved or not) through the server's
// Jira proxy and shows who they authenticate as.
func (d *JiraSettingsDialog) handleTest(ctx app.Context, e app.Event) {
	if d.testing {
		return
	}

	d.errorMsg = ""
	d.testResult = ""

	cfg, err := jira.NormalizeConfig(jira.Config{BaseURL: d.baseURL, Email: d.email, APIToken: d.apiToken})
	if err != nil {
		d.errorMsg = err.Error()
		return
	}

	d.testing = true
	ctx.Async(func() {
		var resp jira.TestResponse
		err := postJiraProxy(jira.TestPath, cfg, &resp)
		ctx.Dispatch(func(c app.Context) {
			d.testing = false
			if err != nil {
				d.errorMsg = err.Error()
				return
			}
			d.testResult = "Connected as " + resp.DisplayName
		})
	})
}

// onFieldInput stores an input's value in field and clears a previous test
// result, which no longer applies to the edited settings.
func (d *JiraSettingsDialog) onFieldInput(field *string) app.EventHandler {
	return func(ctx app.Context, e app.Event) {
		*field = ctx.JSSrc().Get("value").String()
		d.testResult = ""
	}
}

func (d *JiraSettingsDialog) handleDisconnect(ctx app.Context, e app.Event) {
	if d.OnSave != nil {
		d.OnSave(ctx, jira.Config{})
	}
}

func (d *JiraSettingsDialog) onKeyDown(ctx app.Context, e app.Event) {
	switch e.Get("key").String() {
	case "Enter":
		d.handleSave(ctx, e)
	case "Escape":
		d.handleCancel(ctx, e)
	}
}
