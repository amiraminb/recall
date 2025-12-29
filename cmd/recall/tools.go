package main

import (
	"fmt"

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
