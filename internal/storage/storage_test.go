package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	store, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	return store
}

// Two notes in different folders can share a filename, so they must stay
// separate topics rather than collapsing into one.
func TestSameTitleDifferentFilesStaySeparate(t *testing.T) {
	store := newTestStorage(t)

	if _, err := store.AddTopic("Networking", "aws/Networking.md", nil); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}
	if _, err := store.AddTopic("Networking", "k8s/Networking.md", nil); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	if got := len(store.GetAllTopics()); got != 2 {
		t.Fatalf("expected 2 tracked topics, got %d", got)
	}

	matches := store.FindTopicsByTitle("Networking")
	if len(matches) != 2 {
		t.Errorf("FindTopicsByTitle found %d matches, want 2", len(matches))
	}

	if topic := store.GetTopicByFile("k8s/Networking.md"); topic == nil {
		t.Error("GetTopicByFile could not find the k8s note")
	}
}

// Stored paths use forward slashes so the path a user copies from an ambiguity
// error resolves the same way on any platform.
func TestGetTopicByFileUsesSlashPaths(t *testing.T) {
	store := newTestStorage(t)

	if _, err := store.AddTopic("Networking", "aws/vpc/Networking.md", nil); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	if store.GetTopicByFile("aws/vpc/Networking.md") == nil {
		t.Error("could not resolve a topic by its stored slash path")
	}
	if store.GetTopicByFile("aws/vpc/Missing.md") != nil {
		t.Error("resolved a path that was never tracked")
	}
}

func TestAddTopicIsIdempotent(t *testing.T) {
	store := newTestStorage(t)

	first, err := store.AddTopic("Docker", "Docker.md", []string{"devops"})
	if err != nil {
		t.Fatalf("AddTopic: %v", err)
	}
	second, err := store.AddTopic("Docker", "Docker.md", []string{"devops"})
	if err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	if first.ID != second.ID {
		t.Errorf("re-adding the same file produced a new ID: %s then %s", first.ID, second.ID)
	}
	if got := len(store.GetAllTopics()); got != 1 {
		t.Errorf("expected 1 topic after a duplicate add, got %d", got)
	}
}

// A returned topic must alias stored data, or edits made through it are lost.
func TestAddTopicReturnsLiveReference(t *testing.T) {
	store := newTestStorage(t)

	if _, err := store.AddTopic("Go", "Go.md", nil); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}
	existing, err := store.AddTopic("Go", "Go.md", nil)
	if err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	existing.Card.Stability = 42
	if err := store.UpdateTopic(existing); err != nil {
		t.Fatalf("UpdateTopic: %v", err)
	}

	if got := store.GetTopicByFile("Go.md").Card.Stability; got != 42 {
		t.Errorf("stability did not persist: got %v, want 42", got)
	}
}

func TestFirstReadAndReviewAreDistinguished(t *testing.T) {
	store := newTestStorage(t)

	topic, err := store.AddTopic("FSRS", "FSRS.md", nil)
	if err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	if err := store.AddFirstRead(topic.ID, fsrs.Good); err != nil {
		t.Fatalf("AddFirstRead: %v", err)
	}
	if err := store.AddReview(topic.ID, fsrs.Again); err != nil {
		t.Fatalf("AddReview: %v", err)
	}

	history := store.GetReviewHistory(topic.ID)
	if len(history) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(history))
	}
	if history[0].Kind != KindFirstRead {
		t.Errorf("first entry kind = %q, want %q", history[0].Kind, KindFirstRead)
	}
	if history[1].Kind != KindReview {
		t.Errorf("second entry kind = %q, want %q", history[1].Kind, KindReview)
	}
}

// Data written before Kind existed has no kind field; the first entry for a
// topic must still read as the first read.
func TestEffectiveKindBackfillsLegacyLogs(t *testing.T) {
	legacy := ReviewLog{TopicID: "abc", Rating: fsrs.Good}

	if got := legacy.EffectiveKind(true); got != KindFirstRead {
		t.Errorf("legacy first entry = %q, want %q", got, KindFirstRead)
	}
	if got := legacy.EffectiveKind(false); got != KindReview {
		t.Errorf("legacy later entry = %q, want %q", got, KindReview)
	}

	tagged := ReviewLog{TopicID: "abc", Rating: fsrs.Good, Kind: KindReview}
	if got := tagged.EffectiveKind(true); got != KindReview {
		t.Errorf("explicit kind was overridden: got %q, want %q", got, KindReview)
	}
}

