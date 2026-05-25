package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ariary/cmem/memory"
	"github.com/spf13/cobra"
)

var listProject string
var listType string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all memory entries across projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := memory.Scan(baseDir)
		if err != nil {
			return err
		}

		for _, e := range entries {
			if listProject != "" && !strings.Contains(strings.ToLower(e.Project), strings.ToLower(listProject)) {
				continue
			}
			if listType != "" && e.Type != listType {
				continue
			}
			fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%s\n", e.Project, e.Type, e.Name, e.Description)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listProject, "project", "", "filter by project name (substring match)")
	listCmd.Flags().StringVar(&listType, "type", "", "filter by type (user|feedback|project|reference)")
	rootCmd.AddCommand(listCmd)
}
