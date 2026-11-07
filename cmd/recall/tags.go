package main

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	GroupID: "managing",
	Use:     "tags",
	Short:   "List tags and how many topics use each",
	Long: `Show all tags used across your topics and how many topics use each tag.

Useful for seeing how your knowledge is organized and finding topics
to review by category.

Example:
  recall tags

Output:
  #devops (5)
  #algorithms (3)
  #k8s (2)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		tags := store.GetAllTags()
		if len(tags) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No tags found.")
			return nil
		}

		var names []string
		for tag := range tags {
			names = append(names, tag)
		}
		slices.Sort(names)
		fmt.Fprintln(cmd.OutOrStdout(), "Tags:")
		for _, tag := range names {
			fmt.Fprintf(cmd.OutOrStdout(), "  #%s (%d)\n", tag, tags[tag])
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
