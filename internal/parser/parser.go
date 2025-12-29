package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type ParsedTopic struct {
	Title string
	File  string
	Tags  []string
}

// Frontmatter represents the YAML frontmatter structure
type Frontmatter struct {
	ID     string   `yaml:"id"`
	Tags   []string `yaml:"tags"`
	Review bool     `yaml:"review"`
}

var headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// parseFrontmatter extracts YAML frontmatter from file content
func parseFrontmatter(scanner *bufio.Scanner) (*Frontmatter, error) {
	// Check for opening ---
	if !scanner.Scan() {
		return nil, nil
	}
	firstLine := scanner.Text()
	if firstLine != "---" {
		// No frontmatter, return nil (not an error)
		return nil, nil
	}

	// Collect YAML content until closing ---
	var yamlContent strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		yamlContent.WriteString(line)
		yamlContent.WriteString("\n")
	}

	// Parse YAML
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlContent.String()), &fm); err != nil {
		return nil, err
	}

	return &fm, nil
}

// findFirstHeading scans remaining content for first heading
func findFirstHeading(scanner *bufio.Scanner) string {
	for scanner.Scan() {
		line := scanner.Text()
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			return strings.TrimSpace(matches[2])
		}
	}
	return ""
}

// getTitleFromFilename extracts title from filename (without .md extension)
func getTitleFromFilename(filePath string) string {
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, ".md")
}

func ScanFile(filePath string) ([]ParsedTopic, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Parse frontmatter
	fm, err := parseFrontmatter(scanner)
	if err != nil {
		return nil, err
	}

	// If no frontmatter or review is not true, skip this file
	if fm == nil || !fm.Review {
		return nil, nil
	}

	// Use id from frontmatter, fallback to filename, then first heading
	title := fm.ID
	if title == "" {
		title = getTitleFromFilename(filePath)
	}
	if title == "" {
		title = findFirstHeading(scanner)
	}

	topic := ParsedTopic{
		Title: title,
		File:  filePath,
		Tags:  fm.Tags,
	}

	return []ParsedTopic{topic}, nil
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
