package main

import (
	"slices"
	"testing"
	"time"

	"github.com/fatih/color"
)

func TestSortedTags(t *testing.T) {
	tags := []string{"zsh", "api", "go"}

	got := sortedTags(tags)
	want := []string{"api", "go", "zsh"}
	if !slices.Equal(got, want) {
		t.Errorf("sortedTags() = %v, want %v", got, want)
	}
	if !slices.Equal(tags, []string{"zsh", "api", "go"}) {
		t.Errorf("sortedTags() mutated the input: %v", tags)
	}
}

func TestFormatDueDate(t *testing.T) {
	due := time.Date(2026, time.January, 12, 15, 30, 0, 0, time.UTC)
	if got := formatDueDate(due); got != "Jan 12, 2026" {
		t.Errorf("formatDueDate() = %q, want %q", got, "Jan 12, 2026")
	}
	if got := formatDueDate(time.Time{}); got != "-" {
		t.Errorf("formatDueDate(zero) = %q, want %q", got, "-")
	}
}

func TestColorHistoryDueDateUsesCalendarDay(t *testing.T) {
	location := time.FixedZone("test", -5*60*60)
	now := time.Date(2026, time.January, 12, 15, 30, 0, 0, location)

	for _, test := range []struct {
		name      string
		due       time.Time
		wantColor color.Attribute
	}{
		{name: "past", due: now.AddDate(0, 0, -1), wantColor: color.FgRed},
		{name: "today", due: now.Add(-8 * time.Hour), wantColor: color.FgGreen},
		{name: "future", due: now.AddDate(0, 0, 1), wantColor: color.FgBlue},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := historyDueColor(test.due, now); got != test.wantColor {
				t.Errorf("historyDueColor() = %v, want %v", got, test.wantColor)
			}
		})
	}
}
