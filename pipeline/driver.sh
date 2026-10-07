#!/usr/bin/env bash
# pipeline/driver.sh — the continuous build pipeline supervisor.
#
# No LLM lives in this script. It owns: worktree lifecycle, brief rendering,
# capped child runs (implementer/reviewer/fixer via `opencode run`), the
# scoped gates, the file-ownership diff, lane-branch commits, and queue.json
# status transitions. It never edits tracked repo files itself and never
# merges: PR creation is the principal's, merging is the product owner's
# (plan §4 H1, §7).
#
# Usage:
#   driver.sh status                     # queue table
#   driver.sh next                       # next schedulable WP
#   driver.sh start  <wp>                # worktree + implementer + gates + commit
#   driver.sh review <wp>                # read-only reviewer in the worktree
#   driver.sh fix    <wp> <findings>     # fix round + gates + commit
#   driver.sh gates  <wp>                # re-run gates only
#   driver.sh mark   <wp> <status> [pr]  # manual transition (principal)
#   driver.sh brief  <wp> <role>         # render a brief only (debug)
#   driver.sh commit <wp> [message]      # checkpoint-commit a lane (small+often)
#   driver.sh collect <wp>               # show run artifacts
#   driver.sh cleanup <wp>               # remove worktree (branch stays)
#   driver.sh loop [--for <dur>]         # 24/7 mode; dur: 90m / 8h / 3600s
#
# <wp> is the queue's wp value (10, 15, "IR-3", ...). <status> values: see
# queue.json's $comment. Requires: bash, git, python3, go, opencode on PATH.
set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/.." && pwd)"
RUNS="$HERE/runs"
QUEUE="$HERE/queue.json"

# shellcheck source=/dev/null
[ -f "$HERE/models.env" ] && . "$HERE/models.env"
LANES_DIR="${LANES_DIR:-$HOME/Sleipnir-lanes}"
IMPLEMENTER_MODEL="${IMPLEMENTER_MODEL:?models.env: IMPLEMENTER_MODEL unset}"
REVIEWER_MODEL="${REVIEWER_MODEL:?models.env: REVIEWER_MODEL unset}"
PRINCIPAL_MODEL="${PRINCIPAL_MODEL:?models.env: PRINCIPAL_MODEL unset}"
FIXER_MODEL="${FIXER_MODEL:-$IMPLEMENTER_MODEL}"
IMPLEMENTER_TIMEOUT_S="${IMPLEMENTER_TIMEOUT_S:-14400}"
REVIEWER_TIMEOUT_S="${REVIEWER_TIMEOUT_S:-7200}"
FIXER_TIMEOUT_S="${FIXER_TIMEOUT_S:-7200}"
MAX_FIX_ROUNDS="${MAX_FIX_ROUNDS:-2}"
MAX_OPEN_PRS="${MAX_OPEN_PRS:-2}"
AWAITING_HUMAN_MAX="${AWAITING_HUMAN_MAX:-3}"
LOOP_SLEEP_S="${LOOP_SLEEP_S:-300}"
DIFF_WARN_LINES="${DIFF_WARN_LINES:-5000}"
MIN_STAGE_S="${MIN_STAGE_S:-1800}"
GOMEMLIMIT="${GOMEMLIMIT:-512MiB}"; export GOMEMLIMIT
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-300s}"; export GO_TEST_TIMEOUT

# Deadline handling: loop --for sets CHILD_DEADLINE (epoch seconds); every
# capped child run gets min(its stage cap, time left), so a timed run never
# overshoots the owner's chosen window by more than the running stage.
CHILD_DEADLINE="${CHILD_DEADLINE:-0}"

now_s() { date +%s; }

eff_cap() { # eff_cap <stage-cap-secs> — capped by the run deadline when set
  local cap="$1"
  if [ "$CHILD_DEADLINE" -gt 0 ]; then
    local rem=$((CHILD_DEADLINE - $(now_s)))
    [ "$rem" -lt 60 ] && rem=60
    [ "$rem" -lt "$cap" ] && cap="$rem"
  fi
  echo "$cap"
}

