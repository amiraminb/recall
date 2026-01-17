package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var markdownLinkRegex = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

var openCmd = &cobra.Command{
	Use:   "open <topic-title>",
	Short: "Open the first link in a topic",
	Long: `Open the first markdown link in a topic file using your default browser.

Example:
  recall open "Kubernetes Architecture"`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTopicTitles,
	RunE: func(cmd *cobra.Command, args []string) error {
		wikiPath, err := getWikiPath()
		if err != nil {
			return err
		}

		title := args[0]
		filePath, err := findTopicFile(wikiPath, title)
		if err != nil {
			return err
		}
		if filePath == "" {
			return fmt.Errorf("topic not found: %s", title)
		}

		link, err := findFirstLink(filePath)
		if err != nil {
			return err
		}
		if link == "" {
			return fmt.Errorf("no link found in topic: %s", title)
		}

		if err := openURL(link); err != nil {
			return err
		}

		fmt.Printf("Opening %s\n", link)
		return nil
	},
}

func findTopicFile(wikiPath, title string) (string, error) {
	var matches []string
	searchName := strings.TrimSpace(title)
	searchNameLower := strings.ToLower(searchName)

	err := filepath.WalkDir(wikiPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		baseName := strings.TrimSuffix(d.Name(), ".md")
		baseNameLower := strings.ToLower(baseName)
		if baseNameLower != searchNameLower {
			return nil
		}

		matches = append(matches, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple files matched title: %s", searchName)
	}

	return matches[0], nil
}

func completeTopicTitles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wikiPath, err := getWikiPath()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	titles, err := listWikiTitles(wikiPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	prefix := strings.ToLower(toComplete)
	var matches []string
	for _, title := range titles {
		if strings.HasPrefix(strings.ToLower(title), prefix) {
			matches = append(matches, title)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

func findFirstLink(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := markdownLinkRegex.FindStringSubmatch(scanner.Text())
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1]), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
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
