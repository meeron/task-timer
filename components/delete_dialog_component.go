package components

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type DeleteDialog struct {
	app.Compo

	TaskName  string
	OnConfirm func(ctx app.Context)
	OnCancel  func(ctx app.Context)
}

func (d *DeleteDialog) Render() app.UI {
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
							app.Div().Class("h-9 w-9 shrink-0 rounded-xl bg-red-50 text-red-600 flex items-center justify-center text-lg").
								Body(app.Span().Text("🗑️")),
							app.Div().Class("min-w-0").Body(
								app.H3().Class("text-sm font-bold text-slate-900").Text("Delete Task"),
								app.P().Class("text-xs text-slate-500 truncate").Text(d.TaskName),
							),
						),
						app.Button().
							Type("button").
							Class("text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer").
							Text("✕").
							OnClick(d.handleCancel),
					),

					// Message
					app.Div().Class("py-1").Body(
						app.P().Class("text-sm text-slate-600 leading-relaxed").
							Text("Are you sure you want to delete this task? This action cannot be undone."),
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
							Class("px-4 py-2 rounded-xl text-xs font-semibold text-white bg-red-600 hover:bg-red-700 active:scale-95 transition-all shadow-xs cursor-pointer").
							Text("Delete").
							OnClick(d.handleConfirm).
							AutoFocus(true),
					),
				),
		)
}

func (d *DeleteDialog) handleCancel(ctx app.Context, e app.Event) {
	if d.OnCancel != nil {
		d.OnCancel(ctx)
	}
}

func (d *DeleteDialog) handleConfirm(ctx app.Context, e app.Event) {
	if d.OnConfirm != nil {
		d.OnConfirm(ctx)
	}
}

func (d *DeleteDialog) onKeyDown(ctx app.Context, e app.Event) {
	key := e.Get("key").String()
	if key == "Enter" {
		d.handleConfirm(ctx, e)
	} else if key == "Escape" {
		d.handleCancel(ctx, e)
	}
}