parse_duration() { # parse_duration 90m|8h|3600s|3600 -> seconds
  local d="$1"
  case "$d" in
    *h) echo $(( ${d%h} * 3600 )) ;;
    *m) echo $(( ${d%m} * 60 )) ;;
    *s) echo "${d%s}" ;;
    *)  echo "$d" ;;
  esac
}

log() { printf '%s %s\n' "$(date '+%F %T')" "$*"; }
die() { log "ERROR: $*"; exit 1; }

lane_dir() { echo "$LANES_DIR/wp-$1"; }
run_dir()  { echo "$RUNS/wp-$1"; }

# --- queue access (python3 is the JSON engine) ------------------------------

qget() { # qget <wp> <field> — prints "" when absent
  python3 -c '
import json, sys
q = json.load(open(sys.argv[1]))
for e in q["work"]:
    if str(e["wp"]) == str(sys.argv[2]):
        v = e.get(sys.argv[3], "")
        print(v if not isinstance(v, list) else "\n".join(map(str, v)))
        break
' "$QUEUE" "$1" "$2"
}

qset() { # qset <wp> <field> <value>
  python3 -c '
import json, sys
p, wp, f, v = sys.argv[1:5]
q = json.load(open(p))
for e in q["work"]:
    if str(e["wp"]) == str(wp):
        e[f] = v
        break
else:
    sys.exit("no queue entry for wp " + wp)
json.dump(q, open(p, "w"), indent=2)
open(p, "a").write("\n")
' "$QUEUE" "$1" "$2" "$3"
}

schedulable() { # prints wp ids whose deps are merged and status is ready
  python3 -c '
import json, sys
q = json.load(open(sys.argv[1]))
st = {str(e["wp"]): e.get("status", "") for e in q["work"]}
for e in q["work"]:
    if e.get("status") != "ready":
        continue
    if all(st.get(str(d)) == "merged" for d in e.get("deps", [])):
        print(e["wp"])
' "$QUEUE"
}

awaiting_human() { # count of pr_open + principal_validation lanes
  python3 -c '
import json, sys
q = json.load(open(sys.argv[1]))
print(sum(1 for e in q["work"] if e.get("status") in ("pr_open", "principal_validation")))
' "$QUEUE"
}

# --- capped child runs (AGENTS.md 2026-09-21: never uncapped) ---------------

run_capped() { # run_capped <secs> <logfile> <cmd...> — kills the whole group
  local secs="$1" log="$2"; shift 2
  local t0=$SECONDS rc
  perl -e 'setpgrp(0,0); exec @ARGV or exit 127' "$@" >"$log" 2>&1 &
  local pid=$!
  ( sleep "$secs"; kill -TERM -- "-$pid" 2>/dev/null
    sleep 30;  kill -KILL -- "-$pid" 2>/dev/null ) &
  local wpid=$!
  wait "$pid"; rc=$?
  kill "$wpid" 2>/dev/null; wait "$wpid" 2>/dev/null
  if [ $rc -ne 0 ] && [ $((SECONDS - t0)) -ge "$secs" ]; then rc=124; fi
  return $rc
}

# --- brief rendering ---------------------------------------------------------

render_brief() { # render_brief <wp> <role: implementer|reviewer|fixer> <findings-file>
  local wp="$1" role="$2" findings="${3:-}"
  local rd; rd="$(run_dir "$wp")"; mkdir -p "$rd"
  python3 - "$QUEUE" "$wp" "$role" "$findings" "$rd" "$LANES_DIR" "$HERE/templates" <<'EOF'
import json, sys
queue, wp, role, findings, rd, lanes, tpldir = sys.argv[1:8]
entry = next(e for e in json.load(open(queue))["work"] if str(e["wp"]) == str(wp))
tpl = open(f"{tpldir}/brief-{role}.md").read()
subs = {
    "WP": str(entry["wp"]), "SLUG": entry.get("slug", ""),
    "CLAUSES": entry.get("clauses", ""), "FILES": "\n".join(entry.get("files", [])),
    "TESTS": entry.get("tests", ""), "GATE": entry.get("gate", ""),
    "NOTES": entry.get("notes", ""),
    "LANE": f"{lanes}/wp-{entry['wp']}",
    "REPORT": f"{lanes}/wp-{entry['wp']}/pipeline/runs/wp-{entry['wp']}/report.md",
    "RUNS": f"{lanes}/wp-{entry['wp']}/pipeline/runs/wp-{entry['wp']}",
    "FINDINGS": open(findings).read() if findings else "",
}
for k, v in subs.items():
    tpl = tpl.replace("{{" + k + "}}", v)
out = f"{rd}/brief-{role}.md"
open(out, "w").write(tpl)
print(out)
EOF
}

