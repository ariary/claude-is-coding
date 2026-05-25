package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var baseDir string

var rootCmd = &cobra.Command{
	Use:   "cmem",
	Short: "Browse and promote Claude Code memory across projects",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&baseDir, "base-dir", "", "base directory for Claude projects (default: ~/.claude/projects)")
}
