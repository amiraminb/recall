package main

import (
	"fmt"
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

Without arguments, lists all topics and their next review dates.
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

		if len(args) == 0 {
			printTopics(cmd.OutOrStdout(), filterTopics(store.GetAllTopics(), ""), time.Now())
			return nil
		}

		// With args: show history for specific topic
		title := args[0]
		topic, err := resolveTopic(store, title)
		if err != nil {
			return err
		}

		history := store.GetReviewHistory(topic.ID)
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "History for: %s\nAdded: %s\nNext review: %s\n", topic.Title, topic.Created.Format("Jan 2, 2006"), topic.Card.Due.Format("Jan 2, 2006"))
		if topic.Source != "" {
			fmt.Fprintf(out, "Source: %s\n", topic.Source)
		}
		if len(topic.Tags) > 0 {
			fmt.Fprintf(out, "Tags: %s\n", strings.Join(topic.Tags, ", "))
		}
		if len(history) == 0 {
			fmt.Fprintln(out, "\nNo reviews yet.")
			return nil
		}

		fmt.Fprintln(out)
		table := tablewriter.NewTable(out)
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