# --- ownership diff (plan §3.2: mechanical, not trusted) ---------------------

ownership_ok() { # ownership_ok <wp> <lane> — exit 0 when every change is inside files[]
  python3 - "$QUEUE" "$1" "$2" <<'EOF'
import json, subprocess, sys, fnmatch
queue, wp, lane = sys.argv[1], sys.argv[2], sys.argv[3]
entry = next(e for e in json.load(open(queue))["work"] if str(e["wp"]) == str(wp))
globs = entry.get("files", [])
out = subprocess.run(["git", "-C", lane, "status", "--porcelain"],
                     capture_output=True, text=True).stdout
bad = []
for line in out.splitlines():
    path = line[3:].strip().strip('"')
    if " -> " in path:
        path = path.split(" -> ")[-1]
    if path.startswith("pipeline/runs/"):
        continue  # gitignored run artifacts
    ok = any(path.startswith(g) if g.endswith("/") else fnmatch.fnmatch(path, g)
             for g in globs)
    if not ok:
        bad.append(path)
if bad:
    print("OWNERSHIP VIOLATION — changes outside the queue entry's files:")
    for p in bad:
        print("  " + p)
    sys.exit(1)
print("ownership: clean (all changes within the declared file list)")
EOF
}

commit_lane() { # commit_lane <wp> <lane> <message>
  local wp="$1" lane="$2" msg="$3"
  git -C "$lane" add -A
  if git -C "$lane" diff --cached --quiet; then
    log "wp-$wp: nothing to commit"
    return 1
  fi
  # Small-change discipline: warn (never block) on oversized lane diffs so
  # the principal can re-cut the WP into smaller queue entries.
  local lines
  lines=$(git -C "$lane" diff --cached --numstat | awk '{a+=$1; d+=$2} END {print a+d+0}')
  if [ "$lines" -gt "$DIFF_WARN_LINES" ]; then
    log "wp-$wp: WARNING diff is $lines changed lines (> DIFF_WARN_LINES=$DIFF_WARN_LINES) — consider splitting the WP"
  fi
  git -C "$lane" commit -q -m "$msg" && log "wp-$wp: committed $lines lines to $(git -C "$lane" branch --show-current)"
}

# --- stages ------------------------------------------------------------------

require_lane() { [ -d "$1/.git" ] || [ -f "$1/.git" ] || die "no worktree at $1 (run: driver.sh start <wp>)"; }

