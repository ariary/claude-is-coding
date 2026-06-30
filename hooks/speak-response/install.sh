#!/usr/bin/env bash
# install.sh — install the speak-response Stop hook
set -euo pipefail

HOOK_DIR="$HOME/.claude/hooks/speak-response"
SETTINGS="$HOME/.claude/settings.json"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing speak-response hook..."

# 1. Copy script
mkdir -p "$HOOK_DIR"
cp "$SCRIPT_DIR/speak-response.sh" "$HOOK_DIR/speak-response.sh"
chmod +x "$HOOK_DIR/speak-response.sh"
echo "  copied $HOOK_DIR/speak-response.sh"

# 2. Add Stop hook to settings.json (idempotent)
python3 - "$SETTINGS" <<'PYEOF'
import sys, json

path = sys.argv[1]

with open(path) as f:
    cfg = json.load(f)

entry = {
    "matcher": "",
    "hooks": [{
        "type": "command",
        "command": "bash ~/.claude/hooks/speak-response/speak-response.sh",
        "async": True
    }]
}

hooks = cfg.setdefault("hooks", {})
stop  = hooks.setdefault("Stop", [])

for h in stop:
    for item in h.get("hooks", []):
        if "speak-response" in item.get("command", ""):
            print("  Stop hook already registered — skipping")
            sys.exit(0)

stop.insert(0, entry)

with open(path, "w") as f:
    json.dump(cfg, f, indent=2)
    f.write("\n")

print("  registered Stop hook in settings.json")
PYEOF

echo ""
echo "Done. Reload Claude Code (or open /hooks) to activate."
echo ""
echo "Suggested aliases (add to ~/.zshrc):"
echo "  alias speak-install='bash $SCRIPT_DIR/install.sh'"
echo "  alias speak-uninstall='bash $SCRIPT_DIR/uninstall.sh'"
