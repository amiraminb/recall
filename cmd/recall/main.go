package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "recall",
	Short: "Spaced repetition for your wiki notes",
	Long:  `Recall helps you remember what you learn by scheduling reviews using the FSRS algorithm.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
