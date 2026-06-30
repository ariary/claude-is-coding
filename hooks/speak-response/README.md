# Speak Response

Automatically read every Claude response aloud via macOS `say`.

A single async `Stop` hook fires after every turn, extracts the last assistant text, strips markdown, and pipes it to `say`. Toggle by running `install.sh` / `uninstall.sh`.

## How it works

```
Claude finishes a turn
  → Stop hook fires (async) → speak-response.sh
    → locates transcript by session_id
    → extracts last assistant text block
    → strips markdown (code blocks, headers, bold, links…)
    → killall say  (interrupt any previous speech)
    → say "<cleaned text>"
```

## Prerequisites

- macOS (`say` is built-in)
- Python 3 (pre-installed on macOS)

## Install

```bash
bash hooks/speak-response/install.sh
```

Then reload Claude Code (or open `/hooks`) to activate.

## Uninstall

```bash
bash hooks/speak-response/uninstall.sh
```

## Aliases

Add to `~/.zshrc` for quick toggling from anywhere:

```bash
SPEAK_DIR="$HOME/project/ariary/claude-is-coding/hooks/speak-response"
alias speak-install="bash $SPEAK_DIR/install.sh"
alias speak-uninstall="bash $SPEAK_DIR/uninstall.sh"
```

## Customisation

Edit `speak-response.sh` after installing:

- **Voice**: change `say` to `say -v Samantha`
- **Speed**: add `-r 240` (words per minute)
- **Length cap**: change the `3000` char limit
