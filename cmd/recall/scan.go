package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/amiraminb/recall/internal/parser"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan wiki for @review topics",
	RunE: func(cmd *cobra.Command, args []string) error {
		wikiPath, err := getWikiPath()
		if err != nil {
			return err
		}

		store, err := getStorage()
		if err != nil {
			return err
		}

		topics, err := parser.ScanDirectory(wikiPath)
		if err != nil {
			return err
		}

		added := 0
		for _, t := range topics {
			relPath, _ := filepath.Rel(wikiPath, t.File)
			existing := store.GetTopicByTitle(t.Title)
			if existing == nil {
				store.AddTopic(t.Title, relPath, t.Tags)
				added++
				fmt.Printf("  + %s [%s]\n", t.Title, strings.Join(t.Tags, ", "))
			}
		}

		fmt.Printf("\nScanned %d topics, added %d new\n", len(topics), added)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
