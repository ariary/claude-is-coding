# cmem

Browse and promote Claude Code memory across all local projects.

```sh
go install github.com/ariary/cmem@latest
```

## Commands

| Command | Description |
|---------|-------------|
| `cmem list` | List all memory entries across projects |
| `cmem list --project <name>` | Filter by project (substring match) |
| `cmem list --type <type>` | Filter by type (`user\|feedback\|project\|reference`) |
| `cmem show <project> <file>` | Print full content of one entry |
| `cmem stats` | Counts by type + cross-project duplicates (promote candidates) |
| `cmem promote <project> <file>` | Append entry to `~/.claude/CLAUDE.md` (idempotent) |

Memory is read from `~/.claude/projects/*/memory/`. Use `--base-dir` to override.
