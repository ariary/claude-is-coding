package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ariary/cs/internal/picker"
	"github.com/ariary/cs/internal/sessions"
)

func main() {
	limit := flag.Int("limit", 50, "number of most-recent sessions to show")
	all := flag.Bool("all", false, "show all sessions (bypass limit)")
	flag.Parse()

	if *all {
		*limit = 0
	}

	sess, err := sessions.Load(*limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cs: %v\n", err)
		os.Exit(1)
	}

	if len(sess) == 0 {
		fmt.Fprintln(os.Stderr, "cs: no sessions found")
		os.Exit(0)
	}

	if err := picker.Run(sess); err != nil {
		fmt.Fprintf(os.Stderr, "cs: %v\n", err)
		os.Exit(1)
	}
}
