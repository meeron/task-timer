package pages

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"uuid"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/components"
	"github.com/meeron/task-timer/models"
)

func (h *Home) OnMount(ctx app.Context) {
	h.loadTasks(ctx.LocalStorage())
	ctx.Handle("deleteTask", h.onTaskDelete)
}

// OnAppUpdate satisfies the app.AppUpdater interface. It is called when the app
// is updated in background.
func (h *Home) OnAppUpdate(ctx app.Context) {
	h.updateAvailable = ctx.AppUpdateAvailable() // Reports that an app update is available.
}

func (h *Home) Render() app.UI {
	taskIDs := h.sortedTaskIDs()

	return app.Main().Class("min-h-screen bg-slate-50 text-slate-800 py-10 px-4 sm:px-6 antialiased").Body(
		app.Div().Class("max-w-2xl mx-auto space-y-6").Body(
			// App update banner
			app.If(h.updateAvailable, func() app.UI {
				return app.Div().Class("bg-indigo-600 text-white px-4 py-3 rounded-2xl shadow-sm flex items-center justify-between").Body(
					app.Div().Class("flex items-center gap-2 text-sm font-medium").Body(
						app.Span().Text("✨ An update is available!"),
					),
					app.Button().
						Class("bg-white text-indigo-700 hover:bg-indigo-50 font-semibold px-3.5 py-1.5 rounded-xl text-xs shadow-xs transition-all cursor-pointer").
						Text("Update now").
						OnClick(h.onUpdateClick),
				)
			}),

			// Header
			app.Header().Class("flex items-center justify-between pb-1").Body(
				app.Div().Class("flex items-center gap-3").Body(
					app.Img().Src("/web/logo.png").Alt("logo"),
					app.Div().Body(
						app.H1().Class("text-2xl font-bold tracking-tight text-slate-900").Text("Task Timer"),
						app.P().Class("text-xs text-slate-500 font-medium").Text("Focus and measure your time effortlessly"),
					),
				),
				app.Span().Class("text-xs font-semibold px-3 py-1 rounded-full bg-slate-200/80 text-slate-700").
					Text(fmt.Sprintf("%d tasks", len(h.tasks))),
			),

			// Task input card
			app.Div().Class("bg-white p-2.5 rounded-2xl border border-slate-200 shadow-xs flex flex-col sm:flex-row gap-2").Body(
				app.Input().
					Class("flex-1 px-4 py-2.5 rounded-xl bg-slate-50/70 border border-slate-200 text-sm text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:bg-white transition-all").
					Type("text").
					Placeholder("What are you working on?").
					Value(h.newTaskName).
					OnChange(h.ValueTo(&h.newTaskName)).
					OnKeyDown(h.onInputKeyDown),
				app.Button().
					Class("inline-flex items-center justify-center gap-1.5 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white text-sm font-semibold shadow-xs transition-all cursor-pointer disabled:opacity-50").
					Disabled(h.newTaskName == "").
					OnClick(h.addNewTask).
					Text("Start task"),
			),

			// Tasks list or Empty state
			app.If(len(taskIDs) == 0, func() app.UI {
				return app.Div().Class("bg-white rounded-2xl border border-dashed border-slate-200 py-14 px-6 text-center").Body(
					app.Div().Class("text-4xl mb-3").Text("⏳"),
					app.H3().Class("text-base font-semibold text-slate-700").Text("No tasks yet"),
					app.P().Class("text-xs text-slate-400 mt-1 max-w-xs mx-auto").Text("Add your first task above and click 'Start task' to begin tracking time."),
				)
			}),

			app.If(len(taskIDs) > 0, func() app.UI {
				return app.Div().Class("space-y-3").Body(
					app.Range(taskIDs).Slice(func(i int) app.UI {
						id := taskIDs[i]
						return &components.Task{
							Id:   id,
							Data: h.tasks[id],
						}
					}),
				)
			}),
		),
	)
}

func (h *Home) onInputKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Enter" {
		h.addNewTask(ctx, e)
	}
}

func (h *Home) onUpdateClick(ctx app.Context, e app.Event) {
	// Reloads the page to display the modifications.
	ctx.Reload()
}

func (h *Home) onTaskDelete(ctx app.Context, a app.Action) {
	taskId := a.Value.(string)
	ctx.LocalStorage().Del(taskId)
	delete(h.tasks, taskId)
}

func (h *Home) addNewTask(ctx app.Context, e app.Event) {
	if h.newTaskName == "" {
		return
	}

	taskId := "_task_" + uuid.NewV7().String()

	// Stop any currently running task before starting the new one.
	ctx.NewActionWithValue("stopOtherTasks", taskId)

	h.tasks[taskId] = models.Task{
		Name:      h.newTaskName,
		StartUnix: time.Now().Unix(),
	}

	err := ctx.LocalStorage().Set(taskId, h.tasks[taskId])
	if err != nil {
		app.Log(err)
	}

	h.newTaskName = ""
}

func (h *Home) loadTasks(storage app.BrowserStorage) {
	h.tasks = make(map[string]models.Task)
	storage.ForEach(func(key string) {
		if !strings.HasPrefix(key, "_task_") {
			return
		}

		var data models.Task
		err := storage.Get(key, &data)
		if err != nil {
			app.Log(err)
			return
		}

		h.tasks[key] = data
	})
}

func (h *Home) sortedTaskIDs() []string {
	keys := make([]string, 0, len(h.tasks))
	for k := range h.tasks {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if h.tasks[keys[i]].StartUnix != h.tasks[keys[j]].StartUnix {
			return h.tasks[keys[i]].StartUnix < h.tasks[keys[j]].StartUnix
		}
		return keys[i] < keys[j]
	})
	return keys
}

type Home struct {
	app.Compo

	newTaskName     string
	updateAvailable bool
	tasks           map[string]models.Task
}
