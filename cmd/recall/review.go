package main

import (
	"fmt"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review <topic-title>",
	Short: "Review a topic and rate your recall",
	Long: `Mark a topic as reviewed and rate how well you remembered it.

After reading your notes, use this command to log the review. You'll be asked
to rate your recall from 1-4. The FSRS algorithm then schedules the next
review based on your rating.

Ratings:
  1 (Again) - Forgot completely, reset interval
  2 (Hard)  - Struggled to recall, shorter interval
  3 (Good)  - Recalled with some effort, normal interval
  4 (Easy)  - Instant recall, longer interval

Example:
  recall review "Docker Networking"`,
	Args: cobra.ExactArgs(1),
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

		// Check if topic hasn't been read yet
		if topic.Card.State == fsrs.New {
			return fmt.Errorf("topic not yet read, use 'recall read \"%s\"' first", title)
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
