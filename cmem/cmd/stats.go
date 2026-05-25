package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ariary/cmem/memory"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show entry counts and cross-project duplicates",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := memory.Scan(baseDir)
		if err != nil {
			return err
		}

		// counts per type
		typeCounts := map[string]int{}
		for _, e := range entries {
			typeCounts[e.Type]++
		}

		types := []string{"user", "feedback", "project", "reference"}
		fmt.Fprintln(os.Stdout, "=== counts by type ===")
		for _, t := range types {
			if n := typeCounts[t]; n > 0 {
				fmt.Fprintf(os.Stdout, "%s\t%d\n", t, n)
			}
		}
		fmt.Fprintf(os.Stdout, "total\t%d\n", len(entries))

		// de-dupe candidates: entries whose Name or Description appears in 2+ distinct projects
		nameToProjectSet := map[string]map[string]struct{}{}
		addKey := func(key, project string) {
			if nameToProjectSet[key] == nil {
				nameToProjectSet[key] = map[string]struct{}{}
			}
			nameToProjectSet[key][project] = struct{}{}
		}
		for _, e := range entries {
			addKey(strings.ToLower(e.Name), e.Project)
			if e.Description != "" {
				addKey(strings.ToLower(e.Description), e.Project)
			}
		}

		type dup struct {
			name     string
			projects []string
		}
		var dups []dup
		for name, projectSet := range nameToProjectSet {
			if len(projectSet) >= 2 {
				var projects []string
				for p := range projectSet {
					projects = append(projects, p)
				}
				sort.Strings(projects)
				dups = append(dups, dup{name, projects})
			}
		}
		sort.Slice(dups, func(i, j int) bool {
			if len(dups[i].projects) != len(dups[j].projects) {
				return len(dups[i].projects) > len(dups[j].projects)
			}
			return dups[i].name < dups[j].name
		})

		if len(dups) > 0 {
			fmt.Fprintln(os.Stdout, "\n=== cross-project duplicates (promote candidates) ===")
			for _, d := range dups {
				fmt.Fprintf(os.Stdout, "%s\t(%d projects: %s)\n", d.name, len(d.projects), strings.Join(d.projects, ", "))
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
