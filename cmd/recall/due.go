package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/amiraminb/recall/internal/storage"
	"github.com/spf13/cobra"
)

var dueCmd = &cobra.Command{
	Use:   "due",
	Short: "List topics due for review",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		tag, _ := cmd.Flags().GetString("tag")
		week, _ := cmd.Flags().GetBool("week")

		now := time.Now()
		until := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		if week {
			until = until.AddDate(0, 0, 7)
		}

		topics := store.GetDueTopics(until)

		if tag != "" {
			var filtered []storage.Topic
			for _, t := range topics {
				for _, tg := range t.Tags {
					if tg == tag {
						filtered = append(filtered, t)
						break
					}
				}
			}
			topics = filtered
		}

		if len(topics) == 0 {
			fmt.Println("No topics due for review!")
			return nil
		}

		label := "today"
		if week {
			label = "this week"
		}
		fmt.Printf("Topics due %s:\n\n", label)

		for _, t := range topics {
			days := int(time.Until(t.Card.Due).Hours() / 24)
			dueStr := "today"
			if days < 0 {
				dueStr = fmt.Sprintf("%d days overdue", -days)
			} else if days > 0 {
				dueStr = fmt.Sprintf("in %d days", days)
			}
			fmt.Printf("  - %s\n    [%s] due %s\n", t.Title, strings.Join(t.Tags, ", "), dueStr)
		}

		return nil
	},
}

func init() {
	dueCmd.Flags().String("tag", "", "Filter by tag")
	dueCmd.Flags().Bool("week", false, "Show topics due this week")
	rootCmd.AddCommand(dueCmd)
}
