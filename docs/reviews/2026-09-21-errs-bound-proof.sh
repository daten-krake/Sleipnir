#!/usr/bin/env bash
# Bounded-walk negative proof for internal/errs (WP-01.1, ruling E-a), kept in
# the repo because the first attempt at it destroyed the machine.
#
# On 2026-09-21 an agent proved TestErrorChainWalkIsBounded could fail by
# short-circuiting the guard IN THE WORKING TREE and running it there. With the
# cap gone the cyclic-chain cases append to a strings.Builder without bound:
# errs.test reached 11.4 GB anon-rss on an 11 GiB WSL2 VM, the kernel OOM-killed
# it, and the VM went down — taking the session, two sibling agents and /tmp with
# it. The disabled guard was still in the tree afterwards and its /tmp backup did
# not survive the reboot.
#
# This script is the same proof done the way AGENTS.md now requires:
#   * it mutates a COPY under $TMPDIR and never writes to the repository;
#   * every run is capped three ways — ulimit -v (hard address-space ceiling),
#     go test -timeout (Go's own panic), and timeout --signal=KILL (the shell);
#   * it restores nothing because it changes nothing, so there is no window in
#     which an interrupted run leaves the repo unguarded.
# Under those caps the unbounded walk dies at ~139 MB in ~0.3s instead of eating
# the machine, which is itself the evidence: the cap in errs.go is load-bearing.
#
# Usage: bash docs/reviews/2026-09-21-errs-bound-proof.sh
# Exits non-zero if the positive control fails or the mutation does not fail.
set -u

REPO=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
BASE=$(mktemp -d "${TMPDIR:-/tmp}/slp-errs-proof.XXXXXX")
trap 'rm -rf "$BASE"' EXIT

# Address-space ceiling for the mutated run. 2 GiB is far above what the bounded
# walk needs (it peaks in single-digit MB) and far below what would hurt the
# host. Override only upward, and never remove it.
PROOF_ULIMIT_KB=${PROOF_ULIMIT_KB:-2097152}
PROOF_TIMEOUT=${PROOF_TIMEOUT:-90}
TEST=TestErrorChainWalkIsBounded
GUARD='		if depth == maxChainDepth {'
MUTATED='		if false && depth == maxChainDepth {'

copy() { # copy <dest>
  mkdir -p "$1/internal"
  cp "$REPO/go.mod" "$1/"
  cp -r "$REPO/internal/errs" "$1/internal/"
}

run() { # run <dir> -> capped go test, output on stdout, rc returned
  ( cd "$1" \
    && ulimit -v "$PROOF_ULIMIT_KB" \
    && exec timeout --signal=KILL "$PROOF_TIMEOUT" \
         go test -count=1 -timeout 30s -run "$TEST" ./internal/errs/ ) 2>&1
}

echo "== positive control: the cap as committed"
clean="$BASE/clean"; copy "$clean"
grep -qF "$GUARD" "$clean/internal/errs/errs.go" \
  || { echo "FATAL: the depth guard is not present in internal/errs/errs.go."; \
       echo "       Refusing to run anything until it is restored."; exit 2; }
out=$(run "$clean"); rc=$?
printf '%s\n' "$out" | sed 's/^/  /'
if [ $rc -ne 0 ]; then
  echo "WRONG: the bounded walk fails with the cap intact — fix errs.go first."
  exit 1
fi
echo "ok    cap intact -> $TEST passes"
echo

echo "== negative proof: the 2026-09-21 mutation, on a copy, capped"
mut="$BASE/mutated"; copy "$mut"
python3 - "$mut/internal/errs/errs.go" "$GUARD" "$MUTATED" <<'PY'
import sys
p, guard, mutated = sys.argv[1], sys.argv[2], sys.argv[3]
s = open(p, encoding='utf-8').read()
assert guard in s, "guard not found — the fixture has rotted"
open(p, 'w', encoding='utf-8').write(s.replace(guard, mutated, 1))
print("  mutated the COPY at %s" % p)
PY
out=$(run "$mut"); rc=$?
printf '%s\n' "$out" | grep -iE 'out of memory|^FAIL|ok |fatal error|timed out|panic' \
  | head -5 | sed 's/^/  /'
if [ $rc -eq 0 ]; then
  echo "WRONG: the test still passed with the cap removed — it is vacuous and"
  echo "       proves nothing about the bound."
  exit 1
fi
echo "ok    cap removed -> $TEST fails (rc=$rc), contained by the caps above"
echo

echo "== the repository was never written to"
if git -C "$REPO" diff --quiet -- internal/errs/errs.go; then
  echo "ok    internal/errs/errs.go is unchanged by this script"
else
  echo "WRONG: internal/errs/errs.go differs from HEAD — investigate before"
  echo "       committing anything."
  exit 1
fi
echo
echo "PROOF HOLDS: the bound is load-bearing, and the test detects its absence."
exit 0
