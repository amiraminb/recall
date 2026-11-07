package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	GroupID: "managing",
	Use:     "add [topic-title]",
	Short:   "Add something you've studied; review it tomorrow",
	Long: `Add a topic after studying it. Only a title is required.

Optionally attach a source (a URL or book reference) and tags.
Your first review is scheduled for tomorrow.

Examples:
  recall add "Docker Networking"
  recall add "Go interfaces" --tag go --source "https://go.dev/tour/methods/9"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var title string
		if len(args) == 0 {
			fmt.Fprint(cmd.OutOrStdout(), "What did you study? ")
			line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
			if err != nil && line == "" {
				return fmt.Errorf("enter a topic title, for example: recall add \"TCP handshake\"")
			}
			title = line
		} else {
			title = args[0]
		}
		title = strings.TrimSpace(title)
		if title == "" {
			return fmt.Errorf("topic title cannot be empty")
		}
		source, _ := cmd.Flags().GetString("source")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		store, err := getStorage()
		if err != nil {
			return err
		}
		topic, err := store.AddStudiedTopic(title, source, cleanTags(tags), time.Now())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added: %s\nFirst review: tomorrow (%s)\n\nRun 'recall' to see what's due.\n", topic.Title, topic.Card.Due.Format("Jan 2, 2006"))
		return nil
	},
}

func init() {
	addCmd.Flags().String("source", "", "Source URL or reference, such as a book chapter")
	addCmd.Flags().StringSlice("tag", nil, "Tags, comma-separated or repeated")
	rootCmd.AddCommand(addCmd)
}
