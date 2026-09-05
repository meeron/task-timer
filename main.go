package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/pages"
)

func main() {
	app.Route("/", func() app.Composer { return &pages.Home{} })
	app.RunWhenOnBrowser()

	// Standard HTTP routing (server-side):
	http.Handle("/", &app.Handler{
		Name:        "Task Timer",
		Description: "Measure your tasks time",
		Styles:      []string{"web/styles.css"},
	})

	fmt.Println("Listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
