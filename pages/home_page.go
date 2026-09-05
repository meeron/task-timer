package pages

import (
	"fmt"
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
	return app.Main().Body(
		app.H1().Text("Task Timer"),

		app.Div().Body(
			app.Range(h.tasks).Map(func(key string) app.UI {
				return &components.Task{
					Id:   key,
					Data: h.tasks[key],
				}
			}),
		),

		app.Input().
			Type("text").
			Placeholder("Task name").
			Value(h.newTaskName).
			OnChange(h.ValueTo(&h.newTaskName)),
		app.Button().
			Text("Start new task").
			OnClick(h.addNewTask),
		// Displays an Update button when an update is available.
		app.If(h.updateAvailable, func() app.UI {
			return app.Button().
				Text("Update!").
				OnClick(h.onUpdateClick)
		}),
	)
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

	taskId := "_task_" + fmt.Sprintf("%s", uuid.NewV4())

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

type Home struct {
	app.Compo

	newTaskName     string
	updateAvailable bool
	tasks           map[string]models.Task
}
