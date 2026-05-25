package picker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/ariary/cs/internal/sessions"
)

const (
	// Narrow Unicode circles: both 1 display column, same byte width — safe for tabwriter.
	dotActive   = "●"
	dotInactive = "○"
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

	lines := formatLines(sess)

	selected, err := gumFilter(lines)
	if err != nil {
		return err
	}
	if selected == "" {
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

// formatLines formats all sessions into aligned display lines using tabwriter.
func formatLines(sess []sessions.Session) []string {
	// Build tab-separated lines first.
	tabbed := make([]string, len(sess))
	for i, s := range sess {
		tabbed[i] = formatLineTabbed(s)
	}

	// Run through tabwriter to produce aligned output.
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for _, l := range tabbed {
		fmt.Fprintln(w, l)
	}
	w.Flush()

	out := strings.TrimRight(buf.String(), "\n")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// formatLineTabbed returns a tab-separated line for one session.
func formatLineTabbed(s sessions.Session) string {
	status := dotInactive
	if s.Active {
		status = dotActive
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
		project = "?"
	}
	age := sessions.FormatAge(s.LastUsed)
	dur := sessions.FormatDuration(s.Duration)
	if dur == "" {
		return fmt.Sprintf("%s\t%s\t%s\t%s", status, name, project, age)
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s", status, name, project, age, dur)
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
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil // user cancelled (Escape/Ctrl-C)
		}
		return "", fmt.Errorf("gum filter: %w", err)
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
