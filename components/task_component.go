package components

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/models"
)

func (t *Task) Render() app.UI {
	return app.Div().DataSet("id", t.Id).Styles(map[string]string{"display": "flex", "gap": "1rem"}).Body(
		app.Div().Styles(map[string]string{"display": "flex", "gap": "1rem"}).Body(
			app.Span().Text(t.Data.Name),
			app.Span().Text(formatDuration(t.duration)),
		),
		app.If(t.isRunning, func() app.UI {
			return app.Button().Text("Stop").OnClick(t.onStop)
		}),
		app.If(!t.isRunning, func() app.UI {
			return app.Button().Text("Resume").OnClick(t.onResume)
		}),
		app.Button().Text("Add minutes").OnClick(t.onAddMinutes),
		app.Button().Text("Delete").OnClick(t.onDelete),
	)
}

func (t *Task) OnMount(ctx app.Context) {
	t.isRunning = t.Data.Duration == 0
	t.startUnix = t.Data.StartUnix
	t.ticker = time.NewTicker(1 * time.Second)
	t.ticker.Stop()

	ctx.Async(func() {
		for current := range t.ticker.C {
			ctx.Dispatch(func(c app.Context) {
				t.duration = time.Second * time.Duration(current.Unix()-t.startUnix)
			})
		}
	})

	if !t.isRunning {
		t.duration = time.Duration(t.Data.Duration)
		return
	}

	t.duration = time.Second * time.Duration(time.Now().Unix()-t.Data.StartUnix)
	t.ticker.Reset(1 * time.Second)
}

func (t *Task) onDelete(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue("deleteTask", t.Id)
}

func (t *Task) onStop(ctx app.Context, e app.Event) {
	t.ticker.Stop()
	t.isRunning = false

	// t.Data is not updated
	t.Data.Duration = int64(t.duration)
	ctx.LocalStorage().Set(t.Id, t.Data)
}

func (t *Task) onResume(ctx app.Context, e app.Event) {
	t.isRunning = true
	t.startUnix = time.Now().Unix() - int64(t.duration.Seconds())

	t.Data.StartUnix = t.startUnix
	t.Data.Duration = 0
	ctx.LocalStorage().Set(t.Id, t.Data)

	t.ticker.Reset(1 * time.Second)
}

func (t *Task) onAddMinutes(ctx app.Context, e app.Event) {
	minutesVal := app.Window().Call("prompt", "Minutes:")
	if minutesVal.IsNull() {
		return
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%sm", minutesVal.String()))
	if err != nil {
		app.Logf("%v", err)
		return
	}

	if t.isRunning {
		t.startUnix -= int64(duration.Seconds())
		t.Data.StartUnix = t.startUnix
		ctx.LocalStorage().Set(t.Id, t.Data)
		return
	}

	t.duration += duration
	t.Data.Duration = int64(t.duration)
	ctx.LocalStorage().Set(t.Id, t.Data)
}

func formatDuration(duration time.Duration) []string {
	// 0:hours, 1:minutes, 2:seconds
	parts := make([]string, 3)

	hours := int(duration.Hours())
	if hours > 0 {
		parts[0] = fmt.Sprintf("%dh", hours)
	}

	minutes := int(duration.Minutes())
	if minutes > 59 {
		parts[1] = fmt.Sprintf("%dm", minutes-(hours*60))
	} else {
		parts[1] = fmt.Sprintf("%dm", minutes)
	}

	seconds := int(duration.Seconds())
	if seconds > 59 {
		parts[2] = fmt.Sprintf("%ds", seconds-(minutes*60))
	} else {
		parts[2] = fmt.Sprintf("%ds", seconds)
	}

	return parts
}

type Task struct {
	app.Compo

	Id        string
	Data      models.Task
	ticker    *time.Ticker
	duration  time.Duration
	isRunning bool
	startUnix int64
}
