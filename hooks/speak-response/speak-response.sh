#!/usr/bin/env bash
# speak-response.sh — Stop hook: reads the last Claude assistant response aloud via macOS `say`.
# Wire up as an async Stop hook in settings.json to auto-speak every response.
# To disable: remove the hook entry from settings.json.

input=$(cat)
SESSION_ID=$(echo "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('session_id',''))" 2>/dev/null || true)

[ -z "$SESSION_ID" ] && exit 0

TRANSCRIPT=$(find ~/.claude/projects -name "${SESSION_ID}.jsonl" 2>/dev/null | head -1 || true)

[ -z "$TRANSCRIPT" ] && exit 0

python3 - "$TRANSCRIPT" <<'PYEOF'
import sys, json, re, subprocess, os

transcript_path = sys.argv[1]

with open(transcript_path) as f:
    lines = f.readlines()

text = None
for line in reversed(lines):
    line = line.strip()
    if not line:
        continue
    try:
        obj = json.loads(line)
        if obj.get('type') == 'assistant':
            content = obj.get('message', {}).get('content', [])
            texts = [b['text'] for b in content if isinstance(b, dict) and b.get('type') == 'text']
            if texts:
                text = ' '.join(texts)
                break
    except:
        pass

if not text:
    sys.exit(0)

# Strip markdown so `say` reads clean prose
text = re.sub(r'```[\s\S]*?```', 'code block.', text)
text = re.sub(r'`[^`]+`', 'code', text)
text = re.sub(r'^#{1,6}\s+', '', text, flags=re.MULTILINE)
text = re.sub(r'\*\*([^*]+)\*\*', r'\1', text)
text = re.sub(r'\*([^*]+)\*', r'\1', text)
text = re.sub(r'_([^_]+)_', r'\1', text)
text = re.sub(r'\[([^\]]+)\]\([^\)]+\)', r'\1', text)
text = re.sub(r'^\s*[-*+]\s+', '', text, flags=re.MULTILINE)
text = re.sub(r'^\s*\d+\.\s+', '', text, flags=re.MULTILINE)
text = re.sub(r'\n+', ' ', text)
text = text.strip()

if len(text) > 3000:
    text = text[:3000] + '... response truncated.'

os.system('killall say 2>/dev/null || true')
subprocess.Popen(['say', text])
PYEOF
