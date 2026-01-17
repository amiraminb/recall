package main

import (
	"fmt"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <topic-title>",
	Short: "Mark a topic as read for the first time",
	Long: `Record your first reading of a topic and schedule the first review.

Use this after you've read and understood a topic for the first time.
You'll be asked how well you understood it, which determines when
your first review will be scheduled.

Understanding levels:
  1 - Didn't understand, review tomorrow
  2 - Partially understood, review in 2 days
  3 - Understood well, review in 4 days
  4 - Mastered it, review in 7 days

Example:
  recall read "Docker Networking"`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTopicTitles,
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

		// Check if already read (not New state)
		if topic.Card.State != fsrs.New {
			return fmt.Errorf("topic already read, use 'recall review' instead")
		}

		fmt.Printf("First read: %s\n\n", topic.Title)
		fmt.Println("How well did you understand this topic?")
		fmt.Println("  1) Didn't understand - review tomorrow")
		fmt.Println("  2) Partially understood - review in 2 days")
		fmt.Println("  3) Understood well - review in 4 days")
		fmt.Println("  4) Mastered it - review in 7 days")
		fmt.Print("\nUnderstanding [1-4]: ")

		var input int
		fmt.Scanln(&input)

		if input < 1 || input > 4 {
			return fmt.Errorf("invalid input: %d", input)
		}

		// Map understanding to days until first review
		daysMap := map[int]int{1: 1, 2: 2, 3: 4, 4: 7}
		days := daysMap[input]

		// Update card state
		now := time.Now()
		topic.Card.State = fsrs.Learning
		topic.Card.Due = now.AddDate(0, 0, days)
		topic.Card.LastReview = now
		topic.Card.Reps = 1

		if err := store.UpdateTopic(topic); err != nil {
			return err
		}

		// Log as first read (using Good rating as placeholder)
		if err := store.AddReview(topic.ID, fsrs.Rating(input)); err != nil {
			return err
		}

		fmt.Printf("\nMarked as read! First review: %s\n", topic.Card.Due.Format("Jan 2, 2006"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
