package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [topic-title]",
	Short: "Remove a topic from tracking",
	Args:  cobra.ExactArgs(1),
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
