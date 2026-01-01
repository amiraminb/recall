package main

import (
	"fmt"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [topic-title]",
	Short: "Show reading and review history",
	Long: `Display history for all topics or a specific topic.

Without arguments, shows all topics with their read status and first read date.
With a topic title, shows the complete review history for that topic.

Examples:
  recall history                     # List all topics with status
  recall history "Docker Networking" # Show full history for topic`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		// No args: show all topics with status
		if len(args) == 0 {
			topics := store.GetAllTopics()
			if len(topics) == 0 {
				fmt.Println("No topics tracked yet.")
				return nil
			}

			fmt.Println("All topics:\n")
			for _, t := range topics {
				status := "unread"
				firstRead := ""

				if t.Card.State != fsrs.New {
					status = "read"
					history := store.GetReviewHistory(t.ID)
					if len(history) > 0 {
						firstRead = fmt.Sprintf(" (first: %s)", history[0].ReviewedAt.Format("Jan 2, 2006"))
					}
				}

				fmt.Printf("  - %s [%s]%s\n", t.Title, status, firstRead)
			}
			return nil
		}

		// With args: show history for specific topic
		title := args[0]
		topic := store.GetTopicByTitle(title)
		if topic == nil {
			return fmt.Errorf("topic not found: %s", title)
		}

		history := store.GetReviewHistory(topic.ID)
		if len(history) == 0 {
			fmt.Printf("No history for: %s\n", title)
			return nil
		}

		fmt.Printf("History for: %s\n\n", title)

		for i, r := range history {
			if i == 0 {
				// First entry is the initial read
				understandingNames := map[fsrs.Rating]string{
					1: "didn't understand",
					2: "partially understood",
					3: "understood well",
					4: "mastered",
				}
				fmt.Printf("  %s - First read (%s)\n", r.ReviewedAt.Format("Jan 2, 2006"), understandingNames[r.Rating])
			} else {
				// Subsequent entries are reviews
				ratingNames := map[fsrs.Rating]string{
					fsrs.Again: "Again",
					fsrs.Hard:  "Hard",
					fsrs.Good:  "Good",
					fsrs.Easy:  "Easy",
				}
				fmt.Printf("  %s - Review: %s\n", r.ReviewedAt.Format("Jan 2, 2006"), ratingNames[r.Rating])
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
