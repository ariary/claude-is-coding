package sessions

import (
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
