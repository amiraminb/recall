package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
)

func TestStudiedTopicLifecycle(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 6, 14, 0, 0, 0, time.UTC)
	topic, err := store.AddStudiedTopic(" TCP handshake ", "RFC 9293", []string{"networking"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if topic.Title != "TCP handshake" || topic.File != "" || !topic.Card.Due.Equal(now.AddDate(0, 0, 1)) {
		t.Fatalf("unexpected standalone topic: %+v", topic)
	}
	if topic.Card.State != fsrs.New || topic.Card.Reps != 0 || len(store.GetReviewHistory(topic.ID)) != 0 {
		t.Fatal("adding a studied topic must not fabricate a recall rating")
	}
	if len(store.GetDueTopics(now)) != 0 || len(store.GetDueTopics(now.AddDate(0, 0, 1))) != 1 {
		t.Fatal("first review should be due tomorrow")
	}
	for _, title := range []string{" ", "tcp HANDSHAKE", " TCP handshake "} {
		if _, err := store.AddStudiedTopic(title, "", nil, now); err == nil {
			t.Fatalf("accepted empty or duplicate title %q", title)
		}
	}
	if len(store.GetAllTopics()) != 1 {
		t.Fatal("rejected adds changed the topic collection")
	}
	reviewedAt := now.AddDate(0, 0, 1)
	if err := store.RecordReview(topic.ID, fsrs.Good, reviewedAt); err != nil {
		t.Fatal(err)
	}
	reopened, err := NewStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	saved := reopened.GetTopic(topic.ID)
	if !reflect.DeepEqual(saved, topic) || saved.Source != "RFC 9293" {
		t.Fatalf("topic details or schedule did not round trip: %+v", saved)
	}
	history := reopened.GetReviewHistory(topic.ID)
	if len(history) != 1 || history[0].Kind != KindReview || !history[0].ReviewedAt.Equal(reviewedAt) {
		t.Fatalf("first recall should be recorded as a review: %+v", history)
	}
	if saved.Card.Reps != 1 || !saved.Card.Due.After(reviewedAt) {
		t.Fatalf("first recall did not initialize FSRS: %+v", saved.Card)
	}
}

func TestRenamingAndReaddingDoesNotReuseHistory(t *testing.T) {
	store := newTestStorage(t)
	now := time.Now()
	first, err := store.AddStudiedTopic("Networking", "", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	firstID := first.ID
	if err := store.RecordReview(firstID, fsrs.Good, now); err != nil {
		t.Fatal(err)
	}
	first.Title = "TCP"
	if err := store.UpdateTopic(first); err != nil {
		t.Fatal(err)
	}
	second, err := store.AddStudiedTopic("Networking", "", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == firstID || len(store.GetReviewHistory(second.ID)) != 0 {
		t.Fatal("reusing a renamed title reused its history")
	}
	if err := store.RemoveTopic(firstID); err != nil {
		t.Fatal(err)
	}
	third, err := store.AddStudiedTopic("TCP", "", nil, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if third.ID == firstID || len(store.GetReviewHistory(third.ID)) != 0 {
		t.Fatal("a newly added topic inherited removed history")
	}
}

func TestRecordReviewFailurePreservesScheduleAndHistory(t *testing.T) {
	store := newTestStorage(t)
	now := time.Now()
	topic, err := store.AddStudiedTopic("TCP", "", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	before := topic.Card
	for _, rating := range []fsrs.Rating{0, 5} {
		if err := store.RecordReview(topic.ID, rating, now); err == nil {
			t.Fatal("invalid rating accepted")
		}
	}
	store.path = t.TempDir() // Renaming a data file over a directory must fail.
	if err := store.RecordReview(topic.ID, fsrs.Good, now); err == nil {
		t.Fatal("expected persistence failure")
	}
	if topic.Card != before || len(store.GetReviewHistory(topic.ID)) != 0 {
		t.Fatal("a failed save left the schedule and history changed")
	}
}

func TestMigrateLegacyPreservesDataAndDoesNotOverwrite(t *testing.T) {
	wiki := t.TempDir()
	legacy, err := NewStorage(filepath.Join(wiki, ".srs"))
	if err != nil {
		t.Fatal(err)
	}
	topic, err := legacy.AddTopic("Old topic", "missing.md", []string{"old"})
	if err != nil {
		t.Fatal(err)
	}
	if err := legacy.RecordReview(topic.ID, fsrs.Hard, time.Now()); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(legacy.path)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "recall")
	if err := MigrateLegacy(wiki, dir); err != nil {
		t.Fatal(err)
	}
	migrated, err := NewStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	copied, err := os.ReadFile(migrated.path)
	if err != nil || string(before) != string(copied) {
		t.Fatal("migration changed IDs, schedules, or review history")
	}
	if err := migrated.RemoveTopic(topic.ID); err != nil {
		t.Fatal(err)
	}
	if err := MigrateLegacy(wiki, dir); err != nil {
		t.Fatal(err)
	}
	if err := migrated.Load(); err != nil {
		t.Fatal(err)
	}
	if len(migrated.GetAllTopics()) != 0 {
		t.Fatal("migration overwrote standalone data")
	}
	after, err := os.ReadFile(legacy.path)
	if err != nil || string(before) != string(after) {
		t.Fatal("migration modified the original data")
	}
}

func TestMigrateLegacyRejectsInvalidData(t *testing.T) {
	wiki := t.TempDir()
	srs := filepath.Join(wiki, ".srs")
	if err := os.MkdirAll(srs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srs, "reviews.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := MigrateLegacy(wiki, dir); err == nil {
		t.Fatal("expected an error for invalid legacy data")
	}
	if _, err := os.Stat(filepath.Join(dir, "reviews.json")); !os.IsNotExist(err) {
		t.Fatal("invalid data was written to the new store")
	}
}
