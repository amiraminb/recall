package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/amiraminb/recall/internal/config"
	"github.com/amiraminb/recall/internal/storage"
	"github.com/spf13/cobra"
)

var errTopicNotTracked = errors.New("topic not tracked")

func getStorage() (*storage.Storage, error) {
	dataDir, err := config.DataDir()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(dataDir, "reviews.json")); err == nil {
		return storage.NewStorage(dataDir)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if cfg != nil && cfg.WikiPath != "" {
		if err := storage.MigrateLegacy(cfg.WikiPath, dataDir); err != nil {
			return nil, err
		}
	}
	return storage.NewStorage(dataDir)
}

// IDs also disambiguate titles imported from older versions.
func resolveTopic(store *storage.Storage, name string) (*storage.Topic, error) {
	name = strings.TrimSpace(name)
	if topic := store.GetTopic(name); topic != nil {
		return topic, nil
	}
	if topic := store.GetTopicByFile(filepath.ToSlash(name)); topic != nil {
		return topic, nil
	}
	matches := store.FindTopicsByTitle(name)
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: %s (add it with 'recall add %q')", errTopicNotTracked, name, name)
	case 1:
		return matches[0], nil
	}
	var choices []string
	for _, topic := range matches {
		choices = append(choices, fmt.Sprintf("%s  %s", topic.ID, topic.File))
	}
	return nil, fmt.Errorf("%q matches %d topics; use an ID instead:\n  %s", name, len(matches), strings.Join(choices, "\n  "))
}

func completeTopicTitles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	store, err := getStorage()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	prefix := strings.ToLower(toComplete)
	seen := make(map[string]bool)
	var matches []string
	for _, topic := range store.GetAllTopics() {
		if strings.HasPrefix(strings.ToLower(topic.Title), prefix) && !seen[topic.Title] {
			matches = append(matches, topic.Title)
			seen[topic.Title] = true
		}
	}
	slices.Sort(matches)
	return matches, cobra.ShellCompDirectiveNoFileComp
}

func cleanTags(tags []string) []string {
	var cleaned []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" && !slices.Contains(cleaned, tag) {
			cleaned = append(cleaned, tag)
		}
	}
	return cleaned
}

func filterTopics(topics []storage.Topic, tag string) []storage.Topic {
	var filtered []storage.Topic
	for _, topic := range topics {
		if tag == "" || slices.Contains(topic.Tags, tag) {
			filtered = append(filtered, topic)
		}
	}
	slices.SortFunc(filtered, func(a, b storage.Topic) int {
		if order := a.Card.Due.Compare(b.Card.Due); order != 0 {
			return order
		}
		return strings.Compare(a.Title, b.Title)
	})
	return filtered
}

func endOfDay(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Add(-time.Nanosecond)
}
