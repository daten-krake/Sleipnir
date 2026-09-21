#!/usr/bin/env bash
# Negative/positive fixtures for the two CI gates added by WP-01.3, kept in the
# repo because the first copy lived in /tmp/wp01/ci-fix/ and was destroyed by the
# 2026-09-21 OOM reboot — along with the session that was going to run it.
#
#   AGENTS.md, "Never prove a negative test by removing the bound and running
#   it": fixtures mutate COPIES outside the repository, never the working tree.
#   This file is that rule applied to CI assertions.
#
# What it proves, for .github/workflows/ci.yml job `contracts`:
#
#   A. "recompute every published vector" (finding C-f). Two assertions, both
#      required. The exit code alone is not enough: a verifier whose regexes
#      stopped matching computes zero checks, so its failure list stays empty
#      and it exits 0 while printing 'PASS 0 / FAIL 0' — which is how I-02 got
#      introduced. The summary grep 'PASS [1-9][0-9]* / FAIL 0' is the only
#      assertion that catches that shape, and case A4 exists to keep it honest.
#      A1-A3 are the reason the check was strengthened at all: the published
#      digest used to be compared against a constant hardcoded in the script, so
#      editing the value printed IN THE FROZEN CONTRACT went undetected.
#
#   B. "no credential files or editor droppings are tracked" (finding C-e).
#      WORKFLOW §7 / ADR-0019 §5. Names only, never contents — reading contents
#      would copy secret material into the CI log, which is what §5 forbids.
#
# Both step bodies below are transcribed verbatim from ci.yml. If you edit the
# workflow, edit them here too, or this file starts proving the wrong thing.
#
# Usage: bash docs/reviews/2026-09-21-ci-gate-fixtures.sh
# Exits non-zero if any fixture disagrees with its expectation.
set -u

REPO=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
BASE=$(mktemp -d "${TMPDIR:-/tmp}/slp-ci-fixtures.XXXXXX")
trap 'rm -rf "$BASE"' EXIT
FAILURES=0

report() { # report <expected PASS|FAIL> <label> <rc>
  local expect="$1" label="$2" rc="$3"
  if { [ "$expect" = PASS ] && [ "$rc" -eq 0 ]; } || \
     { [ "$expect" = FAIL ] && [ "$rc" -ne 0 ]; }; then
    printf 'ok    %-52s (rc=%s)\n' "$label" "$rc"
  else
    printf 'WRONG %-52s (rc=%s, expected %s)\n' "$label" "$rc" "$expect"
    FAILURES=$((FAILURES + 1))
  fi
}

# ---------------------------------------------------------------- A: vectors

