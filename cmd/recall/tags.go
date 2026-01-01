package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags with topic counts",
	Long: `Show all tags used across your topics and how many topics use each tag.

Useful for seeing how your knowledge is organized and finding topics
to review by category.

Example:
  recall tags

Output:
  #devops (5)
  #algorithms (3)
  #k8s (2)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		tags := store.GetAllTags()
		if len(tags) == 0 {
			fmt.Println("No tags found.")
			return nil
		}

		fmt.Println("Tags:")
		for tag, count := range tags {
			fmt.Printf("  #%s (%d)\n", tag, count)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
