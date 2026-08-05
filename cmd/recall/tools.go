package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/amiraminb/recall/internal/config"
	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/amiraminb/recall/internal/storage"
)

var errTopicNotTracked = errors.New("topic not tracked")

// expandPath resolves a leading ~ and returns an absolute path so a config
// written on one machine does not silently point somewhere else on another.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return filepath.Abs(path)
}

func getWikiPath() (string, error) {
	path, err := config.GetWikiPath()
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", fmt.Errorf("wiki path not configured. Run: recall init <path>")
	}

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("configured wiki path is unreachable: %s\nRun: recall init <path>", path)
	}

	return path, nil
}

func getStorage() (*storage.Storage, error) {
	wikiPath, err := getWikiPath()
	if err != nil {
		return nil, err
	}
	return storage.NewStorage(wikiPath)
}

// Rejects blank input, which fmt.Scanln would silently leave as rating 0.
func promptRating(in io.Reader) (fsrs.Rating, error) {
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return 0, fmt.Errorf("no rating entered")
	}

	value, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, fmt.Errorf("rating must be a number from 1 to 4")
	}

	rating := fsrs.Rating(value)
	if !rating.Valid() {
		return 0, fmt.Errorf("rating must be from 1 to 4, got %d", value)
	}

	return rating, nil
}

// Uses the tracked path when available so a topic resolves without re-walking
// the wiki, falling back to a filename search for notes that are not tracked.
func resolveTopicPath(wikiPath, title string) (string, error) {
	store, err := getStorage()
	if err != nil {
		return "", err
	}

	topic, err := resolveTopic(store, title)
	if err != nil {
		if !errors.Is(err, errTopicNotTracked) {
			return "", err
		}

		path, err := findTopicFile(wikiPath, title)
		if err != nil {
			return "", err
		}
		if path == "" {
			return "", fmt.Errorf("topic not found: %s", title)
		}
		return path, nil
	}

	path := filepath.Join(wikiPath, topic.File)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("file missing for %q: %s (run 'recall scan')", title, topic.File)
	}
	return path, nil
}

// Accepts a title or, when several notes share one, the relative file path that
// the ambiguity error reports.
func resolveTopic(store *storage.Storage, nameOrPath string) (*storage.Topic, error) {
	if topic := store.GetTopicByFile(filepath.ToSlash(nameOrPath)); topic != nil {
		return topic, nil
	}

	matches := store.FindTopicsByTitle(nameOrPath)

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: %s (run 'recall scan' first)", errTopicNotTracked, nameOrPath)
	case 1:
		return matches[0], nil
	}

	var files []string
	for _, m := range matches {
		files = append(files, m.File)
	}
	return nil, fmt.Errorf("%q matches %d tracked notes; pass one of these paths instead:\n  %s",
		nameOrPath, len(matches), strings.Join(files, "\n  "))
}

func listWikiTitles(wikiPath string) ([]string, error) {
	var titles []string
	seen := make(map[string]bool)

	err := filepath.WalkDir(wikiPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		baseName := strings.TrimSuffix(d.Name(), ".md")
		if baseName == "" || seen[baseName] {
			return nil
		}

		seen[baseName] = true
		titles = append(titles, baseName)
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.Sort(titles)
	return titles, nil
}
