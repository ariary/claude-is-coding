package memory

import (
	"log"
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
			log.Printf("cmem: skipping %s: %v\n", path, err)
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
	content = strings.TrimLeft(content, "\r\n")
	// get first line and trim trailing whitespace before checking delimiter
	firstLine, _, _ := strings.Cut(content, "\n")
	if strings.TrimRight(firstLine, " \t\r") != delim {
		return frontmatter{}, content, nil
	}
	// skip the opening --- line
	rest := content[len(firstLine):]
	rest = strings.TrimPrefix(rest, "\n")
	// find closing ---
	yamlPart, body, found := strings.Cut(rest, "\n"+delim)
	if !found {
		return frontmatter{}, content, nil
	}
	body = strings.TrimPrefix(body, "\n")

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return frontmatter{}, "", err
	}
	return fm, body, nil
}

// ProjectNameFromSlug returns a human-readable name from a project slug.
// The slug is an absolute path with "/" replaced by "-" (e.g. "-Users-foo-project-myapp").
// It strips the home directory prefix and returns the remainder, or falls back to the last token.
func ProjectNameFromSlug(slug string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		// encode home path as slug: "/Users/foo" -> "-Users-foo"
		homeSlug := "-" + strings.ReplaceAll(strings.TrimPrefix(home, "/"), "/", "-")
		if rest, ok := strings.CutPrefix(slug, homeSlug); ok {
			rest = strings.TrimPrefix(rest, "-")
			if rest != "" {
				return rest
			}
		}
	}
	// fallback: last non-empty token
	s := strings.TrimPrefix(slug, "-")
	parts := strings.Split(s, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return slug
}
