package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func main() {
	app.Route("/", func() app.Composer { return &hello{} })
	app.RunWhenOnBrowser()

	// Standard HTTP routing (server-side):
	http.Handle("/", &app.Handler{
		Name:        "Task Timer",
		Description: "Measure your tasks time",
	})

	fmt.Println("Listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

type hello struct {
	app.Compo

	newTaskName     string
	updateAvailable bool
	tasks           map[string]taskData
}

func (h *hello) OnMount(ctx app.Context) {
	h.loadTasks(ctx.LocalStorage())
	ctx.Handle("deleteTask", h.onTaskDelete)
}

// OnAppUpdate satisfies the app.AppUpdater interface. It is called when the app
// is updated in background.
func (a *hello) OnAppUpdate(ctx app.Context) {
	a.updateAvailable = ctx.AppUpdateAvailable() // Reports that an app update is available.
}

func (h *hello) Render() app.UI {
	return app.Main().Body(
		app.H1().Text("Task Timer"),

		app.Div().Body(
			app.Range(h.tasks).Map(func(key string) app.UI {
				return &taskComponent{
					id:   key,
					data: h.tasks[key],
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

func (h *hello) onUpdateClick(ctx app.Context, e app.Event) {
	// Reloads the page to display the modifications.
	ctx.Reload()
}

func (h *hello) onTaskDelete(ctx app.Context, a app.Action) {
	taskId := a.Value.(string)
	app.Log(fmt.Sprintf("Delete task %s", taskId))

	// There's some bug when rendering map after delete.
	// For now let's just reload map again
	//delete(h.tasks, taskId)

	ctx.LocalStorage().Del(taskId)
	h.loadTasks(ctx.LocalStorage())
	ctx.Update()
}

func (h *hello) addNewTask(ctx app.Context, e app.Event) {
	taskId := "_task_" + fmt.Sprintf("%s", uuid.NewV4())

	h.tasks[taskId] = taskData{
		Name:      h.newTaskName,
		StartUnix: time.Now().Unix(),
	}

	err := ctx.LocalStorage().Set(taskId, h.tasks[taskId])
	if err != nil {
		app.Log(err)
	}

	h.newTaskName = ""
}

func (h *hello) loadTasks(storage app.BrowserStorage) {
	h.tasks = make(map[string]taskData)
	storage.ForEach(func(key string) {
		if !strings.HasPrefix(key, "_task_") {
			return
		}

		var data taskData
		err := storage.Get(key, &data)
		if err != nil {
			app.Log(err)
			return
		}

		h.tasks[key] = data
	})
}
