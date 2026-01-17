package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/olekukonko/tablewriter"
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
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeTopicTitles,
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

			table := tablewriter.NewTable(os.Stdout)
			table.Header("Topic", "Tags", "Read Date")

			var rows [][]any
			for _, t := range topics {
				tags := strings.Join(t.Tags, ", ")
				readDate := "-"

				if t.Card.State != fsrs.New {
					history := store.GetReviewHistory(t.ID)
					if len(history) > 0 {
						readDate = history[0].ReviewedAt.Format("Jan 2, 2006")
					}
				}

				rows = append(rows, []any{t.Title, tags, readDate})
			}

			table.Bulk(rows)
			table.Render()
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

		table := tablewriter.NewTable(os.Stdout)
		table.Header("Date", "Type", "Rating")

		var rows [][]any
		for i, r := range history {
			date := r.ReviewedAt.Format("Jan 2, 2006")
			if i == 0 {
				understandingNames := map[fsrs.Rating]string{
					1: "didn't understand",
					2: "partially understood",
					3: "understood well",
					4: "mastered",
				}
				rows = append(rows, []any{date, "First read", understandingNames[r.Rating]})
			} else {
				ratingNames := map[fsrs.Rating]string{
					fsrs.Again: "Again",
					fsrs.Hard:  "Hard",
					fsrs.Good:  "Good",
					fsrs.Easy:  "Easy",
				}
				rows = append(rows, []any{date, "Review", ratingNames[r.Rating]})
			}
		}

		table.Bulk(rows)
		table.Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
