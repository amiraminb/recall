package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <topic-title>",
	Short: "Remove a topic from tracking",
	Long: `Stop tracking a topic and remove it from your review schedule.

Use this to clean up orphaned topics (files that were renamed or deleted)
or topics you no longer want to review.

Note: This only removes the topic from recall's tracking. Your actual
note file is not affected.

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
		topic := store.GetTopicByTitle(title)
		if topic == nil {
			return fmt.Errorf("topic not found: %s", title)
		}

		if err := store.RemoveTopic(topic.ID); err != nil {
			return err
		}

		fmt.Printf("Removed: %s\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
