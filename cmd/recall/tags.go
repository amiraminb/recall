package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags with topic counts",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}

		tags := store.GetAllTags()
		if len(tags) == 0 {
			fmt.Println("No tags found.")
			return nil
		}

		fmt.Println("Tags:")
		for tag, count := range tags {
			fmt.Printf("  #%s (%d)\n", tag, count)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
