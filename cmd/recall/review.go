package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/amiraminb/recall/internal/fsrs"
	"github.com/amiraminb/recall/internal/storage"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	GroupID: "reviewing",
	Use:     "review [topic-title]",
	Short:   "Work through today's queue or review one topic",
	Long: `Try explaining each topic from memory before checking your source.
Then rate how well you recalled it. FSRS schedules the next review.

Without a title, works through today's topics, most overdue first.
Each rating is saved immediately. Skip a topic with s or end the session with q.

Ratings: 1 Again (forgot), 2 Hard, 3 Good, 4 Easy.

Examples:
  recall review
  recall review --tag go
  recall review "Docker Networking"
  recall review "Docker Networking" --rating 3`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeTopicTitles,
	RunE: func(cmd *cobra.Command, args []string) error {
		rating, _ := cmd.Flags().GetInt("rating")
		if cmd.Flags().Changed("rating") && (len(args) != 1 || !fsrs.Rating(rating).Valid()) {
			return fmt.Errorf("use --rating 1-4 with a single topic title")
		}
		tag, _ := cmd.Flags().GetString("tag")
		if len(args) == 1 && cmd.Flags().Changed("tag") {
			return fmt.Errorf("use --tag with the review queue, or pass a single topic title")
		}
		store, err := getStorage()
		if err != nil {
			return err
		}
		var topics []storage.Topic
		if len(args) == 1 {
			topic, err := resolveTopic(store, args[0])
			if err != nil {
				return err
			}
			if rating != 0 {
				if err := store.RecordReview(topic.ID, fsrs.Rating(rating), time.Now()); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Reviewed: %s\nNext review: %s\n", topic.Title, topic.Card.Due.Format("Jan 2, 2006"))
				return nil
			}
			topics = []storage.Topic{*topic}
		} else {
			topics = filterTopics(store.GetDueTopics(endOfDay(time.Now())), tag)
		}
		if len(topics) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No topics due for review. Run 'recall' for your schedule.")
			return nil
		}
		return runReviewSession(cmd.InOrStdin(), cmd.OutOrStdout(), store, topics)
	},
}

func runReviewSession(in io.Reader, out io.Writer, store *storage.Storage, topics []storage.Topic) error {
	scanner := bufio.NewScanner(in)
	reviewed, skipped := 0, 0
	finish := func() {
		fmt.Fprintf(out, "\nSession: %d reviewed, %d skipped, %d remaining.\n", reviewed, skipped, len(topics)-reviewed-skipped)
	}
	for i, topic := range topics {
		fmt.Fprintf(out, "\n[%d/%d] %s\n", i+1, len(topics), topic.Title)
		fmt.Fprintln(out, "Explain the key ideas and an example from memory.")
		fmt.Fprintln(out, "Try recalling first, then check your source before rating.")
		if topic.Source != "" {
			fmt.Fprintf(out, "Source: %s\n", topic.Source)
		}
		fmt.Fprintln(out, "\n1 Again (forgot)  2 Hard  3 Good  4 Easy")
		for {
			fmt.Fprint(out, "Rating [1-4], [s]kip, [q]uit: ")
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return err
				}
				finish()
				return nil
			}
			input := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if input == "q" || input == "quit" {
				finish()
				return nil
			}
			if input == "s" || input == "skip" {
				skipped++
				break
			}
			value, err := strconv.Atoi(input)
			if err != nil || !fsrs.Rating(value).Valid() {
				fmt.Fprintln(out, "Enter 1-4, s to skip, or q to quit.")
				continue
			}
			if err := store.RecordReview(topic.ID, fsrs.Rating(value), time.Now()); err != nil {
				return err
			}
			reviewed++
			fmt.Fprintf(out, "Next review: %s\n", store.GetTopic(topic.ID).Card.Due.Format("Jan 2, 2006"))
			break
		}
	}
	finish()
	return nil
}

func init() {
	reviewCmd.Flags().String("tag", "", "Review only topics with this tag")
	reviewCmd.Flags().Int("rating", 0, "Record a rating (1-4) for a single topic without prompting")
	rootCmd.AddCommand(reviewCmd)
}
