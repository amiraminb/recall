package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "recall",
	Short: "Remember what you learn",
	Long: `Recall helps you remember what you've studied.

Add a topic after learning it. Recall schedules your first review for tomorrow,
then uses FSRS to adjust future reviews based on how well you remember.

Run recall with no arguments to see today's review queue.`,
	Example: `  recall add "Docker Networking"
  recall add "Go interfaces" --source "https://go.dev/tour/methods/9" --tag go
  recall
  recall review`,
	Args:              cobra.NoArgs,
	RunE:              runDue,
	SilenceUsage:      true,
	SilenceErrors:     true,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false},
}

var commandGroups = []*cobra.Group{
	{ID: "reviewing", Title: "Daily practice:"},
	{ID: "managing", Title: "Your topics:"},
}

func main() {
	rootCmd.AddGroup(commandGroups...)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
