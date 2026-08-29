package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
	tasks           []task
}

func (h *hello) OnMount(ctx app.Context) {
	ctx.LocalStorage().ForEach(func(key string) {
		if !strings.HasPrefix(key, "_task") {
			return
		}

		var startUnix int64
		err := ctx.LocalStorage().Get(key, &startUnix)
		if err != nil {
			app.Log(err)
			return
		}

		h.tasks = append(h.tasks, task{
			name:   strings.TrimPrefix(key, "_task_"),
			start:  time.Unix(startUnix, 0),
			ticker: time.NewTicker(1 * time.Second),
		})
	})
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
			app.Range(h.tasks).Slice(func(i int) app.UI {
				return &h.tasks[i]
			}),
		),

		app.Input().Type("text").Placeholder("Task name").Value(h.newTaskName).OnChange(h.onChange),
		app.Button().Text("Start new task").OnClick(h.addNewTask),
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

func (h *hello) addNewTask(ctx app.Context, e app.Event) {
	newTask := task{
		name:   h.newTaskName,
		start:  time.Now(),
		ticker: time.NewTicker(1 * time.Second),
	}
	h.tasks = append(h.tasks, newTask)

	err := ctx.LocalStorage().Set("_task_"+newTask.name, newTask.start.Unix())
	if err != nil {
		app.Log(err)
	}

	h.newTaskName = ""
}

func (h *hello) onChange(ctx app.Context, e app.Event) {
	h.newTaskName = ctx.JSSrc().Get("value").String()
}
