package components

import (
	"testing"
	"time"

	"github.com/meeron/task-timer/models"
)

func TestTotalDuration(t *testing.T) {
	now := time.Unix(10_000, 0)

	tests := []struct {
		name     string
		tasks    []models.Task
		expected time.Duration
	}{
		{"no tasks", nil, 0},
		{
			"stopped tasks",
			[]models.Task{
				{Duration: int64(30 * time.Minute)},
				{Duration: int64(1 * time.Hour)},
			},
			90 * time.Minute,
		},
		{
			"running task counts elapsed time",
			[]models.Task{
				{StartUnix: 10_000 - 120},
				{Duration: int64(1 * time.Minute)},
			},
			3 * time.Minute,
		},
		{
			"start in the future counts as zero",
			[]models.Task{{StartUnix: 10_060}},
			0,
		},
	}

	for _, tc := range tests {
		if got := totalDuration(tc.tasks, now); got != tc.expected {
			t.Errorf("%s: totalDuration() = %v, expected %v", tc.name, got, tc.expected)
		}
	}
}