cmd_start() {
  local wp="$1"
  local slug lane rd; slug="$(qget "$wp" slug)"; lane="$(lane_dir "$wp")"; rd="$(run_dir "$wp")"
  [ -n "$slug" ] || die "wp $wp not in queue"
  [ -d "$lane" ] && die "worktree $lane already exists (cleanup first)"
  mkdir -p "$rd" "$LANES_DIR"
  git -C "$REPO" fetch -q origin || log "wp-$wp: fetch failed (offline? continuing from local main)"
  if [ "$(git -C "$REPO" rev-parse main)" != "$(git -C "$REPO" rev-parse origin/main 2>/dev/null || echo x)" ]; then
    log "wp-$wp: WARNING local main != origin/main — worktree is created from LOCAL main"
  fi
  git -C "$REPO" worktree add "$lane" -b "wp/${wp}-${slug}" main || die "worktree add failed"
  mkdir -p "$lane/pipeline/runs/wp-$wp"
  qset "$wp" status implementing
  qset "$wp" branch "wp/${wp}-${slug}"
  local brief; brief="$(render_brief "$wp" implementer)"
  local cap; cap="$(eff_cap "$IMPLEMENTER_TIMEOUT_S")"
  log "wp-$wp: implementer starting (model $IMPLEMENTER_MODEL, cap ${cap}s)"
  run_capped "$cap" "$rd/implementer.log" \
    opencode run --agent sleipnir-implementer --model "$IMPLEMENTER_MODEL" \
      --auto --title "WP-$wp implementer" "$(< "$brief")"
  local rc=$?
  if [ $rc -eq 124 ]; then
    # Commit-frequently directive: preserve the partial work as a WIP commit
    # on the lane branch before parking it — nothing a child wrote is lost.
    commit_lane "$wp" "$lane" "WP-$wp $slug: WIP at timeout (pipeline)" || true
    qset "$wp" status timed_out
    die "wp-$wp: implementer TIMED OUT after ${cap}s (group killed; WIP committed, worktree kept)"
  fi
  if [ $rc -ne 0 ]; then log "wp-$wp: implementer exited $rc (see $rd/implementer.log) — continuing to gates"
  else cp "$lane/pipeline/runs/wp-$wp/report.md" "$rd/report.md" 2>/dev/null || log "wp-$wp: WARNING no report file found"; fi
  cmd_gates "$wp" || { qset "$wp" status gates_failed; die "wp-$wp: GATES FAILED (lane kept for inspection)"; }
  ownership_ok "$wp" "$lane" > "$rd/ownership.log" 2>&1 || { cat "$rd/ownership.log"; qset "$wp" status ownership_violation; die "wp-$wp: ownership violation (lane NOT committed)"; }
  commit_lane "$wp" "$lane" "WP-$wp $slug: lane delivery (pipeline, unreviewed)" || true
  qset "$wp" status gates_passed
  log "wp-$wp: start complete — gates passed, committed on wp/${wp}-${slug}; next: driver.sh review $wp"
}

cmd_gates() {
  local wp="$1" lane; lane="$(lane_dir "$wp")"; require_lane "$lane"
  local rd; rd="$(run_dir "$wp")"; mkdir -p "$rd"
  log "wp-$wp: running gates in $lane"
  "$HERE/gates.sh" "$lane" 2>&1 | tee "$rd/gates.log" | tail -3
  [ "${PIPESTATUS[0]}" -eq 0 ]
}

cmd_review() {
  local wp="$1" lane rd; lane="$(lane_dir "$wp")"; rd="$(run_dir "$wp")"; require_lane "$lane"; mkdir -p "$rd"
  local before after; before="$(git -C "$lane" status --porcelain)"
  local model="$REVIEWER_MODEL"
  [ "$(qget "$wp" reviewer_model)" = "strong" ] && model="$PRINCIPAL_MODEL"
  local brief; brief="$(render_brief "$wp" reviewer)"
  local cap; cap="$(eff_cap "$REVIEWER_TIMEOUT_S")"
  log "wp-$wp: reviewer starting (model $model, cap ${cap}s)"
  run_capped "$cap" "$rd/reviewer.log" \
    opencode run --agent sleipnir-reviewer --model "$model" \
      --auto --title "WP-$wp reviewer" "$(< "$brief")"
  local rc=$?
  after="$(git -C "$lane" status --porcelain)"
  if [ "$before" != "$after" ]; then
    log "wp-$wp: REVIEWER VIOLATION — the worktree changed during review:"
    diff <(echo "$before") <(echo "$after") | sed 's/^/    /'
    qset "$wp" status reviewer_violation
    die "wp-$wp: review discarded (read-only rule broken)"
  fi
  if [ $rc -eq 124 ]; then qset "$wp" status timed_out; die "wp-$wp: reviewer TIMED OUT after ${REVIEWER_TIMEOUT_S}s"; fi
  cp "$lane/pipeline/runs/wp-$wp/report.md" "$rd/review.md" 2>/dev/null || log "wp-$wp: WARNING no review report found at $rd/review.md"
  qset "$wp" status reviewed
  log "wp-$wp: review complete — report: $rd/review.md"
}