mk_vectors() { # mk_vectors <name> -> fixture dir holding contracts/ + verifier
  local d="$BASE/$1"
  mkdir -p "$d/docs/reviews"
  cp -r "$REPO/contracts" "$d/"
  cp "$REPO"/docs/reviews/*-verify-vectors.py "$d/docs/reviews/"
  printf '%s' "$d"
}

# Verbatim from ci.yml: both assertions, in order, output captured first so the
# script's 'FAIL:' detail lines reach the log either way.
vectors_step() {
  local d="$1" s rc=0 summary
  cd "$d" || return 9
  shopt -s nullglob
  s=(docs/reviews/*-verify-vectors.py)
  [ "${#s[@]}" -eq 0 ] && { echo "no verifier matched"; return 1; }
  summary=$(python3 "${s[0]}") || rc=$?
  printf '%s\n' "$summary" | tail -2 | sed 's/^/      /'
  if [ "$rc" -ne 0 ]; then
    echo "      ASSERT-1 exit code: REJECTED (exited $rc)"
    return 1
  fi
  echo "      ASSERT-1 exit code: accepted (0)"
  if ! printf '%s\n' "$summary" | grep -qx 'PASS [1-9][0-9]* / FAIL 0'; then
    echo "      ASSERT-2 summary grep: REJECTED (no 'PASS n>=1 / FAIL 0')"
    return 1
  fi
  echo "      ASSERT-2 summary grep: accepted"
  return 0
}

A1=contracts/A1-events.md
DIGEST=19d976a83cc9d36ac160313a20b80c0745fff805526b7f43f88d05e33c7be5e5
PRE='    {"kind":"command_executed","payload":{"command":"nmap -sV -p 445 10.20.0.14"'

echo "== A: contract vector verifier (ci.yml job 'contracts')"

d=$(mk_vectors a0-clean)
echo "  A0 untouched contract — the only case that must pass"
vectors_step "$d"; report PASS "A0 untouched contract" $?

d=$(mk_vectors a1-digest)
python3 - "$d/$A1" "$DIGEST" <<'PY'
import sys
p, dg = sys.argv[1], sys.argv[2]
s = open(p, encoding='utf-8').read()
assert dg in s, "published digest not found — the fixture has rotted"
bad = 'f' * 64 if dg[0] != 'f' else '0' * 64
open(p, 'w', encoding='utf-8').write(s.replace(dg, bad, 1))
print("      tampered the digest PUBLISHED IN A1 ->", bad[:12] + "...")
PY
echo "  A1 published digest edited"
vectors_step "$d"; report FAIL "A1 published digest edited" $?

d=$(mk_vectors a2-len)
python3 - "$d/$A1" <<'PY'
import sys
p = sys.argv[1]
s = open(p, encoding='utf-8').read()
assert '**len 293 ' in s, "published length not found — the fixture has rotted"
open(p, 'w', encoding='utf-8').write(s.replace('**len 293 ', '**len 294 ', 1))
print("      tampered the published length 293 -> 294")
PY
echo "  A2 published length edited"
vectors_step "$d"; report FAIL "A2 published length edited" $?

d=$(mk_vectors a3-preimage)
python3 - "$d/$A1" "$PRE" <<'PY'
import sys
p, pre = sys.argv[1], sys.argv[2]
s = open(p, encoding='utf-8').read()
assert pre in s, "preimage block not found — the fixture has rotted"
# Same byte count, so the length still matches and only the digest can catch it.
new = pre.replace('nmap -sV', 'nmap -sS')
assert len(new) == len(pre)
open(p, 'w', encoding='utf-8').write(s.replace(pre, new, 1))
print("      tampered the preimage: 'nmap -sV' -> 'nmap -sS'")
PY
echo "  A3 preimage edited"
vectors_step "$d"; report FAIL "A3 preimage edited" $?

d=$(mk_vectors a4-vacuous)
cat > "$d/docs/reviews/2026-09-11-verify-vectors.py" <<'PY'
import sys
ok, bad = [], []           # every regex silently stopped matching anything
print(f"PASS {len(ok)} / FAIL {len(bad)}")
sys.exit(1 if bad else 0)  # exits 0 having verified nothing
PY
echo "  A4 vacuous verifier — ASSERT-1 accepts it, only ASSERT-2 can catch it"
vectors_step "$d"; report FAIL "A4 vacuous verifier (PASS 0 / FAIL 0, exit 0)" $?

d=$(mk_vectors a5-loud)
cat > "$d/docs/reviews/2026-09-11-verify-vectors.py" <<'PY'
import sys
print("PASS 51 / FAIL 1"); print("  FAIL: A1-7.6 PayloadHash vector")
sys.exit(1)
PY
echo "  A5 loud failure — ASSERT-1 must reject on exit code alone"
vectors_step "$d"; report FAIL "A5 loud failure (PASS 51 / FAIL 1, exit 1)" $?

# --------------------------------------------------------------- B: hygiene

hygiene_step() { # verbatim from ci.yml
  set -euo pipefail
  files=$(git ls-files) # an unreadable tree must fail the step, not pass it
  pat='(^|/)id_rsa|\.pem$|\.key$|\.token$|~$|\.orig$|\.rej$'
  if bad=$(printf '%s\n' "$files" | grep -E "$pat"); then
    printf '%s\n' "$bad"
    echo "tracked file name(s) above look like credentials or editor/merge droppings"
    echo "(WORKFLOW §7, ADR-0019 §5). Remove them from the tree and rotate the secret."
    exit 1
  fi
  echo "ok: tracked file names checked against credential/droppings patterns, contents never read"
}

mk_hygiene() { # mk_hygiene <name> [tracked offender files...]
  local d="$BASE/$1"; shift
  mkdir -p "$d"
  git -C "$d" init -q
  git -C "$d" config user.email fixtures@invalid
  git -C "$d" config user.name  ci-fixtures
  echo "legitimate content" > "$d/README.md"
  git -C "$d" add README.md
  local f
  for f in "$@"; do
    mkdir -p "$d/$(dirname "$f")"
    echo "x" > "$d/$f"
    git -C "$d" add -f "$f"
  done
  git -C "$d" commit -qm fixture
  printf '%s' "$d"
}

echo
echo "== B: credential / editor-droppings hygiene (ci.yml job 'contracts')"

b() { # b <expect> <label> <dir>
  local expect="$1" label="$2" d="$3" out rc=0
  out=$(cd "$d" && (hygiene_step) 2>&1) || rc=$?
  printf '%s\n' "$out" | sed 's/^/      /'
  report "$expect" "$label" "$rc"
}

b PASS "B0 clean tree"                "$(mk_hygiene b0)"
b FAIL "B1 tracked deploy.pem"        "$(mk_hygiene b1 deploy.pem)"
b FAIL "B2 tracked nested id_rsa"     "$(mk_hygiene b2 config/id_rsa)"
b FAIL "B3 tracked signing.key"       "$(mk_hygiene b3 signing.key)"
b FAIL "B4 tracked gh.token"          "$(mk_hygiene b4 gh.token)"
b FAIL "B5 tracked editor backup ~"   "$(mk_hygiene b5 'docs/notes.md~')"
b FAIL "B6 tracked merge drop .orig"  "$(mk_hygiene b6 ci.yml.orig)"
b FAIL "B7 tracked reject drop .rej"  "$(mk_hygiene b7 patch.rej)"
# Names that merely contain the letters must not trip the gate, or it gets
# disabled by whoever hits the first false positive.
near_miss=$(mk_hygiene b8 docs/pemfile.md internal/api/apikey.go \
  assets/keyboard.svg docs/tokenising.md)
b PASS "B8 near-misses stay clean" "$near_miss"

echo "  B9 the real repository"
( cd "$REPO" && hygiene_step ) | sed 's/^/      /'
report PASS "B9 real repository" "${PIPESTATUS[0]}"

echo
if [ "$FAILURES" -eq 0 ]; then
  echo "ALL FIXTURES AGREE with the assertions in .github/workflows/ci.yml"
else
  echo "$FAILURES FIXTURE(S) DISAGREE — ci.yml and this file have drifted apart"
fi
exit "$((FAILURES > 0))"
