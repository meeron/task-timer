package pages

import (
	"fmt"
	"time"
	"uuid"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/components"
	"github.com/meeron/task-timer/models"
	"github.com/meeron/task-timer/pkg/indexeddb"
)

func (h *Home) OnMount(ctx app.Context) {
	ctx.Handle("deleteTask", h.onTaskDelete)
	db, err := indexeddb.Open("task_timer", 1, func(db indexeddb.IDBDatabase) {
		// Handle upgrade needed
		db.CreateObjectStore("tasks", "id")
	})
	if err != nil {
		app.Logf("Failed to open IndexedDB: %v", err)
		return
	}
	app.Logf("IndexedDB opened successfully: %v", db)
}

// OnAppUpdate satisfies the app.AppUpdater interface. It is called when the app
// is updated in background.
func (h *Home) OnAppUpdate(ctx app.Context) {
	h.updateAvailable = ctx.AppUpdateAvailable() // Reports that an app update is available.
}

func (h *Home) Render() app.UI {
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
			app.If(len(h.tasks) == 0, func() app.UI {
				return app.Div().Class("bg-white rounded-2xl border border-dashed border-slate-200 py-14 px-6 text-center").Body(
					app.Div().Class("text-4xl mb-3").Text("⏳"),
					app.H3().Class("text-base font-semibold text-slate-700").Text("No tasks yet"),
					app.P().Class("text-xs text-slate-400 mt-1 max-w-xs mx-auto").Text("Add your first task above and click 'Start task' to begin tracking time."),
				)
			}),

			app.If(len(h.tasks) > 0, func() app.UI {
				return app.Div().Class("space-y-3").Body(
					app.Range(h.tasks).Slice(func(i int) app.UI {
						task := h.tasks[i]
						return &components.Task{
							Id:   task.Id,
							Data: task,
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
	h.loadTasks()
}

func (h *Home) addNewTask(ctx app.Context, e app.Event) {
	if h.newTaskName == "" {
		return
	}

	newTask := models.Task{
		Id:        uuid.NewV4().String(),
		Name:      h.newTaskName,
		StartUnix: time.Now().Unix(),
	}

	objectStore := h.db.Call("transaction", "tasks", "readwrite").
		Call("objectStore", "tasks")
	objectStore.Call("add", map[string]interface{}{
		"id":        newTask.Id,
		"name":      newTask.Name,
		"startUnix": newTask.StartUnix,
		"duration":  newTask.Duration,
	})

	// Stop any currently running task before starting the new one.
	ctx.NewActionWithValue("stopOtherTasks", newTask.Id)

	h.tasks = append(h.tasks, newTask)

	h.newTaskName = ""
}

func (h *Home) loadTasks() {
	tasks := make([]models.Task, 0)

	objectStore := h.db.Call("transaction", "tasks").
		Call("objectStore", "tasks")
	cursor := objectStore.Call("openCursor")
	cursor.Set("onsuccess", app.FuncOf(func(this app.Value, args []app.Value) any {
		cr := args[0].Get("target").Get("result")
		if cr.IsNull() {
			// No more entries
			h.tasks = tasks
			return nil
		}

		value := cr.Get("value")
		task := models.Task{
			Id:        value.Get("id").String(),
			Name:      value.Get("name").String(),
			StartUnix: int64(value.Get("startUnix").Int()),
			Duration:  int64(value.Get("duration").Int()),
		}
		tasks = append(tasks, task)
		cr.Call("continue")
		return nil
	}))
}

type Home struct {
	app.Compo

	newTaskName     string
	updateAvailable bool
	tasks           []models.Task
	db              app.Value
}