cmd_fix() {
  local wp="$1" findings="${2:?usage: driver.sh fix <wp> <findings-file>}"
  local lane rd; lane="$(lane_dir "$wp")"; rd="$(run_dir "$wp")"; require_lane "$lane"
  [ -f "$findings" ] || die "findings file not found: $findings"
  local round=1; while [ -f "$rd/fix-$round.log" ]; do round=$((round + 1)); done
  [ "$round" -le "$MAX_FIX_ROUNDS" ] || die "wp-$wp: MAX_FIX_ROUNDS ($MAX_FIX_ROUNDS) exhausted — principal must take over"
  local brief; brief="$(render_brief "$wp" fixer "$findings")"
  qset "$wp" status implementing
  local cap; cap="$(eff_cap "$FIXER_TIMEOUT_S")"
  log "wp-$wp: fix round $round starting (model $FIXER_MODEL, cap ${cap}s)"
  run_capped "$cap" "$rd/fix-$round.log" \
    opencode run --agent sleipnir-implementer --model "$FIXER_MODEL" \
      --auto --title "WP-$wp fixer r$round" "$(< "$brief")"
  local rc=$?
  [ $rc -eq 124 ] && { commit_lane "$wp" "$lane" "WP-$wp $(qget "$wp" slug): WIP at fix timeout (pipeline)" || true; qset "$wp" status timed_out; die "wp-$wp: fixer TIMED OUT"; }
  cmd_gates "$wp" || { qset "$wp" status gates_failed; die "wp-$wp: GATES FAILED after fix round $round"; }
  ownership_ok "$wp" "$lane" > "$rd/ownership-$round.log" 2>&1 || { cat "$rd/ownership-$round.log"; qset "$wp" status ownership_violation; die "wp-$wp: ownership violation after fix"; }
  commit_lane "$wp" "$lane" "WP-$wp $(qget "$wp" slug): fix round $round (pipeline)" || true
  qset "$wp" status gates_passed
  log "wp-$wp: fix round $round complete — next: driver.sh review $wp"
}

cmd_mark() {
  local wp="$1" status="${2:?usage: driver.sh mark <wp> <status> [pr-number]}" pr="${3:-}"
  qset "$wp" status "$status"
  [ -n "$pr" ] && qset "$wp" pr "$pr"
  log "wp-$wp: status -> $status${pr:+ (PR #$pr)}"
}

cmd_commit() { # checkpoint commit (owner directive: small changes, commit often)
  local wp="$1" msg="${2:-WP-$wp $(qget "$wp" slug): checkpoint (pipeline)}"
  local lane; lane="$(lane_dir "$wp")"; require_lane "$lane"
  ownership_ok "$wp" "$lane" || die "wp-$wp: ownership violation — checkpoint refused"
  commit_lane "$wp" "$lane" "$msg"
}

cmd_collect() {
  local wp="$1" rd; rd="$(run_dir "$wp")"
  echo "== artifacts for wp-$wp ($rd)"; ls -la "$rd" 2>/dev/null
  [ -f "$rd/review.md" ] && { echo; echo "== review report =="; cat "$rd/review.md"; }
}

cmd_cleanup() {
  local wp="$1" lane; lane="$(lane_dir "$wp")"
  [ -d "$lane" ] && git -C "$REPO" worktree remove --force "$lane" && log "wp-$wp: worktree removed (branch kept)"
}

cmd_status() {
  python3 -c '
import json, sys
q = json.load(open(sys.argv[1]))
print("%5s  %-24s %-13s %-22s %4s" % ("wp", "slug", "tier", "status", "pr"))
for e in q["work"]:
    if e.get("status") == "merged":
        continue
    print("%5s  %-24s %-13s %-22s %4s" % (
        str(e["wp"]), e.get("slug", ""), e.get("tier", ""),
        e.get("status", ""), str(e.get("pr", ""))))
merged = [str(e["wp"]) for e in q["work"] if e.get("status") == "merged"]
print("merged:", ",".join(merged))
' "$QUEUE"
}

