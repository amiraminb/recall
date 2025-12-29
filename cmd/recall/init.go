package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amiraminb/recall/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [wiki-path]",
	Short: "Initialize recall with your wiki path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wikiPath := args[0]

		if strings.HasPrefix(wikiPath, "~") {
			home, _ := os.UserHomeDir()
			wikiPath = filepath.Join(home, wikiPath[1:])
		}

		if _, err := os.Stat(wikiPath); os.IsNotExist(err) {
			return fmt.Errorf("wiki path does not exist: %s", wikiPath)
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
