package components

import (
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/models"
)

// TimeSummary shows the total time tracked across all tasks. It ticks on its
// own so a running task keeps the total live without re-rendering the list.
type TimeSummary struct {
	app.Compo

	Tasks []models.Task

	now  time.Time
	done chan struct{}
}

func (s *TimeSummary) OnMount(ctx app.Context) {
	s.now = time.Now()
	s.done = make(chan struct{})
	done := s.done

	ctx.Async(func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case current := <-ticker.C:
				ctx.Dispatch(func(c app.Context) {
					s.now = current
				})
			}
		}
	})
}

func (s *TimeSummary) OnDismount() {
	if s.done != nil {
		close(s.done)
		s.done = nil
	}
}

func (s *TimeSummary) Render() app.UI {
	return app.Div().
		Class("bg-white rounded-2xl border border-slate-200 shadow-xs px-4 py-3 sm:px-5 flex items-baseline gap-3").
		Body(
			app.Span().Class("text-xs font-semibold uppercase tracking-wide text-slate-500").Text("Total time"),
			app.Span().Class("font-mono text-xl font-bold tracking-tight text-slate-900").Text(formatDuration(totalDuration(s.Tasks, s.now))),
		)
}

// totalDuration returns the total tracked time of all tasks at now. A task with
// Duration == 0 is running, so its time is elapsed since StartUnix.
func totalDuration(tasks []models.Task, now time.Time) time.Duration {
	var total time.Duration
	for _, task := range tasks {
		d := time.Duration(task.Duration)
		if task.Duration == 0 {
			d = time.Duration(now.Unix()-task.StartUnix) * time.Second
		}
		if d < 0 {
			d = 0
		}
		total += d
	}
	return total
}
