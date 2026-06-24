#!/bin/bash
# On session resume, pre-touch the rename flag if the session already has a title.
# This prevents auto-rename.sh from firing on the first prompt after a device reboot
# (which clears /tmp), since the session was already named in a previous boot.

INPUT=$(cat /dev/stdin)
SOURCE=$(echo "$INPUT" | jq -r '.source // empty')
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
SESSION_TITLE=$(echo "$INPUT" | jq -r '.session_title // empty')

if [ "$SOURCE" = "resume" ] && [ -n "$SESSION_TITLE" ] && [ -n "$SESSION_ID" ]; then
  touch "/tmp/claude-renamed-${SESSION_ID}"
fi
