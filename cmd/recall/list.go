package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/amiraminb/recall/internal/storage"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	GroupID: "managing",
	Use:     "list",
	Short:   "List all your topics and their next review dates",
	Long: `List all topics, sorted by their next review date.

Examples:
  recall list
  recall list --tag go`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		tag, _ := cmd.Flags().GetString("tag")
		printTopics(cmd.OutOrStdout(), filterTopics(store.GetAllTopics(), tag), time.Now())
		return nil
	},
}

func printTopics(out io.Writer, topics []storage.Topic, now time.Time) {
	if len(topics) == 0 {
		fmt.Fprintln(out, "No topics found. Add one with 'recall add \"Topic\"'.")
		return
	}
	table := tablewriter.NewTable(out)
	table.Header("Topic", "Tags", "Added", "Next review")
	for _, topic := range topics {
		table.Append(topic.Title, strings.Join(sortedTags(topic.Tags), ", "), topic.Created.Format("Jan 2, 2006"), colorHistoryDueDate(topic.Card.Due, now))
	}
	table.Render()
}

func init() {
	listCmd.Flags().String("tag", "", "Filter by tag")
	rootCmd.AddCommand(listCmd)
}
