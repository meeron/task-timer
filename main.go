package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	StartTime string = "2026-08-29T10:20:44"
)

func main() {

	parsedDate, _ := time.Parse("2006-01-02T15:04:05", StartTime)

	app.Route("/", func() app.Composer { return &hello{startTime: parsedDate} })
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

func elapsedFrom(start time.Time, current time.Time) time.Duration {
	return current.Sub(start)
}

type hello struct {
	app.Compo

	startTime       time.Time
	ticker          *time.Ticker
	seconds         int
	updateAvailable bool
	duration        time.Duration
}

// OnAppUpdate satisfies the app.AppUpdater interface. It is called when the app
// is updated in background.
func (a *hello) OnAppUpdate(ctx app.Context) {
	a.updateAvailable = ctx.AppUpdateAvailable() // Reports that an app update is available.
}

func (h *hello) Render() app.UI {
	return app.Main().Body(
		app.H1().Text("Task Timer"),
		app.P().Text(fmt.Sprintf("Task 1: %s", h.duration)),

		// Displays an Update button when an update is available.
		app.If(h.updateAvailable, func() app.UI {
			return app.Button().
				Text("Update!").
				OnClick(h.onUpdateClick)
		}),
	)
}

func (h *hello) OnMount(ctx app.Context) {
	h.ticker = time.NewTicker(1 * time.Second)
	ctx.Async(func() {
		for t := range h.ticker.C {
			ctx.Dispatch(func(c app.Context) {
				h.duration = elapsedFrom(h.startTime, t)
			})
		}
	})
}

func (h *hello) onUpdateClick(ctx app.Context, e app.Event) {
	// Reloads the page to display the modifications.
	ctx.Reload()
}
