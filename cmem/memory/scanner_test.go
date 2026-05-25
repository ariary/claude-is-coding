package memory_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ariary/cmem/memory"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "projects")
}

func TestScan_ReturnsEntries(t *testing.T) {
	entries, err := memory.Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestScan_SkipsMEMORYmd(t *testing.T) {
	entries, err := memory.Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, e := range entries {
		if e.File == "MEMORY.md" {
			t.Error("MEMORY.md should be skipped")
		}
	}
}

func TestScan_ParsesFrontmatter(t *testing.T) {
	entries, err := memory.Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	var found *memory.Entry
	for i := range entries {
		if entries[i].File == "user_profile.md" {
			found = &entries[i]
			break
		}
	}
	if found == nil {
		t.Fatal("user_profile.md entry not found")
	}
	if found.Name != "User profile" {
		t.Errorf("Name: got %q, want %q", found.Name, "User profile")
	}
	if found.Type != "user" {
		t.Errorf("Type: got %q, want %q", found.Type, "user")
	}
	if found.Description != "Security engineer, builds CLI tools" {
		t.Errorf("Description: got %q", found.Description)
	}
}

func TestScan_ParsesBody(t *testing.T) {
	entries, err := memory.Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, e := range entries {
		if e.File == "user_profile.md" && e.Body == "" {
			t.Error("Body should not be empty for user_profile.md")
		}
	}
}

func TestScan_ProjectName(t *testing.T) {
	entries, err := memory.Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, e := range entries {
		if e.Project == "" {
			t.Errorf("entry %s has empty Project name", e.File)
		}
	}
}

func TestProjectNameFromSlug(t *testing.T) {
	cases := []struct {
		slug string
		want string
	}{
		{"-Users-antoine-project-ariary-soa", "soa"},
		{"-Users-antoine-project-ariary-ngrep", "ngrep"},
		{"-Users-antoine-project-ariary", "ariary"},
		{"-Users-antoine", "antoine"},
	}
	for _, c := range cases {
		got := memory.ProjectNameFromSlug(c.slug)
		if got != c.want {
			t.Errorf("ProjectNameFromSlug(%q) = %q, want %q", c.slug, got, c.want)
		}
	}
}
