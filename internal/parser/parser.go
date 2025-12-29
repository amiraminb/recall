package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ParsedTopic struct {
	Title string
	File  string
	Tags  []string
}

var headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s+@review(.*)$`)

var tagRegex = regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)

func ScanFile(filePath string) ([]ParsedTopic, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var topics []ParsedTopic
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		matches := headingRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		title := strings.TrimSpace(matches[2])
		remainder := matches[3]

		// Extract tags from remainder
		var tags []string
		tagMatches := tagRegex.FindAllStringSubmatch(remainder, -1)
		for _, tm := range tagMatches {
			tags = append(tags, tm[1])
		}

		topics = append(topics, ParsedTopic{
			Title: title,
			File:  filePath,
			Tags:  tags,
		})
	}

	return topics, scanner.Err()
}

func ScanDirectory(dir string) ([]ParsedTopic, error) {
	var allTopics []ParsedTopic

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Only process .md files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			topics, err := ScanFile(path)
			if err != nil {
				return err
			}
			allTopics = append(allTopics, topics...)
		}

		return nil
	})

	return allTopics, err
}
