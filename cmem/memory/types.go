package memory

// Entry represents a single Claude memory file.
type Entry struct {
	Project     string // human-readable project name (last path segment)
	ProjectSlug string // full directory slug under ~/.claude/projects/
	File        string // filename (e.g. feedback_no_coauthor.md)
	Name        string // from frontmatter: name
	Description string // from frontmatter: description
	Type        string // from frontmatter: type (user|feedback|project|reference)
	Body        string // content after the closing --- of frontmatter
}
