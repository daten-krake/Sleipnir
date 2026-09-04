#!/usr/bin/env bash
# Refresh the token-usage table of a session tracker file from the live pi session log.
# Usage: ./sessions/update-usage.sh <session-file.md> [session-jsonl]
set -euo pipefail

FILE="${1:?usage: update-usage.sh <session-file.md> [session-jsonl]}"
LOG="${2:-${PI_SESSION_FILE:-}}"
[ -n "$LOG" ] || { echo "no session log: pass path or set PI_SESSION_FILE" >&2; exit 1; }
[ -f "$LOG" ] || { echo "session log not found: $LOG" >&2; exit 1; }

LINE=$(python3 - "$LOG" <<'EOF'
import json, sys
t = {'input': 0, 'output': 0, 'cacheRead': 0, 'cacheWrite': 0, 'reasoning': 0}
for l in open(sys.argv[1]):
    try:
        d = json.loads(l)
    except Exception:
        continue
    u = (d.get('message') or {}).get('usage') or d.get('usage')
    if not u:
        continue
    for k in t:
        t[k] += u.get(k, 0) or 0
total_in = t['input'] + t['cacheRead'] + t['cacheWrite']
print(f"| {total_in} | {t['input']} | {t['cacheRead']} | {t['cacheWrite']} | {t['output']} | {t['reasoning']} |")
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
