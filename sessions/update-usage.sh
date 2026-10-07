#!/usr/bin/env bash
# Refresh the token-usage table of a session tracker file from the OpenCode session API.
# Usage: ./sessions/update-usage.sh <session-file.md> <opencode-session-id>
set -euo pipefail

FILE="${1:?usage: update-usage.sh <session-file.md> <opencode-session-id>}"
SESSION="${2:?usage: update-usage.sh <session-file.md> <opencode-session-id>}"

JSON=$(opencode api get "/api/session/${SESSION}")

LINE=$(python3 - "$JSON" <<'EOF'
import json, sys
d = json.loads(sys.argv[1])
s = d.get('data') or d
t = s.get('tokens') or {}
c = t.get('cache') or {}
inp = t.get('input', 0) or 0
cr = c.get('read', 0) or 0
cw = c.get('write', 0) or 0
out = t.get('output', 0) or 0
rea = t.get('reasoning', 0) or 0
print(f"| {inp + cr + cw} | {inp} | {cr} | {cw} | {out} | {rea} |")
EOF
)

if grep -q '^| total input |' "$FILE"; then
  # keep header + separator, replace the data row after them
  awk -v rep="$LINE" '
    /^\| total input \|/ { print; getline; print; print rep; getline; next }
    { print }
  ' "$FILE" > "$FILE.tmp" && mv "$FILE.tmp" "$FILE"
else
  {
    echo ""
    echo "## Token usage"
    echo ""
    echo "| total input | uncached input | cache read | cache write | output | reasoning |"
    echo "|---|---|---|---|---|---|"
    echo "$LINE"
  } >> "$FILE"
fi
echo "updated $FILE:"
echo "$LINE"
