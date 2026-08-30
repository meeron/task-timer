package main

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func (t *taskComponent) Render() app.UI {
	return app.Div().DataSet("id", t.Id).Styles(map[string]string{"display": "flex", "gap": "1rem"}).Body(
		app.Div().Styles(map[string]string{"display": "flex", "gap": "1rem"}).Body(
			app.Span().Text(t.Data.Name),
			app.Span().Text(fmt.Sprintf("%s", t.duration)),
		),
		app.Button().Text("Delete").OnClick(t.onDelete),
	)
}

func (t *taskComponent) OnMount(ctx app.Context) {
	t.ticker = time.NewTicker(1 * time.Second)
	t.duration = time.Second * time.Duration(time.Now().Unix()-t.Data.StartUnix)

	ctx.Async(func() {
		for current := range t.ticker.C {
			ctx.Dispatch(func(c app.Context) {
				t.duration = time.Second * time.Duration(current.Unix()-t.Data.StartUnix)
			})
		}
	})
}

func (t *taskComponent) onDelete(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue("deleteTask", t.Id)
}

type taskComponent struct {
	app.Compo

	Id       string
	Data     taskData
	ticker   *time.Ticker
	duration time.Duration
}

type taskData struct {
	Name      string
	StartUnix int64
}
