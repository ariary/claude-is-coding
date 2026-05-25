package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
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

// Load returns sessions sorted by LastUsed descending.
// limit == 0 means no limit.
func Load(limit int) ([]Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}

	sessionsDir := filepath.Join(home, ".claude", "sessions")
	projectsDir := filepath.Join(home, ".claude", "projects")

	active, err := loadActiveSessions(sessionsDir)
	if err != nil {
		return nil, err
	}

	historical, err := loadHistoricalSessions(projectsDir, limit)
	if err != nil {
		return nil, err
	}

	return mergeAndDedup(historical, active, limit), nil
}

// mergeAndDedup merges historical and active sessions, deduplicates by ID
// (active wins), sorts by LastUsed descending, applies limit (0 = no limit).
func mergeAndDedup(historical, active []Session, limit int) []Session {
	seen := make(map[string]Session)
	for _, s := range historical {
		seen[s.ID] = s
	}
	for _, s := range active {
		seen[s.ID] = s
	}

	merged := make([]Session, 0, len(seen))
	for _, s := range seen {
		merged = append(merged, s)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].LastUsed > merged[j].LastUsed
	})

	if limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}
	return merged
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
	if secs < 0 {
		secs = 0
	}
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
		// Verify the process is still alive
		if f.PID > 0 {
			if err := syscall.Kill(f.PID, 0); err == syscall.ESRCH {
				continue // process is gone
			}
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

// jsonlEntry is a generic shape for entries in a .jsonl file.
type jsonlEntry struct {
	Type        string `json:"type"`
	CustomTitle string `json:"customTitle"`
	SessionID   string `json:"sessionId"`
	Cwd         string `json:"cwd"`
	Timestamp   string `json:"timestamp"`
}

// loadHistoricalSessions reads projectsDir and returns sessions sorted by mtime desc.
// limit == 0 means no limit.
func loadHistoricalSessions(projectsDir string, limit int) ([]Session, error) {
	projEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading projects dir: %w", err)
	}

	type sessionFile struct {
		path  string
		mtime time.Time
	}
	var files []sessionFile

	for _, proj := range projEntries {
		if !proj.IsDir() {
			continue
		}
		projPath := filepath.Join(projectsDir, proj.Name())
		jsonls, err := filepath.Glob(filepath.Join(projPath, "*.jsonl"))
		if err != nil {
			continue
		}
		for _, p := range jsonls {
			info, err := os.Stat(p)
			if err != nil {
				continue
			}
			files = append(files, sessionFile{path: p, mtime: info.ModTime()})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].mtime.After(files[j].mtime)
	})

	var result []Session
	for _, f := range files {
		s, err := parseSessionFile(f.path, f.mtime)
		if err != nil || s == nil {
			continue
		}
		result = append(result, *s)
	}
	return result, nil
}

// parseSessionFile reads a .jsonl file and extracts Session data.
func parseSessionFile(path string, mtime time.Time) (*Session, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	base := filepath.Base(path)
	sessionID := strings.TrimSuffix(base, ".jsonl")
	if sessionID == "" {
		return nil, nil
	}

	var name, cwd string
	var firstTS, lastTS time.Time

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 1024*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var entry jsonlEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry.Type == "custom-title" && entry.CustomTitle != "" {
			name = entry.CustomTitle
		}
		if entry.Cwd != "" && cwd == "" {
			cwd = entry.Cwd
		}
		if entry.Timestamp != "" {
			t, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
			if err == nil {
				if firstTS.IsZero() {
					firstTS = t
				}
				lastTS = t
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}

	var duration int64
	if !firstTS.IsZero() && !lastTS.IsZero() && lastTS.After(firstTS) {
		duration = int64(lastTS.Sub(firstTS).Seconds())
	}

	return &Session{
		ID:       sessionID,
		Name:     name,
		Cwd:      cwd,
		Project:  projectFromCwd(cwd),
		LastUsed: mtime.Unix(),
		Duration: duration,
		Active:   false,
	}, nil
}
