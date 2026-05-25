package picker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/ariary/cs/internal/sessions"
)

const (
	emojiActive   = "🟢"
	emojiInactive = "⚫"
)

// Run presents a gum filter picker and execs into the selected session.
func Run(sess []sessions.Session) error {
	if _, err := checkDep("gum"); err != nil {
		return err
	}
	claudePath, err := checkDep("claude")
	if err != nil {
		return err
	}

	lines := make([]string, len(sess))
	for i, s := range sess {
		lines[i] = formatLine(s)
	}

	selected, err := gumFilter(lines)
	if err != nil || selected == "" {
		// User hit Escape / Ctrl-C — exit cleanly
		os.Exit(0)
	}

	// Find selected session by matching the display line
	var chosen *sessions.Session
	for i, l := range lines {
		if l == selected {
			chosen = &sess[i]
			break
		}
	}
	if chosen == nil {
		return fmt.Errorf("could not match selection to a session")
	}

	// Chdir to session's working directory
	if chosen.Cwd != "" {
		if err := os.Chdir(chosen.Cwd); err != nil {
			fmt.Fprintf(os.Stderr, "cs: warning: could not chdir to %s: %v\n", chosen.Cwd, err)
		}
	}

	// Exec claude --resume <id>, replacing current process
	return syscall.Exec(claudePath, []string{"claude", "--resume", chosen.ID}, os.Environ())
}

// formatLine returns a display line for one session.
func formatLine(s sessions.Session) string {
	emoji := emojiInactive
	if s.Active {
		emoji = emojiActive
	}
	name := s.Name
	if name == "" {
		name = s.ID
		if len(name) > 8 {
			name = name[:8]
		}
	}
	project := s.Project
	if project == "" {
		project = "unknown"
	}
	age := sessions.FormatAge(s.LastUsed)
	dur := sessions.FormatDuration(s.Duration)

	parts := []string{emoji, padRight(name, 35), padRight(project, 30), padRight(age, 10)}
	if dur != "" {
		parts = append(parts, dur)
	}
	return strings.Join(parts, "  ")
}

// padRight right-pads s to width with spaces.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// gumFilter runs gum filter and returns the selected line.
func gumFilter(lines []string) (string, error) {
	input := strings.Join(lines, "\n")
	cmd := exec.Command("gum", "filter",
		"--placeholder", "search sessions...",
		"--height", "20",
		"--no-limit",
	)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimRight(out.String(), "\n"), nil
}

// checkDep verifies a binary is in PATH and returns its full path.
func checkDep(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s is required but not found in PATH (brew install %s)", name, name)
	}
	return path, nil
}
