package pages

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"uuid"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/components"
	"github.com/meeron/task-timer/models"
	"github.com/meeron/task-timer/pkg/indexeddb"
	"github.com/meeron/task-timer/pkg/jira"
)

func (h *Home) OnMount(ctx app.Context) {
	ctx.Handle("deleteTask", h.onTaskDelete)
	ctx.Handle("saveTask", h.onTaskSave)

	ctx.Async(func() {
		db, err := indexeddb.Open("task_timer", 2, func(db indexeddb.IDBDatabase) {
			// Handle upgrade needed — create missing object stores on first run / version bump.
			if !db.HasObjectStore("tasks") {
				db.CreateObjectStore("tasks", "id")
			}
			if !db.HasObjectStore("settings") {
				db.CreateObjectStore("settings", "id")
			}
		})
		if err != nil {
			app.Logf("Failed to open IndexedDB: %v", err)
			return
		}

		ctx.Dispatch(func(c app.Context) {
			h.db = db
			h.loadTasks(c)
			h.loadJiraConfig(c)
		})
	})
}

// OnAppUpdate satisfies the app.AppUpdater interface. It is called when the app
// is updated in background.
func (h *Home) OnAppUpdate(ctx app.Context) {
	h.updateAvailable = ctx.AppUpdateAvailable() // Reports that an app update is available.
}

