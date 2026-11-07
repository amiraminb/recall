package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	GroupID: "managing",
	Use:     "remove <topic-title>",
	Short:   "Remove a topic from your review schedule",
	Long: `Stop tracking a topic and remove it from your review schedule.

Use this for topics you no longer want to review.

Example:
  recall remove "Old Topic"`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTopicTitles,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		title := args[0]
		topic, err := resolveTopic(store, title)
		if err != nil {
			return err
		}

		title = topic.Title
		if err := store.RemoveTopic(topic.ID); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Removed: %s\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
