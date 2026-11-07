package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	GroupID: "managing",
	Use:     "edit <topic-title>",
	Short:   "Update a topic's title, source, or tags",
	Long: `Update topic details while keeping its review schedule and history.
Use an empty value to clear a source or tags. --tag replaces the tags.

Examples:
  recall edit "Go interfaces" --source "https://go.dev/tour/methods/9"
  recall edit "Docker" --title "Docker Networking" --tag devops,networking
  recall edit "Docker Networking" --source ""`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTopicTitles,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("title") && !cmd.Flags().Changed("source") && !cmd.Flags().Changed("tag") {
			return fmt.Errorf("choose what to update: --title, --source, or --tag")
		}
		store, err := getStorage()
		if err != nil {
			return err
		}
		topic, err := resolveTopic(store, args[0])
		if err != nil {
			return err
		}
		updated := *topic
		if cmd.Flags().Changed("title") {
			title, _ := cmd.Flags().GetString("title")
			title = strings.TrimSpace(title)
			if title == "" {
				return fmt.Errorf("topic title cannot be empty")
			}
			for _, match := range store.FindTopicsByTitle(title) {
				if match.ID != topic.ID {
					return fmt.Errorf("topic %q already exists", title)
				}
			}
			updated.Title = title
		}
		if cmd.Flags().Changed("source") {
			source, _ := cmd.Flags().GetString("source")
			updated.Source = strings.TrimSpace(source)
		}
		if cmd.Flags().Changed("tag") {
			tags, _ := cmd.Flags().GetStringSlice("tag")
			updated.Tags = cleanTags(tags)
		}
		if err := store.UpdateTopic(&updated); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s\n", updated.Title)
		return nil
	},
}

func init() {
	editCmd.Flags().String("title", "", "New topic title")
	editCmd.Flags().String("source", "", "Source URL or reference")
	editCmd.Flags().StringSlice("tag", nil, "Replace tags, comma-separated or repeated")
	rootCmd.AddCommand(editCmd)
}
