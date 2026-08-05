package main

import (
	"fmt"
	"os"

	"github.com/amiraminb/recall/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <wiki-path>",
	Short: "Initialize recall with your wiki path",
	Long: `Set up recall by specifying your wiki directory.

This creates a config file at ~/.config/recall/config.json. Review data is
stored in a .srs directory created inside your wiki on first use.

Example:
  recall init ~/wiki
  recall init ~/notes/obsidian`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wikiPath, err := expandPath(args[0])
		if err != nil {
			return err
		}

		info, err := os.Stat(wikiPath)
		if err != nil {
			return fmt.Errorf("wiki path does not exist: %s", wikiPath)
		}
		if !info.IsDir() {
			return fmt.Errorf("wiki path is not a directory: %s", wikiPath)
		}

		cfg := &config.Config{WikiPath: wikiPath}
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Initialized recall with wiki: %s\n", wikiPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
