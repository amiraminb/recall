package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note <topic-title>",
	Short: "Show the notes for a topic",
	Long: `Print the notes content from a topic file.

Example:
  recall note "Docker Networking"`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTopicTitles,
	RunE: func(cmd *cobra.Command, args []string) error {
		wikiPath, err := getWikiPath()
		if err != nil {
			return err
		}

		title := args[0]
		filePath, err := findTopicFile(wikiPath, title)
		if err != nil {
			return err
		}
		if filePath == "" {
			return fmt.Errorf("topic not found: %s", title)
		}

		notes, err := readNotes(filePath)
		if err != nil {
			return err
		}
		if notes == "" {
			return fmt.Errorf("no notes found in topic: %s", title)
		}

		fmt.Printf("Notes: %s\n\n%s\n", title, notes)
		return nil
	},
}

func readNotes(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	inFrontmatter := false
	firstLine := true

	for scanner.Scan() {
		line := scanner.Text()
		if firstLine {
			firstLine = false
			if line == "---" {
				inFrontmatter = true
				continue
			}
		}

		if inFrontmatter {
			if line == "---" {
				inFrontmatter = false
			}
			continue
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

func init() {
	rootCmd.AddCommand(noteCmd)
}
