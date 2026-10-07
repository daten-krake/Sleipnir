#!/usr/bin/env bash
# pipeline/gates.sh — the WORKFLOW §4 gates for one lane worktree.
#
# Usage: gates.sh <worktree-dir>
#
# Runs the five merge gates plus -race and the contract-vector check, scoped
# to the worktree (which contains main + the lane's changes only). Every go
# command is capped (AGENTS.md 2026-09-21: never run uncapped). Exits 0 only
# if all gates pass; prints a PASS/FAIL line per gate.
set -uo pipefail

WT="${1:?usage: gates.sh <worktree-dir>}"
cd "$WT" || exit 2

GOMEMLIMIT="${GOMEMLIMIT:-512MiB}"
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-300s}"
export GOMEMLIMIT

FAIL=0
run() { # run <label> <command...>
  local label="$1"; shift
  printf '%-60s' "$label"
  if out=$("$@" 2>&1); then
    if [ -n "$out" ]; then echo "PASS (output below)"; echo "$out" | sed 's/^/    /'; else echo "PASS"; fi
  else
    echo "FAIL"; echo "$out" | sed 's/^/    /'; FAIL=1
  fi
}

run "gofmt -l . (must be empty)"        test -z "$(gofmt -l .)"
run "go vet ./..."                      go vet ./...
run "go build ./..."                    go build ./...
run "go test ./... (capped)"            go test -timeout "$GO_TEST_TIMEOUT" -count=1 ./...
run "go test -race ./... (capped)"      go test -timeout "$GO_TEST_TIMEOUT" -count=1 -race ./...

if [ -f docs/reviews/2026-09-11-verify-vectors.py ]; then
  printf '%-60s' "contract vectors (verify-vectors.py)"
  if vout=$(python3 docs/reviews/2026-09-11-verify-vectors.py 2>&1) \
     && echo "$vout" | grep -qE '^PASS [0-9]+ / FAIL 0$'; then
    echo "PASS ($(echo "$vout" | tail -1))"
  else
    echo "FAIL"; echo "$vout" | sed 's/^/    /'; FAIL=1
  fi
fi

echo
if [ "$FAIL" -eq 0 ]; then echo "GATES: ALL PASS"; else echo "GATES: FAILED"; fi
exit "$FAIL"
