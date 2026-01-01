package main

import (
	"fmt"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history <topic-title>",
	Short: "Show review history for a topic",
	Long: `Display the complete review history for a specific topic.

Shows each review date and the rating you gave, helping you track
your learning progress over time.

Example:
  recall history "Docker Networking"

Output:
  Dec 1, 2024 - Good
  Dec 5, 2024 - Easy
  Dec 15, 2024 - Good`,
	Args: cobra.ExactArgs(1),
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

		history := store.GetReviewHistory(topic.ID)
		if len(history) == 0 {
			fmt.Printf("No review history for: %s\n", title)
			return nil
		}

		fmt.Printf("Review history for: %s\n\n", title)
		ratingNames := map[fsrs.Rating]string{
			fsrs.Again: "Again",
			fsrs.Hard:  "Hard",
			fsrs.Good:  "Good",
			fsrs.Easy:  "Easy",
		}

		for _, r := range history {
			fmt.Printf("  %s - %s\n", r.ReviewedAt.Format("Jan 2, 2006"), ratingNames[r.Rating])
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
