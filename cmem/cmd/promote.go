package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ariary/cmem/memory"
	"github.com/spf13/cobra"
)

var promoteCmd = &cobra.Command{
	Use:   "promote <project> <entry-file>",
	Short: "Promote a memory entry to ~/.claude/CLAUDE.md",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFilter := args[0]
		fileFilter := args[1]

		entries, err := memory.Scan(baseDir)
		if err != nil {
			return err
		}

		var target *memory.Entry
		for i := range entries {
			if strings.Contains(strings.ToLower(entries[i].Project), strings.ToLower(projectFilter)) &&
				entries[i].File == fileFilter {
				target = &entries[i]
				break
			}
		}
		if target == nil {
			return fmt.Errorf("entry not found: project=%q file=%q", projectFilter, fileFilter)
		}

		claudePath, err := globalCLAUDEMDPath()
		if err != nil {
			return err
		}

		// read existing content
		existing, _ := os.ReadFile(claudePath)
		header := fmt.Sprintf("## %s", target.Name)
		existingStr := string(existing)
		if strings.Contains(existingStr, "\n"+header+"\n") || strings.HasPrefix(existingStr, header+"\n") {
			fmt.Fprintf(os.Stdout, "already promoted: %q\n", target.Name)
			return nil
		}

		// append section
		f, err := os.OpenFile(claudePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		section := fmt.Sprintf("\n%s\n%s\n", header, target.Body)
		if _, err := f.WriteString(section); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "promoted %q to %s\n", target.Name, claudePath)
		return nil
	},
}

func globalCLAUDEMDPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "CLAUDE.md"), nil
}

func init() {
	rootCmd.AddCommand(promoteCmd)
}
