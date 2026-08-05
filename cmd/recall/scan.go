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

		foundFiles := make(map[string]bool)

		added := 0
		updated := 0
		for _, t := range topics {
			relPath, err := filepath.Rel(wikiPath, t.File)
			if err != nil {
				return err
			}
			foundFiles[relPath] = true

			existing := store.GetTopicByFile(relPath)
			if existing == nil {
				if _, err := store.AddTopic(t.Title, relPath, t.Tags); err != nil {
					return err
				}
				added++
				fmt.Printf("  + %s [%s]\n", t.Title, strings.Join(t.Tags, ", "))
				continue
			}

			if existing.Title == t.Title && slices.Equal(existing.Tags, t.Tags) {
				continue
			}

			existing.Title = t.Title
			existing.Tags = t.Tags
			if err := store.UpdateTopic(existing); err != nil {
				return err
			}
			updated++
			fmt.Printf("  ~ %s [%s]\n", t.Title, strings.Join(t.Tags, ", "))
		}

		var orphans []string
		for _, t := range store.GetAllTopics() {
			if !foundFiles[t.File] {
				orphans = append(orphans, fmt.Sprintf("%s (%s)", t.Title, t.File))
			}
		}

		fmt.Printf("\nScanned: %d topics | Added: %d | Updated: %d\n", len(topics), added, updated)

		if len(orphans) > 0 {
			fmt.Printf("\nOrphaned topics (%d):\n", len(orphans))
			for _, orphan := range orphans {
				fmt.Printf("  ? %s\n", orphan)
			}
			fmt.Println("\nUse 'recall remove <title>' to clean up.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
