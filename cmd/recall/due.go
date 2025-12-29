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
	Short: "Show status and topics due for review",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		tag, _ := cmd.Flags().GetString("tag")
		week, _ := cmd.Flags().GetBool("week")

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		weekEnd := today.AddDate(0, 0, 7)

		// Summary
		allTopics := store.GetAllTopics()
		dueToday := store.GetDueTopics(today)
		dueWeek := store.GetDueTopics(weekEnd)

		fmt.Printf("Topics: %d | Due today: %d | Due this week: %d\n\n",
			len(allTopics), len(dueToday), len(dueWeek))

		// Get topics based on flags
		until := today
		if week {
			until = weekEnd
		}
		topics := store.GetDueTopics(until)

		// Filter by tag if specified
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

		label := "Due today:"
		if week {
			label = "Due this week:"
		}
		fmt.Println(label)

		for _, t := range topics {
			days := int(time.Until(t.Card.Due).Hours() / 24)
			dueStr := "today"
			if days < 0 {
				dueStr = fmt.Sprintf("%d days overdue", -days)
			} else if days > 0 {
				dueStr = fmt.Sprintf("in %d days", days)
			}

			tagStr := ""
			if len(t.Tags) > 0 {
				tagStr = fmt.Sprintf(" (%s)", strings.Join(t.Tags, ", "))
			}
			fmt.Printf("  - %s%s - %s\n", t.Title, tagStr, dueStr)
		}

		return nil
	},
}

func init() {
	dueCmd.Flags().String("tag", "", "Filter by tag")
	dueCmd.Flags().Bool("week", false, "Show topics due this week")
	rootCmd.AddCommand(dueCmd)
}
