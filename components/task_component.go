package components

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/models"
)

func (t *Task) Render() app.UI {
	statusBorder := "border-l-emerald-500"
	timerColor := "text-emerald-600"
	if !t.isRunning {
		statusBorder = "border-l-slate-300"
		timerColor = "text-slate-600"
	}

	return app.Div().
		DataSet("id", t.Id).
		Class("bg-white rounded-xl border border-slate-200/80 border-l-4 shadow-xs hover:shadow-md transition-all p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 "+statusBorder).
		Body(
			// Left side: Status badge & Task name
			app.Div().Class("flex flex-col gap-1.5 min-w-0").Body(
				app.Div().Class("flex items-center gap-2").Body(
					app.If(t.isRunning, func() app.UI {
						return app.Span().Class("inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200").Body(
							app.Span().Class("relative flex h-2 w-2").Body(
								app.Span().Class("animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"),
								app.Span().Class("relative inline-flex rounded-full h-2 w-2 bg-emerald-500"),
							),
							app.Text("RUNNING"),
						)
					}),
					app.If(!t.isRunning, func() app.UI {
						return app.Span().Class("inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600 border border-slate-200").Body(
							app.Span().Class("inline-block h-2 w-2 rounded-full bg-slate-400"),
							app.Text("STOPPED"),
						)
					}),
				),
				app.H3().Class("text-base font-semibold text-slate-900 truncate").Text(t.Data.Name),
			),

			// Right side: Timer display & Action buttons
			app.Div().Class("flex flex-wrap items-center sm:justify-end gap-3 sm:gap-4").Body(
				app.Span().Class("font-mono text-xl font-bold tracking-tight px-3 py-1 bg-slate-50 rounded-lg border border-slate-200 "+timerColor).
					Text(formatDuration(t.duration)),

				app.Div().Class("flex items-center gap-1.5").Body(
					app.If(t.isRunning, func() app.UI {
						return app.Button().
							Class("inline-flex items-center px-3 py-1.5 rounded-lg text-xs font-semibold bg-amber-500 text-white hover:bg-amber-600 active:scale-95 transition-all shadow-xs cursor-pointer").
							Text("Stop").
							OnClick(t.onStop)
					}),
					app.If(!t.isRunning, func() app.UI {
						return app.Button().
							Class("inline-flex items-center px-3 py-1.5 rounded-lg text-xs font-semibold bg-emerald-600 text-white hover:bg-emerald-700 active:scale-95 transition-all shadow-xs cursor-pointer").
							Text("Resume").
							OnClick(t.onResume)
					}),
					app.Button().
						Class("inline-flex items-center px-2.5 py-1.5 rounded-lg text-xs font-medium bg-slate-100 text-slate-700 hover:bg-slate-200 active:scale-95 transition-all cursor-pointer").
						Text("Edit").
						OnClick(t.onEdit),
					app.Button().
						Class("inline-flex items-center px-2.5 py-1.5 rounded-lg text-xs font-medium text-red-600 hover:bg-red-50 hover:text-red-700 active:scale-95 transition-all cursor-pointer").
						Text("Delete").
						OnClick(t.onDelete),
				),
			),

			// Tailwind dialog for editing timer
			app.If(t.isEditing, func() app.UI {
				return &EditDialog{
					TaskName:     t.Data.Name,
					InitialValue: formatDurationTemplate(t.currentDuration()),
					OnSave:       t.onSaveEdit,
					OnCancel:     t.onCancelEdit,
				}
			}),

			// Tailwind dialog for confirming delete
			app.If(t.isDeleting, func() app.UI {
				return &DeleteDialog{
					TaskName:  t.Data.Name,
					OnConfirm: t.onConfirmDelete,
					OnCancel:  t.onCancelDelete,
				}
			}),
		)
}

func (t *Task) OnMount(ctx app.Context) {
	t.isRunning = t.Data.Duration == 0
	t.startUnix = t.Data.StartUnix
	t.ticker = time.NewTicker(1 * time.Second)
	t.ticker.Stop()

	ctx.Async(func() {
		for current := range t.ticker.C {
			ctx.Dispatch(func(c app.Context) {
				t.duration = time.Second * time.Duration(current.Unix()-t.startUnix)
			})
		}
	})

	if !t.isRunning {
		t.duration = time.Duration(t.Data.Duration)
		return
	}

	t.duration = time.Second * time.Duration(time.Now().Unix()-t.Data.StartUnix)
	t.ticker.Reset(1 * time.Second)
}

func (t *Task) onDelete(ctx app.Context, e app.Event) {
	t.isDeleting = true
}

func (t *Task) onCancelDelete(ctx app.Context) {
	t.isDeleting = false
}

func (t *Task) onConfirmDelete(ctx app.Context) {
	t.isDeleting = false
	ctx.NewActionWithValue("deleteTask", t.Id)
}

func (t *Task) onStop(ctx app.Context, e app.Event) {
	t.ticker.Stop()
	t.isRunning = false

	// t.Data is not updated
	t.Data.Duration = int64(t.duration)
	ctx.LocalStorage().Set(t.Id, t.Data)
}

func (t *Task) onResume(ctx app.Context, e app.Event) {
	t.isRunning = true
	t.startUnix = time.Now().Unix() - int64(t.duration.Seconds())

	t.Data.StartUnix = t.startUnix
	t.Data.Duration = 0
	ctx.LocalStorage().Set(t.Id, t.Data)

	t.ticker.Reset(1 * time.Second)
}

func (t *Task) currentDuration() time.Duration {
	if t.isRunning {
		return time.Second * time.Duration(time.Now().Unix()-t.startUnix)
	}
	return t.duration
}

func (t *Task) onEdit(ctx app.Context, e app.Event) {
	t.isEditing = true
}

func (t *Task) onCancelEdit(ctx app.Context) {
	t.isEditing = false
}

func (t *Task) onSaveEdit(ctx app.Context, newDuration time.Duration) {
	t.duration = newDuration
	if t.isRunning {
		t.startUnix = time.Now().Unix() - int64(newDuration.Seconds())
		t.Data.StartUnix = t.startUnix
		t.Data.Duration = 0
		ctx.LocalStorage().Set(t.Id, t.Data)
	} else {
		t.Data.Duration = int64(t.duration)
		ctx.LocalStorage().Set(t.Id, t.Data)
	}

	t.isEditing = false
}

func formatDuration(duration time.Duration) string {
	totalSeconds := int64(duration.Seconds())
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02dm %02ds", minutes, seconds)
}

type Task struct {
	app.Compo

	Id         string
	Data       models.Task
	ticker     *time.Ticker
	duration   time.Duration
	isRunning  bool
	startUnix  int64
	isEditing  bool
	isDeleting bool
}
