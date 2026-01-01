package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amiraminb/recall/internal/parser"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan wiki for topics to review",
	Long: `Scan your wiki directory for files marked with 'review: true' in frontmatter.

This command:
  - Discovers new topics and adds them to tracking
  - Updates tags if they've changed in the frontmatter
  - Detects orphaned topics (renamed or deleted files)

Run this after adding new notes or modifying tags.`,
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
		updated := 0
		for _, t := range topics {
			relPath, _ := filepath.Rel(wikiPath, t.File)
			foundTitles[t.Title] = true

			existing := store.GetTopicByTitle(t.Title)
			if existing == nil {
				// New topic
				store.AddTopic(t.Title, relPath, t.Tags)
				added++
				fmt.Printf("  + %s [%s]\n", t.Title, strings.Join(t.Tags, ", "))
			} else {
				// Check if tags changed
				if !slices.Equal(existing.Tags, t.Tags) {
					existing.Tags = t.Tags
					store.UpdateTopic(existing)
					updated++
					fmt.Printf("  ~ %s [%s]\n", t.Title, strings.Join(t.Tags, ", "))
				}
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

		fmt.Printf("\nScanned: %d topics | Added: %d | Updated: %d\n", len(topics), added, updated)

		if len(orphans) > 0 {
			fmt.Printf("\nOrphaned topics (%d):\n", len(orphans))
			for _, title := range orphans {
				fmt.Printf("  ? %s\n", title)
			}
			fmt.Println("\nUse 'recall remove <title>' to clean up.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
