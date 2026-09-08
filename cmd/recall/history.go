package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/amiraminb/recall/internal/storage"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var understandingNames = map[fsrs.Rating]string{
	fsrs.Again: "didn't understand",
	fsrs.Hard:  "partially understood",
	fsrs.Good:  "understood well",
	fsrs.Easy:  "mastered",
}

var ratingNames = map[fsrs.Rating]string{
	fsrs.Again: "Again",
	fsrs.Hard:  "Hard",
	fsrs.Good:  "Good",
	fsrs.Easy:  "Easy",
}

var historyCmd = &cobra.Command{
	GroupID: "managing",
	Use:     "history [topic-title]",
	Short:   "Show what you studied and how you rated it",
	Long: `Display history for all topics or a specific topic.

Without arguments, shows all topics with their read status, first read date, and next due date.
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

			now := time.Now()
			table := tablewriter.NewTable(os.Stdout)
			table.Header("Topic", "Tags", "Read Date", "Due Date")

			var rows [][]any
			for _, t := range topics {
				tags := strings.Join(sortedTags(t.Tags), ", ")
				readDate := "-"

				if t.Card.State != fsrs.New {
					history := store.GetReviewHistory(t.ID)
					if len(history) > 0 {
						readDate = history[0].ReviewedAt.Format("Jan 2, 2006")
					}
				}

				rows = append(rows, []any{t.Title, tags, readDate, colorHistoryDueDate(t.Card.Due, now)})
			}

			table.Bulk(rows)
			table.Render()
			return nil
		}

		// With args: show history for specific topic
		title := args[0]
		topic, err := resolveTopic(store, title)
		if err != nil {
			return err
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
			if r.EffectiveKind(i == 0) == storage.KindFirstRead {
				rows = append(rows, []any{date, "First read", understandingNames[r.Rating]})
				continue
			}
			rows = append(rows, []any{date, "Review", ratingNames[r.Rating]})
		}

		table.Bulk(rows)
		table.Render()
		return nil
	},
}

func sortedTags(tags []string) []string {
	sorted := slices.Clone(tags)
	slices.Sort(sorted)
	return sorted
}

func formatDueDate(due time.Time) string {
	if due.IsZero() {
		return "-"
	}
	return due.Format("Jan 2, 2006")
}

func colorHistoryDueDate(due, now time.Time) string {
	text := formatDueDate(due)
	if due.IsZero() {
		return text
	}
	return color.New(historyDueColor(due, now)).Sprint(text)
}

func historyDueColor(due, now time.Time) color.Attribute {
	date := due.In(now.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, now.Location())

	if dueDay.Before(today) {
		return color.FgRed
	}
	if dueDay.Equal(today) {
		return color.FgGreen
	}
	return color.FgBlue
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
