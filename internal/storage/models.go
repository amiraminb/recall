package storage

import (
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
)

type Topic struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	File    string    `json:"file"`
	Tags    []string  `json:"tags"`
	Card    fsrs.Card `json:"card"`
	Created time.Time `json:"created"`
}

// ReviewLog represents a single review event
type ReviewLog struct {
	TopicID    string      `json:"topic_id"`
	ReviewedAt time.Time   `json:"reviewed_at"`
	Rating     fsrs.Rating `json:"rating"`
}

// Data is the root structure for the JSON storage file
type Data struct {
	Topics  []Topic     `json:"topics"`
	Reviews []ReviewLog `json:"reviews"`
}

// NewData creates an empty data structure
func NewData() *Data {
	return &Data{
		Topics:  []Topic{},
		Reviews: []ReviewLog{},
	}
}
