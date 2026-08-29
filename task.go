package main

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func (t *task) Render() app.UI {
	return app.Div().Body(
		app.P().Text(fmt.Sprintf("%s: %s", t.name, fmt.Sprintf("%s", t.duration))),
	)
}

func (t *task) OnMount(ctx app.Context) {
	t.ticker = time.NewTicker(1 * time.Second)
	ctx.Async(func() {
		for current := range t.ticker.C {
			ctx.Dispatch(func(c app.Context) {
				t.duration = current.Sub(t.start)
			})
		}
	})
}

type task struct {
	app.Compo

	name     string
	start    time.Time
	ticker   *time.Ticker
	duration time.Duration
}
