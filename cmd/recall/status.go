package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show review status overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		topics := store.GetAllTopics()
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		week := today.AddDate(0, 0, 7)

		dueToday := store.GetDueTopics(today)
		dueWeek := store.GetDueTopics(week)

		fmt.Println("Recall Status")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf("Total topics:    %d\n", len(topics))
		fmt.Printf("Due today:       %d\n", len(dueToday))
		fmt.Printf("Due this week:   %d\n", len(dueWeek))

		if len(dueToday) > 0 {
			fmt.Println("\nDue Today:")
			for _, t := range dueToday {
				fmt.Printf("  - %s [%s]\n", t.Title, strings.Join(t.Tags, ", "))
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