func (h *Home) Render() app.UI {
	jiraStatus, jiraStatusColor := "Not connected", "bg-slate-300"
	if h.jiraConfig.IsSet() {
		jiraStatus, jiraStatusColor = "Connected to "+h.jiraConfig.BaseURL, "bg-emerald-500"
	}

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
				app.Div().Class("flex items-center gap-2").Body(
					app.Button().
						Class("inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 active:scale-95 transition-all cursor-pointer").
						Title("Jira settings").
						OnClick(h.onJiraSettingsClick).
						Body(
							app.Span().
								Class("inline-block h-2 w-2 rounded-full "+jiraStatusColor).
								Title(jiraStatus),
							app.Text("Jira"),
						),
					app.Span().Class("text-xs font-semibold px-3 py-1 rounded-full bg-slate-200/80 text-slate-700").
						Text(fmt.Sprintf("%d tasks", len(h.tasks))),
					app.A().
						Href(githubRepoURL).
						Target("_blank").
						Rel("noopener noreferrer").
						Title("View source on GitHub").
						Aria("label", "GitHub repository").
						Class("inline-flex items-center justify-center h-7 w-7 rounded-full text-slate-600 hover:text-slate-900 hover:bg-slate-200/80 active:scale-95 transition-all").
						Body(app.Raw(githubLogoSVG)),
				),
			),

			app.If(h.isEditingJira, func() app.UI {
				return &components.JiraSettingsDialog{
					Config:   h.jiraConfig,
					OnSave:   h.onSaveJiraConfig,
					OnCancel: h.onCancelJiraSettings,
				}
			}),

			// Task input card
			app.Div().Class("bg-white p-2.5 rounded-2xl border border-slate-200 shadow-xs flex flex-col gap-2").Body(
				app.Div().Class("flex flex-col sm:flex-row gap-2").Body(
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
				app.Textarea().
					Class("w-full px-4 py-2.5 rounded-xl bg-slate-50/70 border border-slate-200 text-sm text-slate-800 placeholder-slate-400 resize-y focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:bg-white transition-all").
					Rows(2).
					Placeholder("Description (optional)").
					Attr("value", h.newTaskDescription).
					OnInput(h.ValueTo(&h.newTaskDescription)).
					OnChange(h.ValueTo(&h.newTaskDescription)),
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
				return &components.TimeSummary{
					// A copy, since onTaskSave updates h.tasks in place.
					Tasks: slices.Clone(h.tasks),
				}
			}),

			app.If(len(h.tasks) > 0, func() app.UI {
				return app.Div().Class("space-y-3").Body(
					app.Range(h.tasks).Slice(func(i int) app.UI {
						task := h.tasks[i]
						return &components.Task{
							Id:         task.Id,
							Data:       task,
							JiraConfig: h.jiraConfig,
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

	if h.db == nil {
		app.Logf("Cannot delete task: IndexedDB not ready")
		return
	}

	ctx.Async(func() {
		store := h.db.WriteTransaction("tasks")
		if err := store.Delete(taskId); err != nil {
			app.Logf("Failed to delete task %s: %v", taskId, err)
		}
		ctx.Dispatch(func(c app.Context) {
			h.loadTasks(c)
		})
	})
}

// onTaskSave is fired by task components when they need to persist state
// changes (stop, resume, edit). The action value is the updated models.Task.
func (h *Home) onTaskSave(ctx app.Context, a app.Action) {
	task, ok := a.Value.(models.Task)
	if !ok {
		return
	}

	if h.db == nil {
		app.Logf("Cannot save task: IndexedDB not ready")
		return
	}

	// Keep the in-memory list in sync so a later re-render doesn't pass stale
	// props (e.g. the old name) back into the task component.
	for i := range h.tasks {
		if h.tasks[i].Id == task.Id {
			h.tasks[i] = task
			break
		}
	}

	ctx.Async(func() {
		store := h.db.WriteTransaction("tasks")
		err := store.Put(taskRecord(task))
		if err != nil {
			app.Logf("Failed to save task %s: %v", task.Id, err)
		}
	})
}

func (h *Home) addNewTask(ctx app.Context, e app.Event) {
	if h.newTaskName == "" {
		return
	}

	if h.db == nil {
		app.Logf("Cannot add task: IndexedDB not ready")
		return
	}

	newTask := models.Task{
		Id:          uuid.NewV7().String(),
		Name:        h.newTaskName,
		Description: strings.TrimSpace(h.newTaskDescription),
		StartUnix:   time.Now().Unix(),
	}

	ctx.Async(func() {
		store := h.db.WriteTransaction("tasks")
		err := store.Add(taskRecord(newTask))
		if err != nil {
			app.Logf("Failed to add task: %v", err)
		}
	})

	// Stop any currently running task before starting the new one.
	ctx.NewActionWithValue("stopOtherTasks", newTask.Id)

	h.tasks = append(h.tasks, newTask)
	h.newTaskName = ""
	h.newTaskDescription = ""
}

func (h *Home) loadTasks(ctx app.Context) {
	if h.db == nil {
		return
	}

	ctx.Async(func() {
		store := h.db.ReadTransaction("tasks")
		values, err := store.GetAll()
		if err != nil {
			app.Logf("Failed to load tasks: %v", err)
			return
		}

		tasks := make([]models.Task, 0, len(values))
		for _, v := range values {
			task := models.Task{
				Id:        v.Get("id").String(),
				Name:      v.Get("name").String(),
				StartUnix: int64(v.Get("startUnix").Int()),
				Duration:  int64(v.Get("duration").Int()),
			}
			// Records created before descriptions existed have no such key.
			if d := v.Get("description"); d.Truthy() {
				task.Description = d.String()
			}
			if l := v.Get("loggedUnix"); l.Truthy() {
				task.LoggedUnix = int64(l.Int())
			}
			tasks = append(tasks, task)
		}

		ctx.Dispatch(func(c app.Context) {
			h.tasks = tasks
		})
	})
}

// taskRecord maps a task to its IndexedDB record.
func taskRecord(task models.Task) map[string]any {
	return map[string]any{
		"id":          task.Id,
		"name":        task.Name,
		"description": task.Description,
		"startUnix":   task.StartUnix,
		"duration":    task.Duration,
		"loggedUnix":  task.LoggedUnix,
	}
}

const githubRepoURL = "https://github.com/meeron/task-timer"

// githubLogoSVG is the GitHub mark, filled with the link's text color.
const githubLogoSVG = `<svg viewBox="0 0 16 16" width="18" height="18" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>`

// jiraSettingsId is the key of the Jira settings record in the "settings" store.
const jiraSettingsId = "jira"

func (h *Home) onJiraSettingsClick(ctx app.Context, e app.Event) {
	h.isEditingJira = true
}

func (h *Home) onCancelJiraSettings(ctx app.Context) {
	h.isEditingJira = false
}

// onSaveJiraConfig persists the Jira settings; an empty config removes them.
func (h *Home) onSaveJiraConfig(ctx app.Context, cfg jira.Config) {
	h.isEditingJira = false

	if h.db == nil {
		app.Logf("Cannot save Jira settings: IndexedDB not ready")
		return
	}

	h.jiraConfig = cfg

	ctx.Async(func() {
		store := h.db.WriteTransaction("settings")
		var err error
		if cfg.IsSet() {
			err = store.Put(map[string]any{
				"id":       jiraSettingsId,
				"baseUrl":  cfg.BaseURL,
				"email":    cfg.Email,
				"apiToken": cfg.APIToken,
			})
		} else {
			err = store.Delete(jiraSettingsId)
		}
		if err != nil {
			app.Logf("Failed to save Jira settings: %v", err)
		}
	})
}

func (h *Home) loadJiraConfig(ctx app.Context) {
	if h.db == nil {
		return
	}

	ctx.Async(func() {
		store := h.db.ReadTransaction("settings")
		v, err := store.Get(jiraSettingsId)
		if err != nil {
			app.Logf("Failed to load Jira settings: %v", err)
			return
		}
		if !v.Truthy() {
			return
		}

		cfg := jira.Config{
			BaseURL:  v.Get("baseUrl").String(),
			Email:    v.Get("email").String(),
			APIToken: v.Get("apiToken").String(),
		}

		ctx.Dispatch(func(c app.Context) {
			h.jiraConfig = cfg
		})
	})
}

type Home struct {
	app.Compo

	newTaskName        string
	newTaskDescription string
	updateAvailable    bool
	tasks              []models.Task
	db                 indexeddb.IDBDatabase
	jiraConfig         jira.Config
	isEditingJira      bool
}
