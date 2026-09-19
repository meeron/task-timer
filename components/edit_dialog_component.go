package components

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type EditDialog struct {
	app.Compo

	TaskName     string
	InitialValue string
	OnSave       func(ctx app.Context, d time.Duration)
	OnCancel     func(ctx app.Context)

	input    string
	errorMsg string
}

func (d *EditDialog) OnMount(ctx app.Context) {
	d.input = d.InitialValue
	d.errorMsg = ""
}

func (d *EditDialog) Render() app.UI {
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
								Body(app.Span().Text("⏱️")),
							app.Div().Class("min-w-0").Body(
								app.H3().Class("text-sm font-bold text-slate-900").Text("Edit Timer"),
								app.P().Class("text-xs text-slate-500 truncate").Text(d.TaskName),
							),
						),
						app.Button().
							Type("button").
							Class("text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer").
							Text("✕").
							OnClick(d.handleCancel),
					),

					// Input field
					app.Div().Class("space-y-2").Body(
						app.Label().Class("block text-xs font-semibold text-slate-700").Text("Timer duration (template: 1h 15m)"),
						app.Input().
							Class("w-full px-3.5 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-sm font-mono text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:bg-white transition-all").
							Type("text").
							Placeholder("e.g. 1h 15m").
							Value(d.input).
							OnInput(d.ValueTo(&d.input)).
							OnChange(d.ValueTo(&d.input)).
							OnKeyDown(d.onKeyDown).
							AutoFocus(true),
						app.If(d.errorMsg != "", func() app.UI {
							return app.P().Class("text-xs text-red-600 font-medium").Text(d.errorMsg)
						}),
						// Quick presets
						app.Div().Class("flex flex-wrap items-center gap-1.5 pt-1").Body(
							app.Span().Class("text-[11px] text-slate-400 font-medium mr-1").Text("Presets:"),
							app.Range([]string{"15m", "30m", "45m", "1h", "1h 15m", "2h"}).Slice(func(i int) app.UI {
								preset := []string{"15m", "30m", "45m", "1h", "1h 15m", "2h"}[i]
								return app.Button().
									Type("button").
									Class("text-xs px-2 py-0.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-600 font-medium transition-colors cursor-pointer").
									Text(preset).
									OnClick(func(ctx app.Context, e app.Event) {
										d.input = preset
										d.errorMsg = ""
									})
							}),
						),
					),

					// Action buttons
					app.Div().Class("flex items-center justify-end gap-2 pt-2 border-t border-slate-100").Body(
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

func (d *EditDialog) handleCancel(ctx app.Context, e app.Event) {
	if d.OnCancel != nil {
		d.OnCancel(ctx)
	}
}

func (d *EditDialog) handleSave(ctx app.Context, e app.Event) {
	input := strings.TrimSpace(d.input)
	if input == "" {
		d.errorMsg = "Please enter a duration (e.g. 1h 15m)"
		return
	}

	newDuration, err := parseDurationInput(input)
	if err != nil {
		d.errorMsg = "Invalid time format. Example: 1h 15m"
		return
	}

	if d.OnSave != nil {
		d.OnSave(ctx, newDuration)
	}
}

func (d *EditDialog) onKeyDown(ctx app.Context, e app.Event) {
	key := e.Get("key").String()
	if key == "Enter" {
		d.handleSave(ctx, e)
	} else if key == "Escape" {
		d.handleCancel(ctx, e)
	}
}

func formatDurationTemplate(duration time.Duration) string {
	totalSeconds := int64(duration.Seconds())
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func parseDurationInput(input string) (time.Duration, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		if len(parts) == 2 {
			h, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			m, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && h >= 0 && m >= 0 {
				return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute, nil
			}
		} else if len(parts) == 3 {
			h, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			m, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			s, err3 := strconv.Atoi(strings.TrimSpace(parts[2]))
			if err1 == nil && err2 == nil && err3 == nil && h >= 0 && m >= 0 && s >= 0 {
				return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second, nil
			}
		}
	}

	cleaned := strings.ReplaceAll(input, "hours", "h")
	cleaned = strings.ReplaceAll(cleaned, "hour", "h")
	cleaned = strings.ReplaceAll(cleaned, "hrs", "h")
	cleaned = strings.ReplaceAll(cleaned, "hr", "h")
	cleaned = strings.ReplaceAll(cleaned, "minutes", "m")
	cleaned = strings.ReplaceAll(cleaned, "minute", "m")
	cleaned = strings.ReplaceAll(cleaned, "mins", "m")
	cleaned = strings.ReplaceAll(cleaned, "min", "m")
	cleaned = strings.ReplaceAll(cleaned, "seconds", "s")
	cleaned = strings.ReplaceAll(cleaned, "second", "s")
	cleaned = strings.ReplaceAll(cleaned, "secs", "s")
	cleaned = strings.ReplaceAll(cleaned, "sec", "s")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	if num, err := strconv.Atoi(cleaned); err == nil {
		if num < 0 {
			return 0, fmt.Errorf("duration cannot be negative")
		}
		return time.Duration(num) * time.Minute, nil
	}

	d, err := time.ParseDuration(cleaned)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("duration cannot be negative")
	}
	return d, nil
}
