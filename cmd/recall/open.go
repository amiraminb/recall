package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	GroupID: "managing",
	Use:     "open <topic-title>",
	Short:   "Open a topic's source URL",
	Long: `Open the source URL you attached to a topic in your default browser.

Example:
  recall open "Go interfaces"`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTopicTitles,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		topic, err := resolveTopic(store, args[0])
		if err != nil {
			return err
		}
		if topic.Source == "" {
			return fmt.Errorf("no source for %q; add one with 'recall edit %q --source <url>'", topic.Title, topic.Title)
		}
		parsed, err := url.Parse(topic.Source)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return fmt.Errorf("source is a reference, not a web URL: %s", topic.Source)
		}
		if err := openURL(topic.Source); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Opening %s\n", topic.Source)
		return nil
	},
}

func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(openCmd)
}