func TestLoadTolerLegacyFileWithoutKind(t *testing.T) {
	dir := t.TempDir()
	srsDir := filepath.Join(dir, ".srs")
	if err := os.MkdirAll(srsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	legacy := `{
      "topics": [{"id":"a1","title":"Old","file":"Old.md","tags":["x"],
        "card":{"due":"2026-01-01T00:00:00Z","stability":5,"difficulty":5,
        "reps":1,"lapses":0,"state":2,"last_review":"2025-12-01T00:00:00Z"},
        "created":"2025-12-01T00:00:00Z"}],
      "reviews": [{"topic_id":"a1","reviewed_at":"2025-12-01T00:00:00Z","rating":3}]
    }`
	if err := os.WriteFile(filepath.Join(srsDir, "reviews.json"), []byte(legacy), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage on legacy data: %v", err)
	}

	topic := store.GetTopicByTitle("Old")
	if topic == nil {
		t.Fatal("legacy topic was not loaded")
	}
	if topic.Card.Stability != 5 {
		t.Errorf("legacy stability = %v, want 5", topic.Card.Stability)
	}

	history := store.GetReviewHistory("a1")
	if len(history) != 1 {
		t.Fatalf("expected 1 legacy log, got %d", len(history))
	}
	if history[0].Kind != "" {
		t.Errorf("legacy log gained a kind on load: %q", history[0].Kind)
	}
}

func TestRemoveTopic(t *testing.T) {
	store := newTestStorage(t)

	keep, err := store.AddTopic("Keep", "Keep.md", nil)
	if err != nil {
		t.Fatalf("AddTopic: %v", err)
	}
	drop, err := store.AddTopic("Drop", "Drop.md", nil)
	if err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	if err := store.RemoveTopic(drop.ID); err != nil {
		t.Fatalf("RemoveTopic: %v", err)
	}

	remaining := store.GetAllTopics()
	if len(remaining) != 1 || remaining[0].ID != keep.ID {
		t.Errorf("remove left %d topics, expected only %q", len(remaining), keep.Title)
	}
}

func TestGetDueTopics(t *testing.T) {
	store := newTestStorage(t)
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		title string
		due   time.Time
	}{
		{"Overdue", now.AddDate(0, 0, -3)},
		{"DueNow", now},
		{"Future", now.AddDate(0, 0, 10)},
	} {
		topic, err := store.AddTopic(tc.title, tc.title+".md", nil)
		if err != nil {
			t.Fatalf("AddTopic: %v", err)
		}
		topic.Card.Due = tc.due
		if err := store.UpdateTopic(topic); err != nil {
			t.Fatalf("UpdateTopic: %v", err)
		}
	}

	due := store.GetDueTopics(now)
	if len(due) != 2 {
		t.Errorf("GetDueTopics returned %d topics, want 2 (overdue + due now)", len(due))
	}
	for _, topic := range due {
		if topic.Title == "Future" {
			t.Error("a future topic was reported as due")
		}
	}
}

func TestSavedDataRoundTrips(t *testing.T) {
	dir := t.TempDir()

	store, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	topic, err := store.AddTopic("Round", "Round.md", []string{"a", "b"})
	if err != nil {
		t.Fatalf("AddTopic: %v", err)
	}
	if err := store.AddReview(topic.ID, fsrs.Hard); err != nil {
		t.Fatalf("AddReview: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".srs", "reviews.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var parsed Data
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	reopened, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if got := reopened.GetTopicByTitle("Round"); got == nil {
		t.Error("topic did not survive a save and reload")
	}
	if got := len(reopened.GetReviewHistory(topic.ID)); got != 1 {
		t.Errorf("review history after reload = %d entries, want 1", got)
	}
}

func TestGetAllTags(t *testing.T) {
	store := newTestStorage(t)

	if _, err := store.AddTopic("A", "A.md", []string{"go", "cli"}); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}
	if _, err := store.AddTopic("B", "B.md", []string{"go"}); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	tags := store.GetAllTags()
	if tags["go"] != 2 {
		t.Errorf("tag go counted %d times, want 2", tags["go"])
	}
	if tags["cli"] != 1 {
		t.Errorf("tag cli counted %d times, want 1", tags["cli"])
	}
}
