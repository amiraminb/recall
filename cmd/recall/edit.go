package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <topic-title>",
	Short: "Edit a topic in your editor",
	Long: `Open a topic file in your editor so you can update your notes.

Example:
  recall edit "Docker Networking"`,
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

		if err := openEditor(filePath); err != nil {
			return err
		}

		fmt.Printf("Editing %s\n", filePath)
		return nil
	},
}

func openEditor(filePath string) error {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		editor = "nvim"
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(editCmd)
}
