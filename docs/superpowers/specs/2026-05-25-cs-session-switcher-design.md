# `cs` — Claude Session Switcher

**Date:** 2026-05-25
**Status:** Approved

## Summary

A Go CLI tool (`cs`) that presents a fuzzy-searchable terminal overview of all Claude Code sessions (active and historical) and directly launches the selected one — handling the working directory switch automatically.

## Problem

`claude -r` exists but is minimal. When working across many projects, you want to see sessions at a glance (name, project, active status, recency) and jump into one without manually `cd`-ing first.

## Architecture

Two internal packages wired by a thin `main.go`:

- **`internal/sessions`** — data layer: reads `~/.claude/projects/` (all historical `.jsonl` files) and `~/.claude/sessions/` (active PIDs), merges and deduplicates into a sorted `[]Session` slice.
- **`internal/picker`** — UI layer: formats sessions into display lines, calls `gum filter`, parses the selection, then `chdir` + `exec`s `claude --resume <id>`.

## Data Model

```go
type Session struct {
    ID       string
    Name     string        // from custom-title entry in .jsonl
    Cwd      string        // full working directory path
    Project  string        // last 2 segments of Cwd (display only)
    LastUsed time.Time     // mtime of .jsonl file
    Duration time.Duration // active: time.Since(startedAt); historical: .jsonl mtime − ctime (best-effort)
    Active   bool          // corresponding PID file exists in ~/.claude/sessions/
}
```

## Data Sources

| Source | Purpose |
|--------|---------|
| `~/.claude/projects/<encoded-path>/<session-id>.jsonl` | One file per session. `custom-title` in first lines gives the name. File mtime = last activity. |
| `~/.claude/sessions/<pid>.json` | Active sessions. Contains `sessionId` to cross-reference. |

The `cwd` for active sessions is read directly from `~/.claude/sessions/<pid>.json`. For historical-only sessions, it is decoded from the project directory name (Claude encodes `/` as `-` with a leading `-`). Since paths can contain `-`, decoding is best-effort; edge cases fall back to the raw encoded string shown as-is.

## Display Format

Each line fed to `gum filter`:

```
🟢  fix-mobile-login-bug        ariary/claude-is-coding   2m ago    45min
⚫  add-auth-middleware          ariary/api                3d ago    20min
⚫  build-credential-proxy       ariary/x-raygent          1w ago    1h20min
```

- `🟢` = active (running), `⚫` = inactive
- Columns: status emoji, name, project (last 2 path segments), last used, duration
- `gum filter` fuzzy-searches across the full display line (name + project + timing)
- Session ID is stored in a parallel slice indexed by line number — not shown in display

## Launch Behavior

After selection, the binary:
1. `syscall.Chdir(session.Cwd)` — switches to the session's working directory
2. `syscall.Exec("claude", ["claude", "--resume", session.ID], env)` — replaces itself with `claude`

No subshell, no `eval`, no wrapper needed. The terminal ends up in `claude` directly.

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--limit N` | 50 | Number of most-recent sessions to show |
| `--all` | false | Bypass limit, show all sessions |

## File Structure

```
tools/cs/
├── go.mod
├── cmd/cs/
│   └── main.go          # parse flags, call sessions.Load + picker.Run
├── internal/sessions/
│   └── sessions.go      # Load(limit int) ([]Session, error)
└── internal/picker/
    └── picker.go        # Run(sessions []Session) error
```

## Error Handling

- `gum` not in PATH → print actionable message: `gum is required: brew install gum`
- `claude` not in PATH → error before exec
- No sessions found → print friendly message, exit 0
- Malformed `.jsonl` → skip silently, continue

## Installation

```bash
cd tools/cs && go build -o ~/bin/cs ./cmd/cs
```

Drop `~/bin` in `$PATH`. Run `cs` from anywhere.
