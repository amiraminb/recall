package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
)

// Storage handles reading/writing the JSON data file.
type Storage struct {
	path string
	data *Data
}

// NewStorage opens topics and review history in dataDir.
func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	s := &Storage{
		path: filepath.Join(dataDir, "reviews.json"),
	}

	if err := s.Load(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) Load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.data = NewData()
		return nil
	}
	if err != nil {
		return err
	}

	s.data = &Data{}
	return json.Unmarshal(data, s.data)
}

func (s *Storage) Save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return writeData(s.path, data)
}

func writeData(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".reviews-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := file.Chmod(0o644); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

// MigrateLegacy copies an existing wiki's review data on first use of standalone storage.
func MigrateLegacy(wikiPath, dataDir string) error {
	destination := filepath.Join(dataDir, "reviews.json")
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	data, err := os.ReadFile(filepath.Join(wikiPath, ".srs", "reviews.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var legacy Data
	if err := json.Unmarshal(data, &legacy); err != nil {
		return fmt.Errorf("read existing review history: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	return writeData(destination, data)
}

// AddStudiedTopic schedules a first recall attempt tomorrow without inventing a rating.
func (s *Storage) AddStudiedTopic(title, source string, tags []string, now time.Time) (*Topic, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("topic title cannot be empty")
	}
	if len(s.FindTopicsByTitle(title)) > 0 {
		return nil, fmt.Errorf("topic %q already exists; use 'recall review' to review it", title)
	}
	card := fsrs.NewCard()
	card.Due = now.AddDate(0, 0, 1)
	id := generateID(now.UTC().Format(time.RFC3339Nano), strings.ToLower(title))
	for s.GetTopic(id) != nil || len(s.GetReviewHistory(id)) > 0 {
		id = generateID(id, strings.ToLower(title))
	}
	topic := Topic{
		ID: id, Title: title,
		Source: strings.TrimSpace(source),
		Tags:   tags, Card: card, Created: now,
	}
	s.data.Topics = append(s.data.Topics, topic)
	if err := s.Save(); err != nil {
		s.data.Topics = s.data.Topics[:len(s.data.Topics)-1]
		return nil, err
	}
	return &s.data.Topics[len(s.data.Topics)-1], nil
}

// RecordReview saves the schedule and its review event together.
func (s *Storage) RecordReview(topicID string, rating fsrs.Rating, now time.Time) error {
	if !rating.Valid() {
		return fmt.Errorf("rating must be from 1 to 4")
	}
	topic := s.GetTopic(topicID)
	if topic == nil {
		return fmt.Errorf("topic not found: %s", topicID)
	}
	previous := topic.Card
	topic.Card = fsrs.NewScheduler().Review(topic.Card, rating, now)
	s.data.Reviews = append(s.data.Reviews, ReviewLog{
		TopicID: topicID, ReviewedAt: now, Rating: rating, Kind: KindReview,
	})
	if err := s.Save(); err != nil {
		topic.Card = previous
		s.data.Reviews = s.data.Reviews[:len(s.data.Reviews)-1]
		return err
	}
	return nil
}

func (s *Storage) AddTopic(title, file string, tags []string) (*Topic, error) {
	id := generateID(file, title)

	for i := range s.data.Topics {
		if s.data.Topics[i].ID == id {
			return &s.data.Topics[i], nil
		}
	}

	topic := Topic{
		ID:      id,
		Title:   title,
		File:    file,
		Tags:    tags,
		Card:    fsrs.NewCard(),
		Created: time.Now(),
	}

	s.data.Topics = append(s.data.Topics, topic)
	return &topic, s.Save()
}

func (s *Storage) GetTopic(id string) *Topic {
	for i := range s.data.Topics {
		if s.data.Topics[i].ID == id {
			return &s.data.Topics[i]
		}
	}
	return nil
}

func (s *Storage) GetTopicByTitle(title string) *Topic {
	for i := range s.data.Topics {
		if s.data.Topics[i].Title == title {
			return &s.data.Topics[i]
		}
	}
	return nil
}

func (s *Storage) GetTopicByFile(file string) *Topic {
	if file == "" {
		return nil
	}
	for i := range s.data.Topics {
		if s.data.Topics[i].File == file {
			return &s.data.Topics[i]
		}
	}
	return nil
}

// FindTopicsByTitle returns every topic sharing a title so callers can report an
// ambiguous match instead of silently acting on the wrong note.
func (s *Storage) FindTopicsByTitle(title string) []*Topic {
	var matches []*Topic
	for i := range s.data.Topics {
		if strings.EqualFold(s.data.Topics[i].Title, title) {
			matches = append(matches, &s.data.Topics[i])
		}
	}
	return matches
}

func (s *Storage) GetAllTopics() []Topic {
	return s.data.Topics
}

func (s *Storage) GetDueTopics(until time.Time) []Topic {
	var due []Topic
	for _, t := range s.data.Topics {
		if t.Card.Due.Before(until) || t.Card.Due.Equal(until) {
			due = append(due, t)
		}
	}
	return due
}

func (s *Storage) GetTopicsByTag(tag string) []Topic {
	var matched []Topic
	for _, t := range s.data.Topics {
		if slices.Contains(t.Tags, tag) {
			matched = append(matched, t)
		}
	}
	return matched
}

func (s *Storage) UpdateTopic(topic *Topic) error {
	for i := range s.data.Topics {
		if s.data.Topics[i].ID == topic.ID {
			s.data.Topics[i] = *topic
			return s.Save()
		}
	}
	return nil
}

func (s *Storage) AddReview(topicID string, rating fsrs.Rating) error {
	return s.addLog(topicID, rating, KindReview)
}

func (s *Storage) AddFirstRead(topicID string, rating fsrs.Rating) error {
	return s.addLog(topicID, rating, KindFirstRead)
}

func (s *Storage) addLog(topicID string, rating fsrs.Rating, kind LogKind) error {
	s.data.Reviews = append(s.data.Reviews, ReviewLog{
		TopicID:    topicID,
		ReviewedAt: time.Now(),
		Rating:     rating,
		Kind:       kind,
	})
	return s.Save()
}

func (s *Storage) GetReviewHistory(topicID string) []ReviewLog {
	var history []ReviewLog
	for _, r := range s.data.Reviews {
		if r.TopicID == topicID {
			history = append(history, r)
		}
	}
	return history
}

func (s *Storage) GetAllTags() map[string]int {
	tags := make(map[string]int)
	for _, t := range s.data.Topics {
		for _, tag := range t.Tags {
			tags[tag]++
		}
	}
	return tags
}

func generateID(file, title string) string {
	hash := sha256.Sum256([]byte(file + ":" + title))
	return hex.EncodeToString(hash[:8])
}

func (s *Storage) RemoveTopic(id string) error {
	for i, t := range s.data.Topics {
		if t.ID == id {
			s.data.Topics = append(s.data.Topics[:i], s.data.Topics[i+1:]...)
			return s.Save()
		}
	}
	return nil
}
