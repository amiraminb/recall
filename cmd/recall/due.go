package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var dueCmd = &cobra.Command{
	GroupID: "reviewing",
	Use:     "due",
	Short:   "See what's due, most overdue first",
	Long: `Show today's review queue, or the next seven days with --week.

Examples:
  recall
  recall due --week
  recall due --tag go`,
	Args: cobra.NoArgs,
	RunE: runDue,
}

func runDue(cmd *cobra.Command, args []string) error {
	store, err := getStorage()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if len(store.GetAllTopics()) == 0 {
		fmt.Fprintln(out, "Add something you've studied to start remembering it:\n\n  recall add \"Docker Networking\"\n\nYour first review will be tomorrow.")
		return nil
	}
	tag, _ := cmd.Flags().GetString("tag")
	week, _ := cmd.Flags().GetBool("week")
	now := time.Now()
	today := endOfDay(now)
	weekEnd := endOfDay(now.AddDate(0, 0, 6))
	allTopics := filterTopics(store.GetAllTopics(), tag)
	dueToday := filterTopics(store.GetDueTopics(today), tag)
	dueWeek := filterTopics(store.GetDueTopics(weekEnd), tag)
	fmt.Fprintf(out, "Topics: %d | Due today: %d | Due next 7 days: %d\n\n", len(allTopics), len(dueToday), len(dueWeek))
	topics := dueToday
	if week {
		topics = dueWeek
	}
	if len(topics) == 0 {
		fmt.Fprintln(out, "All caught up!")
		if len(allTopics) > 0 {
			fmt.Fprintf(out, "Next review: %s (%s)\n", allTopics[0].Title, allTopics[0].Card.Due.Format("Jan 2, 2006"))
		} else if tag != "" {
			fmt.Fprintf(out, "No topics tagged %q.\n", tag)
		}
		return nil
	}
	table := tablewriter.NewWriter(out)
	table.Header("Topic", "Due")
	for _, topic := range topics {
		days := calendarDaysUntil(topic.Card.Due, now)
		due := "today"
		if days < 0 {
			due = fmt.Sprintf("%d days overdue", -days)
		} else if days > 0 {
			due = fmt.Sprintf("in %d days", days)
		}
		table.Append(truncateText(topic.Title, maxTitleWidth), colorDue(days, due))
	}
	table.Render()
	if len(dueToday) > 0 {
		if tag != "" {
			fmt.Fprintf(out, "\nStart reviewing: recall review --tag %q\n", tag)
		} else {
			fmt.Fprintln(out, "\nStart reviewing: recall review")
		}
	}
	return nil
}

func calendarDaysUntil(due, now time.Time) int {
	due = due.In(now.Location())
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return int(dueDay.Sub(today).Hours() / 24)
}

func colorDue(days int, text string) string {
	if days < 0 {
		return color.New(color.FgRed).Sprint(text)
	}
	if days == 0 {
		return color.New(color.FgMagenta).Sprint(text)
	}
	return color.New(color.FgCyan).Sprint(text)
}

const maxTitleWidth = 48

func truncateText(text string, max int) string {
	r := []rune(text)
	if len(r) <= max {
		return text
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

func init() {
	for _, cmd := range []*cobra.Command{rootCmd, dueCmd} {
		cmd.Flags().String("tag", "", "Filter by tag")
		cmd.Flags().Bool("week", false, "Show topics due over the next seven days")
	}
	rootCmd.AddCommand(dueCmd)
}
