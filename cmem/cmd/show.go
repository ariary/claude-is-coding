package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ariary/cmem/memory"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <project> <entry-file>",
	Short: "Print full content of a memory entry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFilter := args[0]
		fileFilter := args[1]

		entries, err := memory.Scan(baseDir)
		if err != nil {
			return err
		}

		for _, e := range entries {
			if !strings.Contains(strings.ToLower(e.Project), strings.ToLower(projectFilter)) {
				continue
			}
			if e.File != fileFilter {
				continue
			}
			fmt.Fprintf(os.Stdout, "# %s\n", e.Name)
			fmt.Fprintf(os.Stdout, "project:     %s\n", e.Project)
			fmt.Fprintf(os.Stdout, "type:        %s\n", e.Type)
			fmt.Fprintf(os.Stdout, "description: %s\n", e.Description)
			fmt.Fprintf(os.Stdout, "file:        %s\n\n", e.File)
			fmt.Fprintln(os.Stdout, e.Body)
			return nil
		}

		fmt.Fprintf(os.Stderr, "entry not found: project=%q file=%q\n", projectFilter, fileFilter)
		os.Exit(1)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
