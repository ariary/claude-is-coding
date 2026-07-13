#!/usr/bin/env bash
# uninstall.sh — remove the speak-response Stop hook
set -euo pipefail

HOOK_DIR="$HOME/.claude/hooks/speak-response"
SETTINGS="$HOME/.claude/settings.json"

echo "Uninstalling speak-response hook..."

# 1. Remove Stop hook from settings.json
python3 - "$SETTINGS" <<'PYEOF'
import sys, json

path = sys.argv[1]

with open(path) as f:
    cfg = json.load(f)

stop = cfg.get("hooks", {}).get("Stop", [])
before = len(stop)
stop[:] = [
    h for h in stop
    if not any("speak-response" in item.get("command", "") for item in h.get("hooks", []))
]

if len(stop) < before:
    with open(path, "w") as f:
        json.dump(cfg, f, indent=2)
        f.write("\n")
    print("  removed Stop hook from settings.json")
else:
    print("  Stop hook not found in settings.json — skipping")
PYEOF

# 2. Remove installed script
if [ -d "$HOOK_DIR" ]; then
    rm -rf "$HOOK_DIR"
    echo "  removed $HOOK_DIR"
fi

echo ""
echo "Done. Reload Claude Code (or open /hooks) to deactivate."
