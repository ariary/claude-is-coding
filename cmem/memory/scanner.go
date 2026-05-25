package memory

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
}

// Scan walks baseDir/*/memory/*.md and returns all parsed entries.
// baseDir defaults to ~/.claude/projects if empty.
func Scan(baseDir string) ([]Entry, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		baseDir = filepath.Join(home, ".claude", "projects")
	}

	pattern := filepath.Join(baseDir, "*", "memory", "*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, path := range matches {
		filename := filepath.Base(path)
		if filename == "MEMORY.md" {
			continue
		}

		slug := filepath.Base(filepath.Dir(filepath.Dir(path)))
		entry, err := parseEntry(path, slug)
		if err != nil {
			// skip malformed files silently
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func parseEntry(path, slug string) (Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, err
	}

	content := string(data)
	fm, body, err := parseFrontmatter(content)
	if err != nil {
		return Entry{}, err
	}

	return Entry{
		Project:     ProjectNameFromSlug(slug),
		ProjectSlug: slug,
		File:        filepath.Base(path),
		Name:        fm.Name,
		Description: fm.Description,
		Type:        fm.Type,
		Body:        strings.TrimSpace(body),
	}, nil
}

// parseFrontmatter splits "---\n<yaml>\n---\n<body>" into frontmatter and body.
func parseFrontmatter(content string) (frontmatter, string, error) {
	const delim = "---"
	// strip leading newlines/spaces
	content = strings.TrimLeft(content, "\r\n")
	if !strings.HasPrefix(content, delim) {
		return frontmatter{}, content, nil
	}
	// skip the opening ---
	rest := content[len(delim):]
	idx := strings.Index(rest, "\n"+delim)
	if idx == -1 {
		return frontmatter{}, content, nil
	}
	yamlPart := rest[:idx]
	body := rest[idx+1+len(delim):]
	// strip leading newline from body
	body = strings.TrimPrefix(body, "\n")

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return frontmatter{}, "", err
	}
	return fm, body, nil
}

// ProjectNameFromSlug returns the last meaningful segment of a project slug.
// e.g. "-Users-antoine-project-ariary-soa" -> "soa"
func ProjectNameFromSlug(slug string) string {
	// slug starts with "-", path segments were joined with "-"
	// strip leading "-"
	s := strings.TrimPrefix(slug, "-")
	parts := strings.Split(s, "-")
	// find last non-empty part
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return slug
}
