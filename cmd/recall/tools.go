package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amiraminb/recall/internal/config"
	"github.com/amiraminb/recall/internal/storage"
)

func getWikiPath() (string, error) {
	path, err := config.GetWikiPath()
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", fmt.Errorf("wiki path not configured. Run: recall init <path>")
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
