package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

// projectFromCwd returns the last 2 path segments of cwd for display.
func projectFromCwd(cwd string) string {
	if cwd == "" {
		return ""
	}
	dir := filepath.Clean(cwd)
	base := filepath.Base(dir)
	parent := filepath.Base(filepath.Dir(dir))
	if parent == "." || parent == "/" {
		return base
	}
	return parent + "/" + base
}

// FormatAge returns a human-readable age string for a Unix timestamp.
func FormatAge(ts int64) string {
	secs := time.Now().Unix() - ts
	switch {
	case secs < 60:
		return fmt.Sprintf("%ds ago", secs)
	case secs < 3600:
		return fmt.Sprintf("%dm ago", secs/60)
	case secs < 86400:
		return fmt.Sprintf("%dh ago", secs/3600)
	default:
		return fmt.Sprintf("%dd ago", secs/86400)
	}
}

// FormatDuration returns a compact human-readable duration string.
func FormatDuration(secs int64) string {
	if secs <= 0 {
		return ""
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh%dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", s)
	}
}

type activeSessionFile struct {
	PID       int    `json:"pid"`
	SessionID string `json:"sessionId"`
	Cwd       string `json:"cwd"`
	StartedAt int64  `json:"startedAt"` // milliseconds
	Name      string `json:"name"`
}

// loadActiveSessions reads all .json files in dir and returns active Sessions.
func loadActiveSessions(dir string) ([]Session, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sessions dir: %w", err)
	}

	now := time.Now().Unix()
	var result []Session
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var f activeSessionFile
		if err := json.Unmarshal(data, &f); err != nil || f.SessionID == "" {
			continue
		}
		startedSec := f.StartedAt / 1000
		s := Session{
			ID:       f.SessionID,
			Name:     f.Name,
			Cwd:      f.Cwd,
			Project:  projectFromCwd(f.Cwd),
			LastUsed: now,
			Duration: now - startedSec,
			Active:   true,
		}
		result = append(result, s)
	}
	return result, nil
}
