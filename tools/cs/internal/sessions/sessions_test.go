package sessions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectFromCwd(t *testing.T) {
	cases := []struct {
		cwd  string
		want string
	}{
		{"/Users/foo/project/ariary/claude-is-coding", "ariary/claude-is-coding"},
		{"/Users/foo/project/myrepo", "project/myrepo"},
		{"/myrepo", "myrepo"},
		{"", ""},
	}
	for _, c := range cases {
		got := projectFromCwd(c.cwd)
		if got != c.want {
			t.Errorf("projectFromCwd(%q) = %q, want %q", c.cwd, got, c.want)
		}
	}
}

func TestFormatAge(t *testing.T) {
	now := time.Now().Unix()
	cases := []struct {
		ts   int64
		want string
	}{
		{now - 30, "30s ago"},
		{now - 90, "1m ago"},
		{now - 3700, "1h ago"},
		{now - 86500, "1d ago"},
		{now - 7*86400 + 100, "6d ago"},
	}
	for _, c := range cases {
		got := FormatAge(c.ts)
		if got != c.want {
			t.Errorf("FormatAge(%d) = %q, want %q", c.ts, got, c.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		secs int64
		want string
	}{
		{0, ""},
		{45, "45s"},
		{90, "1m"},
		{3660, "1h1m"},
		{7200, "2h"},
	}
	for _, c := range cases {
		got := FormatDuration(c.secs)
		if got != c.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", c.secs, got, c.want)
		}
	}
}

func TestLoadActiveSessions(t *testing.T) {
	dir := t.TempDir()
	pid := os.Getpid()
	content := fmt.Sprintf(`{"pid":%d,"sessionId":"aaa-bbb","cwd":"/tmp/myproject","startedAt":%d,"kind":"interactive","entrypoint":"cli","name":"my-test-session"}`,
		pid, (time.Now().Unix()-300)*1000)
	if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.json", pid)), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := loadActiveSessions(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 session, got %d", len(result))
	}
	s := result[0]
	if s.ID != "aaa-bbb" {
		t.Errorf("ID = %q, want %q", s.ID, "aaa-bbb")
	}
	if s.Name != "my-test-session" {
		t.Errorf("Name = %q, want %q", s.Name, "my-test-session")
	}
	if s.Cwd != "/tmp/myproject" {
		t.Errorf("Cwd = %q, want %q", s.Cwd, "/tmp/myproject")
	}
	if !s.Active {
		t.Error("Active = false, want true")
	}
	if s.Duration < 290 || s.Duration > 310 {
		t.Errorf("Duration = %d, want ~300", s.Duration)
	}
	if s.Project != "tmp/myproject" {
		t.Errorf("Project = %q, want %q", s.Project, "tmp/myproject")
	}
}

func TestLoadHistoricalSessions(t *testing.T) {
	root := t.TempDir()
	projDir := filepath.Join(root, "-Users-foo-myproject")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}

	sessionID := "abc-123"
	lines := []string{
		`{"type":"custom-title","customTitle":"my-feature","sessionId":"abc-123"}`,
		`{"type":"user","cwd":"/Users/foo/myproject","sessionId":"abc-123","timestamp":"2026-01-01T10:00:00.000Z"}`,
		`{"type":"assistant","sessionId":"abc-123","timestamp":"2026-01-01T10:30:00.000Z"}`,
	}
	content := strings.Join(lines, "\n") + "\n"
	jsonlPath := filepath.Join(projDir, sessionID+".jsonl")
	if err := os.WriteFile(jsonlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := loadHistoricalSessions(root, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 session, got %d", len(result))
	}
	s := result[0]
	if s.ID != "abc-123" {
		t.Errorf("ID = %q, want %q", s.ID, "abc-123")
	}
	if s.Name != "my-feature" {
		t.Errorf("Name = %q, want %q", s.Name, "my-feature")
	}
	if s.Cwd != "/Users/foo/myproject" {
		t.Errorf("Cwd = %q, want %q", s.Cwd, "/Users/foo/myproject")
	}
	if s.Duration != 1800 {
		t.Errorf("Duration = %d, want 1800", s.Duration)
	}
}

func TestMergeAndDedupActiveWins(t *testing.T) {
	active := Session{ID: "x", Name: "active-name", Cwd: "/a", Active: true, LastUsed: 999}
	historical := Session{ID: "x", Name: "old-name", Cwd: "/a", Active: false, LastUsed: 500}

	merged := mergeAndDedup([]Session{historical}, []Session{active}, 0)
	if len(merged) != 1 {
		t.Fatalf("expected 1 session after dedup, got %d", len(merged))
	}
	if !merged[0].Active {
		t.Error("merged session should be active")
	}
	if merged[0].Name != "active-name" {
		t.Errorf("Name = %q, want %q", merged[0].Name, "active-name")
	}
}

func TestMergeAndDedupSortsByLastUsedDesc(t *testing.T) {
	historical := []Session{
		{ID: "a", LastUsed: 100},
		{ID: "b", LastUsed: 300},
		{ID: "c", LastUsed: 200},
	}
	merged := mergeAndDedup(historical, nil, 0)
	if len(merged) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(merged))
	}
	if merged[0].ID != "b" || merged[1].ID != "c" || merged[2].ID != "a" {
		t.Errorf("wrong sort order: got [%s %s %s]", merged[0].ID, merged[1].ID, merged[2].ID)
	}
}

func TestMergeAndDedupLimit(t *testing.T) {
	historical := []Session{
		{ID: "a", LastUsed: 100},
		{ID: "b", LastUsed: 300},
		{ID: "c", LastUsed: 200},
	}
	merged := mergeAndDedup(historical, nil, 2)
	if len(merged) != 2 {
		t.Fatalf("expected 2 sessions with limit=2, got %d", len(merged))
	}
	if merged[0].ID != "b" || merged[1].ID != "c" {
		t.Errorf("wrong sessions after limit: got [%s %s]", merged[0].ID, merged[1].ID)
	}
}
