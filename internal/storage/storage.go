package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
)

// Storage handles reading/writing the JSON data file.
type Storage struct {
	path string
	data *Data
}

func NewStorage(wikiPath string) (*Storage, error) {
	srsDir := filepath.Join(wikiPath, ".srs")
	if err := os.MkdirAll(srsDir, 0o755); err != nil {
		return nil, err
	}

	s := &Storage{
		path: filepath.Join(srsDir, "reviews.json"),
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
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Storage) AddTopic(title, file string, tags []string) (*Topic, error) {
	id := generateID(file, title)

	for _, t := range s.data.Topics {
		if t.ID == id {
			return &t, nil // Already exists
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
	review := ReviewLog{
		TopicID:    topicID,
		ReviewedAt: time.Now(),
		Rating:     rating,
	}
	s.data.Reviews = append(s.data.Reviews, review)
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
