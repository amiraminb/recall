package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "recall",
	Short: "Spaced repetition for your wiki notes",
	Long: `Recall stops what you have studied from fading.

It watches a directory of markdown notes, tracks how well you remember each
one, and tells you which are worth revisiting today. You study in your own
editor; recall only decides what and when.

A note joins your review schedule when its YAML frontmatter says so:

  ---
  id: Docker Networking
  tags: [devops, docker]
  review: true
  ---

Notes without "review: true" are ignored, so the rest of your wiki is left
alone. Recall never modifies your notes.

Scheduling uses FSRS, the algorithm behind modern Anki. A review asks how well
you remembered: 1 Again, 2 Hard, 3 Good, 4 Easy. Strong recall pushes the next
review further out, up to a year. A lapse pulls it back to a few days.

Reviewing late is fine and is accounted for, not punished: recalling something
you had nearly forgotten proves more than an on-time review, so it earns a
longer interval than if you had been prompt.

Review data lives in <wiki>/.srs/reviews.json.`,
	Example: `  Set up once:
    recall init ~/wiki                 point recall at your notes
    recall scan                        pick up newly flagged notes
    recall read "Docker Networking"    log a first study session

  Then each day:
    recall due                         see what needs revisiting
    recall note "Docker Networking"    reread a topic
    recall review "Docker Networking"  rate recall, set the next date`,
	SilenceUsage:      true,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false},
}

var commandGroups = []*cobra.Group{
	{ID: "setup", Title: "Setup:"},
	{ID: "reviewing", Title: "Reviewing:"},
	{ID: "notes", Title: "Working with a topic:"},
	{ID: "managing", Title: "Managing what is tracked:"},
}

func main() {
	rootCmd.AddGroup(commandGroups...)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
