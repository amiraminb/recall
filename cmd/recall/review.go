package main

import (
	"fmt"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review [topic-title]",
	Short: "Mark a topic as reviewed",
	Args:  cobra.ExactArgs(1),
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

		fmt.Printf("Reviewing: %s\n\n", topic.Title)
		fmt.Println("How well did you recall this topic?")
		fmt.Println("  1) Again - Forgot completely")
		fmt.Println("  2) Hard  - Difficult to recall")
		fmt.Println("  3) Good  - Recalled with effort")
		fmt.Println("  4) Easy  - Recalled effortlessly")
		fmt.Print("\nRating [1-4]: ")

		var input int
		fmt.Scanln(&input)

		if input < 1 || input > 4 {
			return fmt.Errorf("invalid rating: %d", input)
		}

		rating := fsrs.Rating(input)
		scheduler := fsrs.NewScheduler()
		topic.Card = scheduler.Review(topic.Card, rating, time.Now())

		if err := store.UpdateTopic(topic); err != nil {
			return err
		}
		if err := store.AddReview(topic.ID, rating); err != nil {
			return err
		}

		fmt.Printf("\nReviewed! Next review: %s\n", topic.Card.Due.Format("Jan 2, 2006"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}
