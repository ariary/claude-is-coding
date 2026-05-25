package sessions

// Session holds all display and launch data for one Claude Code session.
type Session struct {
	ID       string
	Name     string // from custom-title entry in .jsonl
	Cwd      string // working directory
	Project  string // last 2 segments of Cwd, for display
	LastUsed int64  // Unix timestamp (mtime of .jsonl file)
	Duration int64  // seconds; active: time.Since(startedAt); historical: mtime-first-entry-time
	Active   bool   // true if PID exists in ~/.claude/sessions/
}

// Load returns up to limit sessions sorted by LastUsed descending.
// If limit == 0, all sessions are returned.
func Load(limit int) ([]Session, error) {
	return nil, nil
}
