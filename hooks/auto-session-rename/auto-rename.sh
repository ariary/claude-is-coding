#!/bin/bash
# Auto-rename the Claude Code session title.
#
# Session title: concise 3-6 word description of what the user is asking Claude.
#
# Only runs once per session (flag file keyed to session_id).

INPUT=$(cat /dev/stdin)
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')

# No session ID → can't track state reliably, skip
if [ -z "$SESSION_ID" ]; then
  echo '{}'
  exit 0
fi

# Already ran for this session → exit immediately (~1ms)
FLAG="/tmp/claude-renamed-${SESSION_ID}"
if [ -f "$FLAG" ]; then
  echo '{}'
  exit 0
fi

touch "$FLAG"

PROMPT=$(echo "$INPUT" | jq -r '.prompt // empty')

# No prompt content → skip (nothing to base the title on)
if [ -z "$PROMPT" ]; then
  echo '{}'
  exit 0
fi

SYSTEM_PROMPT="You generate session titles for coding sessions. The title must be the TECHNICAL SUBJECT only — what is being built, fixed, debugged, or discussed. The output title MUST NOT contain these words: help, me, you, can, please, want, need, how, explain, possible, i, is, it, yes, fact, do, get, make, just, some, more, clearly. Reply with JSON only, no markdown, no explanation."

USER_PROMPT="User's message: ${PROMPT}

Good examples (topic-first, no filler):
- 'is it possible in claude hook to detect if it is the first prompt' → 'detect-first-prompt-resumed-session'
- 'how do i add pagination to the users endpoint' → 'add-users-endpoint-pagination'
- 'yes in fact the issue occurs when I restart my device' → 'rename-flag-lost-on-reboot'
- 'can you explain option 2 more clearly' → 'clarify-sessionstart-resume-option'

Bad examples (filler leaked into title — NEVER do this):
- 'help me fix the login bug' → ❌ 'help-me-fix-login' → ✅ 'fix-login-bug'
- 'can you help me understand pagination' → ❌ 'help-understand-pagination' → ✅ 'pagination-explained'

Return JSON:
{
  \"session_title\": \"3-6 words, lowercase, hyphen-separated, the technical subject only\"
}"

RESPONSE=$(curl -s https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d "$(jq -n \
    --arg sys "$SYSTEM_PROMPT" \
    --arg usr "$USER_PROMPT" \
    '{
      "model": "claude-haiku-4-5-20251001",
      "max_tokens": 60,
      "system": $sys,
      "messages": [{"role": "user", "content": $usr}]
    }')" \
  | jq -r '.content[0].text // empty')

sanitize() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' \
    | sed 's/[^a-z0-9-]//g; s/--*/-/g; s/^-//; s/-$//'
}

SESSION_TITLE=$(echo "$RESPONSE" | jq -r '.session_title // empty' 2>/dev/null)

# Fallback if JSON parsing fails
if [ -z "$SESSION_TITLE" ] || [ "$SESSION_TITLE" = "null" ]; then
  SESSION_TITLE=$(echo "$PROMPT" | head -c 40 | sed 's/[^a-zA-Z0-9 -]//g' | xargs)
fi

SESSION_TITLE=$(sanitize "$SESSION_TITLE")

# Persist session title for other hooks (e.g. stop-notify)
echo "$SESSION_TITLE" > "/tmp/claude-title-${SESSION_ID}"

jq -n --arg t "$SESSION_TITLE" '{
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "sessionTitle": $t
  }
}'
