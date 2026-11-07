package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/amiraminb/recall/internal/storage"
)

func TestReviewSessionSavesProgressAndHandlesInput(t *testing.T) {
	for _, test := range []struct {
		name    string
		input   string
		reps    []int
		summary string
	}{
		{"multiple ratings", "3\n4\n1\n", []int{1, 1, 1}, "3 reviewed, 0 skipped, 0 remaining"},
		{"retry skip quit", "\nwrong\n5\n2\ns\nq\n", []int{1, 0, 0}, "1 reviewed, 1 skipped, 1 remaining"},
		{"eof saves earlier ratings", "3\n", []int{1, 0, 0}, "1 reviewed, 0 skipped, 2 remaining"},
		{"quit without rating", "q\n", []int{0, 0, 0}, "0 reviewed, 0 skipped, 3 remaining"},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			store, err := storage.NewStorage(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, title := range []string{"TCP", "UDP", "DNS"} {
				if _, err := store.AddStudiedTopic(title, "Book chapter 3", nil, time.Now()); err != nil {
					t.Fatal(err)
				}
			}
			var out bytes.Buffer
			if err := runReviewSession(strings.NewReader(test.input), &out, store, store.GetAllTopics()); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), test.summary) || !strings.Contains(out.String(), "Source: Book chapter 3") {
				t.Fatalf("unexpected session output:\n%s", out.String())
			}
			reopened, err := storage.NewStorage(dir)
			if err != nil {
				t.Fatal(err)
			}
			for i, topic := range reopened.GetAllTopics() {
				if topic.Card.Reps != test.reps[i] || len(reopened.GetReviewHistory(topic.ID)) != test.reps[i] {
					t.Fatalf("unexpected saved progress for %s: %+v", topic.Title, topic.Card)
				}
			}
		})
	}
}
