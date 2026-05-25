# cs

Fuzzy-searchable session switcher for Claude Code. Browse all sessions (active and historical) and jump into one — working directory switch included.

Requires [`gum`](https://github.com/charmbracelet/gum).

```sh
cd cs && go build -o ~/.local/bin/cs ./cmd/cs
```

## Usage

```
cs           # pick from last 50 sessions
cs --all     # show all sessions
cs --limit N # show last N sessions
```

Each line shows: status (`🟢` active / `⚫` inactive), session name, project, last used, duration. Fuzzy-search across all columns. Selecting a session `cd`s to its directory and resumes it in Claude Code.
