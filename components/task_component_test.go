package components

import (
	"testing"
	"time"
)

func TestFormatDurationTemplate(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0h 0m"},
		{15 * time.Minute, "0h 15m"},
		{1*time.Hour + 15*time.Minute, "1h 15m"},
		{2*time.Hour + 5*time.Minute + 30*time.Second, "2h 5m"},
	}

	for _, tc := range tests {
		got := formatDurationTemplate(tc.duration)
		if got != tc.expected {
			t.Errorf("formatDurationTemplate(%v) = %q, expected %q", tc.duration, got, tc.expected)
		}
	}
}

func TestParseDurationInput(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"1h 15m", 1*time.Hour + 15*time.Minute, false},
		{"1h15m", 1*time.Hour + 15*time.Minute, false},
		{"0h 45m", 45 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"45m", 45 * time.Minute, false},
		{"1h 15m 30s", 1*time.Hour + 15*time.Minute + 30*time.Second, false},
		{"1 hour 15 mins", 1*time.Hour + 15*time.Minute, false},
		{"2 hrs 30 min", 2*time.Hour + 30*time.Minute, false},
		{"1:15", 1*time.Hour + 15*time.Minute, false},
		{"01:15:30", 1*time.Hour + 15*time.Minute + 30*time.Second, false},
		{"30", 30 * time.Minute, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"-5m", 0, true},
	}

	for _, tc := range tests {
		got, err := parseDurationInput(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("parseDurationInput(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseDurationInput(%q) unexpected error: %v", tc.input, err)
			} else if got != tc.expected {
				t.Errorf("parseDurationInput(%q) = %v, expected %v", tc.input, got, tc.expected)
			}
		}
	}
}
