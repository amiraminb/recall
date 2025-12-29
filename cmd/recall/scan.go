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

		// Track which existing topics were found
		foundTitles := make(map[string]bool)

		added := 0
		for _, t := range topics {
			relPath, _ := filepath.Rel(wikiPath, t.File)
			foundTitles[t.Title] = true

			existing := store.GetTopicByTitle(t.Title)
			if existing == nil {
				store.AddTopic(t.Title, relPath, t.Tags)
				added++
				fmt.Printf("  + %s [%s]\n", t.Title, strings.Join(t.Tags, ", "))
			}
		}

		// Detect orphans
		existingTopics := store.GetAllTopics()
		var orphans []string
		for _, t := range existingTopics {
			if !foundTitles[t.Title] {
				orphans = append(orphans, t.Title)
			}
		}

		fmt.Printf("\nScanned %d topics, added %d new\n", len(topics), added)

		if len(orphans) > 0 {
			fmt.Printf("\nWarning: %d orphaned topics (not found in wiki):\n", len(orphans))
			for _, title := range orphans {
				fmt.Printf("  ? %s\n", title)
			}
			fmt.Println("\nThese may have been renamed or deleted.")
			fmt.Println("Use 'recall remove <title>' to clean up.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