cmd_next() {
  local n; n="$(schedulable | head -5)"
  if [ -z "$n" ]; then echo "nothing schedulable (deps unmerged, all blocked, or queue empty)"; return 1; fi
  echo "$n"
}

cmd_loop() {
  local dur=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --for) dur="${2:?--for needs a duration: 90m / 8h / 3600s}"; shift 2 ;;
      *) die "loop: unknown option $1 (usage: driver.sh loop [--for <dur>])" ;;
    esac
  done
  if [ -n "$dur" ]; then
    CHILD_DEADLINE=$(( $(now_s) + $(parse_duration "$dur") ))
    log "loop: timed run — deadline $(date -r "$CHILD_DEADLINE" '+%F %T') (--for $dur)"
  else
    log "loop: untimed run (stop with: pkill -f 'driver.sh loop'; running children keep their own caps)"
  fi
  log "loop: started (max open PRs $MAX_OPEN_PRS, awaiting-human cap $AWAITING_HUMAN_MAX, sleep ${LOOP_SLEEP_S}s)"
  while true; do
    if [ "$CHILD_DEADLINE" -gt 0 ]; then
      local rem=$((CHILD_DEADLINE - $(now_s)))
      if [ "$rem" -lt "$MIN_STAGE_S" ]; then
        log "loop: deadline reached (remaining ${rem}s < MIN_STAGE_S) — finishing"
        break
      fi
    fi
    local ah; ah="$(awaiting_human)"
    if [ "$ah" -ge "$AWAITING_HUMAN_MAX" ]; then
      log "loop: $ah lanes awaiting human (>= $AWAITING_HUMAN_MAX) — paused (plan §7: never outrun H1)"
      if [ "$CHILD_DEADLINE" -gt 0 ]; then log "loop: timed run does not idle on a full human queue — finishing"; break; fi
      sleep "$LOOP_SLEEP_S"; continue
    fi
    local wp; wp="$(schedulable | head -1 || true)"
    if [ -z "$wp" ]; then
      log "loop: nothing schedulable — sleeping"
      if [ "$CHILD_DEADLINE" -gt 0 ]; then log "loop: timed run with empty queue — finishing"; break; fi
      sleep "$LOOP_SLEEP_S"; continue
    fi
    log "loop: picking up wp-$wp"
    if cmd_start "$wp" && cmd_review "$wp"; then
      local rd; rd="$(run_dir "$wp")"
      if [ -f "$rd/review.md" ] && grep -q "^## MUST FIX" "$rd/review.md" \
         && ! grep -A1 "^## MUST FIX" "$rd/review.md" | grep -qi "^\*\*None"; then
        log "loop: wp-$wp has MUST FIX findings — fix round"
        cmd_fix "$wp" "$rd/review.md" && cmd_review "$wp" || log "loop: wp-$wp fix/re-review failed — parking for principal"
      fi
      qset "$wp" status principal_validation
      log "loop: wp-$wp parked at principal_validation (report: $(run_dir "$wp")/review.md)"
    else
      log "loop: wp-$wp stopped at $(qget "$wp" status) — parking; next iteration skips it"
    fi
    sleep 30
  done
  echo
  log "loop: run summary"
  cmd_status
}

# --- dispatch ----------------------------------------------------------------

cmd="${1:-help}"; shift || true
case "$cmd" in
  status)  cmd_status ;;
  next)    cmd_next ;;
  start)   cmd_start "${1:?wp required}" ;;
  gates)   cmd_gates "${1:?wp required}" ;;
  review)  cmd_review "${1:?wp required}" ;;
  fix)     cmd_fix "${1:?wp required}" "${2:-}" ;;
  mark)    cmd_mark "${1:?wp required}" "${2:-}" "${3:-}" ;;
  brief)   render_brief "${1:?wp required}" "${2:?role required}" ;;
  commit)  cmd_commit "${1:?wp required}" "${2:-}" ;;
  collect) cmd_collect "${1:?wp required}" ;;
  cleanup) cmd_cleanup "${1:?wp required}" ;;
  loop)    cmd_loop "$@" ;;
  *) sed -n '2,30p' "$0" ;;
esac
