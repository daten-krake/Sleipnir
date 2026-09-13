# A0 / A1 / A2 contract review — adversarial architecture pass

- Reviewer: `sleipnir-architect` (read-only run; no repository file touched)
- Repo: `/home/wtadmin/Sleipnir`, branch `contracts/a0-a2-conventions`
- Documents: `contracts/A0-conventions.md` (773 L), `contracts/A1-events.md` (2472 L),
  `contracts/A2-graph.md` (1200 L), `contracts/README.md`; authority inputs
  `adr/*`, `SPEC.md`, `DESIGN.md`, `AGENTS.md`, `WORKFLOW.md`,
  `sessions/2026-09-04-program-layout.md` (Q1–Q15).
- Authority order applied: ADR > SPEC > DESIGN > contract.
- Threat model applied: hostile orchestrator, hostile worker, compromised remote
  agent node, careless/insider operator, store-write-capable attacker (A11).
- Already verified by parent (not re-checked): the A1 §4.3 chain vector reproduces
  byte-exactly.

Counts: **MUST FIX 28 · SHOULD FIX 20 · NICE 9** (57 findings).

Severity legend: MUST FIX = an attack gets through, evidence can be destroyed, a
binding ADR/constraint is contradicted, or the build cannot be deterministic.
SHOULD FIX = weakened defence, unenforceable rule, or a defect that will be
implemented wrong. NICE = wording/consistency.

---

## 1. ADR / SPEC conflicts

### C-01 · MUST FIX · A1-6.5 (vs Q11, SPEC §3)
Override authority is self-contradictory and stricter than Q11.
"Only a **user** principal with the admin role may override" is followed by "an
operator-scoped user MAY NOT override **another engagement's** break", which
states a rule that only makes sense if operators may override their own.
Q11 (binding PO decision) says "explicit **operator** override allowed".
Attack/failure: two builders implement two different authz rules on the one
control that releases a customer artifact from a damaged chain; the permissive
reading lets an assigned operator release a report whose integrity failed,
with no admin separation (adversarial A15 insider).
Replacement: pick one and delete the other sentence. Recommended wording:
> "Only a `user` principal holding the **admin** role may compose
> `integrity_override` (SPEC §3 places integrity-class controls next to the hard
> stop; Q11's "operator" reads as "human"). An operator-scoped user — including
> one assigned to the engagement — MUST NOT override any break. A5 MUST gate the
> endpoint on the admin role and MUST return `forbidden` (A0-3.1, principal-level,
> naming no object) to an operator. Record the narrowing of Q11 as an explicit PO
> decision in §6.6 before Freeze."

### C-02 · MUST FIX · A1-3.3 `approval_requested` / `approval_executed` (vs ADR-0018 §1, §3)
The approved action is not in the audit trail. ADR-0018 §1 defines the
fingerprint as *tool id + exact arguments + target(s) + hash of the full action
spec*, and §3 requires "the executed action and its fingerprint are recorded in
the event log next to the approval record". A1-3.3 records only
`fingerprint_hash` (64 hex), prose `action_summary` (512 B, untrusted) and
`target`; A1-4.3 forbids content in payloads; neither kind carries an
`*_evidence_id`, so A1-7.3 derives `evidence_refs:[]`.
Attack/failure: an operator approves fingerprint `X`; nobody — auditor, customer,
incident responder, ADR-0018 §2 re-validation reviewer — can ever recover *what*
`X` was. The 64-hex digest is unlinkable to any stored bytes, so the "audit
pairing is evidence-grade" consequence of ADR-0018 fails, and a substituted
action spec that hashes to a *different* fingerprint is indistinguishable from
a lost one.
Replacement: add to both kinds (pre-Freeze, additive):
> `approval_requested`: `action_spec_evidence_id:string` (`evi_`, MUST be
> non-empty; the platform MUST store the exact canonical bytes (A0-2) of the A7
> action spec as a write-once evidence artifact at request time, and MUST record
> its `evi_` id here and in `evidence_refs`).
> `approval_executed`: `action_spec_evidence_id:string` MUST equal the one on the
> `approval_requested` event with the same `approval_id` (A1-4.2 obligation).
> A7 MUST define the action-spec document whose canonical bytes that artifact
> holds; `fingerprint_hash` MUST be the A0-2.15 digest of exactly those bytes.

### C-03 · MUST FIX · A1-4.2 `container_killed` (vs A1-7.12, ADR-0005 §4, SPEC §6)
Unimplementable ordering obligation on the kill path. A1-4.2:
"`kill_reason:"hard_stop"` MUST be preceded in the chain by a `hard_stop_fired`
event". A1-7.12: the kill happens **first**, its event is written **after** and
retried. Two goroutines (kill-all-containers, compose hard_stop_fired) race; if
a `container_killed{hard_stop}` row wins the chain lock, A1-7.10 step 8 rejects
it → the platform has killed a container with **no record of the kill**, which
is exactly the evidence destruction A1-7.11 exists to prevent. If instead the
validator is lenient, the clause is dead letter and T-matrix-style tests cannot
assert it.
Replacement:
> `container_killed{kill_reason:"hard_stop"}` MUST carry `stop_event_id:string`
> (`evt_`, the `hard_stop_fired` event it answers). The platform MUST commit
> `hard_stop_fired` **before** it issues any kill, in its own transaction, and
> MUST NOT block the kill on that commit (A1-7.12): if the commit fails, the
> platform MUST kill anyway and MUST retry the append until it lands. The
> ordering obligation is therefore "correlated by `stop_event_id`", not
> "preceded in `seq` order"; a `container_killed{hard_stop}` whose
> `stop_event_id` is empty or unresolvable is a platform defect → `internal`.

### C-04 · MUST FIX · A1 §6.9 + A1-3.3 `scope_changed` (vs ADR-0005 §3, C5, A2-8.5)
The **global** blacklist — the control that "beats allowlist and cannot be
overridden by scopes or approvals" — has no audit record and no owner. A1-3.3's
`scope_changed` is engagement-scoped (`engagement_id` is chain scope, A1-5.1);
A1 §6.9 explicitly parks "global-blacklist writes outside an engagement" in an
unowned gap; SPEC §3 gives the admin the global blacklist.
Attack/failure: an insider admin removes `10.99.0.0/16` from the global blacklist
while runs are live. No engagement chain records it, A2-8.5's mandated
per-engagement `quarantine_recomputed{blacklist_changed}` has no
`trigger_event_id` to point at (A1-4.2 requires it to resolve *in this
engagement*), previously quarantined nodes are silently released, and the
customer report shows "not tested" for a range that was in fact tested.
Replacement:
> A1-3.3 `scope_changed` MUST also be composed into **every affected engagement
> chain** for a global blacklist change, with `change_kind:blacklist_added` /
> `blacklist_removed` and `entry`/`entry_hash` of the global entry; the
> composing subsystem is `scope`, `actor` is the admin user, and `run_id` is
> `""`. The `quarantine_recomputed{blacklist_changed}` event of each engagement
> MUST reference that engagement's own `scope_changed` event as
> `trigger_event_id` (A1-4.2 stays satisfiable). A1 §6.9's deferral MUST NOT
> cover blacklist mutation: it is engagement-reachable safety state, not
> session audit.

### C-05 · MUST FIX · A2-2.8 / A2-8.7 / A2-8.8 (vs Q5, ADR-0016 §1)
No MUST restricts `report_excluded` to quarantined nodes. A2-8.7's prose
describes only the quarantined case ("MAY exclude a quarantined node"), while
A2-2.8's authority table grants the operator the field unconditionally, A2-8.8
adds only "not by a machine principal", and A2-12.5 requires reporting views to
omit **every** `report_excluded` node. Q5 authorizes removal of *quarantined
discoveries* from the report only.
Attack/failure: a careless or insider operator (A15) excludes a confirmed
`critical` `finding` node from the customer report. The exclusion is audited
(A1 `report_inclusion_changed`) but the customer receives a report that omits a
known compromise — the exact evidence-suppression path ADR-0016 §1's
"graph content is evidence" rules out.
Replacement:
> `report_excluded` MUST be settable to `true` only on a node with
> `quarantined:true` (Q5). An attempt on a non-quarantined node →
> `conflict` (A0-3.1, immutable state) and MUST be recorded as
> `action_blocked{reason:"graph_write_rejected", action_kind:"graph_write"}`.
> Removing a confirmed finding from a report is expressed by a revision
> (A2-4) to `status:"refuted"` or `severity:"info"`, never by a flag.

### C-06 · MUST FIX · A1-3.3 `graph_node_quarantined{operator_release}` (vs A2-8.5, ADR-0016 §2)
A1 offers a direct operator release of quarantine; A2 forbids it. A2-8.5:
"`out_of_scope` changes **only** when an operator changes the engagement scope";
A2-2.8: `quarantined` changes "only as a consequence of an operator
scope/blacklist change". ADR-0016 §2: an out-of-scope node "can **never** be a
target of planned actions".
Attack/failure: an operator (or an injected UI request from a careless session)
releases a single out-of-scope host node without touching scope; the node
becomes current in planning views (A2-12.5) and targetable, while the
engagement's allowlist still does not contain it — the only thing standing
between the agent and an out-of-scope target is then the scope engine's string
match, which A2-8.2's lexical derivation (see S-04) does not guarantee.
Replacement: delete `operator_release` from the `quarantine_kind` enum (it is
additive-only after Freeze, so it must go now). A release is expressed only as
`quarantine_recomputed{scope_changed}` + per-node
`graph_node_quarantined{out_of_scope_discovery}` … i.e. by the recomputation
that a scope change causes. If the PO wants a manual release, it MUST be a new
ADR that also amends ADR-0016 §2, and it MUST require the released target to be
inside the (possibly newly widened) allowlist at release time.

### C-07 · SHOULD FIX · A1-4.9 (bullet 3) and A2-9.3 (vs ADR-0020 §4, §5, C11)
Unsounded deduction: "because secrets never enter an event, no event read,
stage view, report or SSE frame can carry one to a model endpoint". The
premise is a `regexp` pattern scan over free text that a hostile worker
authors (`result_summary` 2048 B, `command` 2048 B, A2 `attrs` values 512 B,
`label` 128 B). ADR-0020 §5 already declares masking best-effort.
Attack: a worker base32-encodes an NTLM hash, splits it across two `attrs`
keys, or hex-dumps it with spaces; the scan misses; the string reaches a
`cloud_raw`/`cloud_masked` endpoint via a stage view (C11 violation).
Replacement:
> The scan is a **filter, not a guarantee**. Egress exclusion (ADR-0020 §4) is
> the enforcement point and MUST be applied independently at the gateway to
> every string that leaves the platform. Additionally, for any engagement whose
> policy is not `local_only`, the gateway MUST exclude by **kind**, not by
> pattern: no `credential` node, no node with `credential_kind` set, and no
> `attrs` of such a node may appear in a cloud request, and `llm_call`
> MUST record the exclusion in `excluded_secret_count`.

### C-08 · SHOULD FIX · A1 §6.9 (vs SPEC §3, ADR-0012 §1, adversarial A15)
Engagement **operator assignment** — the fact that decides who may approve
dangerous actions — has no event kind anywhere and is parked in an unowned gap.
Attack: an insider admin assigns themselves to an engagement, then approves
their own orchestrator's exploitation request; ADR-0012 §1's separation
("only assigned operators plus admins") is satisfied on paper and unaudited in
fact.
Replacement: add kind `engagement_assignment_changed` (platform · `scope`,
user-caused) with `user_id:string`, `role:enum{operator,viewer}`,
`change_kind:enum{added,removed}`, `run_id:""`; `engagement_id` is the chain
scope. Ownership of *authentication* audit may stay with A5, but
assignment/role changes inside an engagement belong in the engagement chain.

### C-09 · SHOULD FIX · A1-3 taxonomy (vs ADR-0013, Q8, Q9)
Machine-credential and node lifecycle is unauditable: Q8 mandates instant token
revocation on hard stop / run end, Q9 mandates node pairing plus revoke/wipe,
A1-7.5 has a `token_revoked` refusal reason and A1-3.3 has
`action_blocked{node_not_paired}` — but the closed 39 kinds contain no
`token_revoked`, `node_paired`, `node_revoked` or `node_wiped` event.
Attack/failure: a compromised Pi keeps appending `command_executed` after a
hard stop; the chain cannot show whether its credential was ever revoked, so
"the platform cut the node off" (ADR-0005 §4 respawn-proof lifecycle) is
unprovable and the post-stop appends look legitimate.
Replacement: add three kinds — `credential_revoked` (`subject_type:enum{job,
task,node,user_session}`, `subject_id:string`, `reason:enum{hard_stop,run_ended,
operator,node_lost,compromised}`), `node_paired`, `node_unpaired`
(`wipe_issued:bool`) — and require A1-7.12's kill path to compose
`credential_revoked` for every job/task/node token of the stopped scope.

### C-10 · NICE · A0-1.2 (vs A1 §6.2/AM-1, A2 §6.2/AM-1)
Both A1 and A2 declare themselves **freeze-blocked** on a human-principal prefix
that A0-1.2 does not register, and A0-1.2's table is the only place it can be
registered. Add the row now (`user` · `usr_` · `^usr_B{26}$` · 30) rather than
carrying two identical amendment requests through Freeze.

---

## 2. Safety holes

### S-01 · MUST FIX · A0-4.4 + A1-8.2 (+ A2-11.4 `TestCursorFromEngagementARejectedInB`)
A0-4.4's claim "a forged cursor yields **at worst an empty page or
`validation`**" is false, and A1-8.2's fix only covers the cross-engagement
case. The cursor is `{"k":<seq>,"id":<event_id>}` with **no MAC**, and the
platform is told to look up only `id`. A hostile or careless client sends
`{"k":5,"id":"<evt at seq 900>"}` inside its own engagement: the id resolves,
the platform seeks past `k=5`, and rows 6…900 are **silently skipped**.
A1-8.9 then guarantees "no skips, no duplicates" and A1-8.2 tells the client it
MAY detect gaps itself — an auditor paging with a doctored cursor concludes the
log has a gap, or (with a UI that hides seq) sees a clean page walk over a
hole. Evidence is not destroyed but the *completeness claim* is.
Replacement:
> A0-4.4: "…a forged cursor yields an empty page, `validation`, **or a page
> whose ordering value is the one the platform resolved — never a silently
> skipped range**. The platform MUST derive the seek position from the row it
> looks the cursor's `id` up to, and MUST ignore `k` for seeking; if `k`
> disagrees with that row's ordering value the response MUST be `validation`
> (A0-4.8). Cursors are not integrity-protected and grant nothing."
> A1-8.2: add the same sentence with `k` = `seq`, plus the negative test
> `TestCursorWithInconsistentKAndIDRejected`.

### S-02 · MUST FIX · A1-5.8 + A1-6.3 `break_kind` (vs Q11, adversarial A11)
Tail truncation is declared undetectable from inside the chain, and the three
v1 mitigations do not add up to a detection path: (a) the seam exposes no
delete — irrelevant to an attacker with store access; (b) `slog` emission every
1000 `seq` leaves **up to 999 rows** truncatable with no contradiction, and in
the default Docker deployment (C3, SPEC §10) container stdout lands on the same
host as PostgreSQL, so "a different store, with different access" is not true;
(c) exports carry the head hash only for engagements that were exported.
Critically, there is **no `break_kind` and no event** for "head regressed
against the out-of-band record", so even when the platform *does* notice, it
cannot record it, cannot set `integrity_state:"failed"`, and cannot block the
export (A1-6.4 keys everything off `break_kind`).
Attack: store-capable attacker deletes the last 400 rows, rewrites
`chain-head`, and the platform reports `verified` forever.
Replacement:
> A1-6.3 `break_kind` MUST add `head_regression`: the stored head
> `(head_seq, head_hash)` is **lower or different** from the highest head this
> platform recorded out-of-band for that engagement. A1-5.8's emission MUST be
> at every verification, at every append crossing a **100**-`seq` boundary, and
> at every `run_ended`/`hard_stop_fired`, and MUST be written to a store the
> event-store role cannot UPDATE or DELETE (a separate `chain_head_trail` table
> with `REVOKE UPDATE, DELETE`, A1-7.9's rule). Startup verification MUST
> compare against that trail and report `head_regression` as a break, with the
> same consequences as any other break (A1-6.4). If the PO accepts §6.7's
> webhook anchor, the trail MUST also be pushed there; if not, the residual
> (an attacker with both store and log access forges history) MUST be printed
> into A1-6.6's export wording.

### S-03 · MUST FIX · A1-3.3 `approval_executed` + A1-4.2 (vs Q10, ADR-0018 §2, C6)
Single-use approval consumption has **no atomicity requirement**. A1 records
`single_use_consumed:bool` as an after-the-fact field; nothing says the
consumption, the ADR-0018 §2 re-validation and the spawn authorization happen
in one transaction, and nothing puts a uniqueness constraint on
`(engagement_id, approval_id)` for executions.
Attack (hostile orchestrator — the party Q10 exists to constrain): issue N
concurrent spawn/execute requests citing one granted `approval_id`. All N
re-validate before any consumes; N dangerous actions run on one human decision.
Both `task_spawned` events carry the same `approval_id`, and the log looks
consistent.
Replacement:
> A1-4.2 `approval_executed`: the platform MUST consume the approval and
> authorize the execution **in one transaction** (or under the same
> per-engagement lock as A1-5.4's append), guarded by a uniqueness constraint on
> `(engagement_id, approval_id)` in the consumption table. A second attempt MUST
> fail **before** any container is created, with `conflict` (A0-3.1) and
> `action_blocked{reason:"approval_consumed"}`, and MUST NOT compose a second
> `approval_executed`. Contract tests: `TestSingleUseApprovalRaceConsumesOnce`
> (N concurrent executions, exactly one `approval_executed`, N−1
> `action_blocked{approval_consumed}`) and
> `TestConsumedApprovalCannotSpawnAgain`.

### S-04 · MUST FIX · A2-8.2 / A2-8.3 (vs ADR-0005 §2–§3, ADR-0016 §2, C5)
Quarantine is a **lexical** match over `label`, `addresses`, `cidr`, `domain`,
`sid` only. `summary`, `attrs`, `protocol`, and every `finding`/`hypothesis`
node (which has no identity field at all) are never evaluated; the ingest
mapping from A1 event text to node fields is nowhere specified, so the values
that *are* evaluated are derived from worker-authored strings
(`command_executed.target`, `task_result.result_summary`).
Attack: a hostile worker reports a target as `printer.corp.local` (or puts the
real address only in `attrs.ip`), the node ingests un-quarantined, appears in
planning views (A2-12.5), and the orchestrator proposes a spawn "against that
node". A2-8.3's prohibition is then evaluated against a target string the same
worker supplied — the blacklist (C5, "last line of defense") is bypassed by
naming, and A2-8.10's "the agent saw a forbidden target and did not touch it"
event never fires.
Replacement:
> A2-8.2 MUST list the complete evaluated field set **per node kind** and MUST
> state that a node whose kind has no identity field (`finding`, `hypothesis`,
> `evidence_ref`) is quarantined by **derivation from its edges**: if any
> non-retracted edge connects it to a quarantined node, it is quarantined with
> the same reason. Quarantine MUST be recomputed on (a) a scope/blacklist
> change (A2-8.5), (b) a revision that changes any identity field, (c) a new
> edge touching a quarantined node — with `quarantine_recomputed.trigger`
> extended by `node_written` and `edge_written`.
> A2-8.3 MUST add: "the target of a spawn or action is **never** taken from a
> graph field. The scope engine resolves the target itself (A7 action spec),
> and the graph node a request cites MUST be named by `gn_` id in the request so
> the quarantine check is on the id, not on a string. A request that cites a
> quarantined node id is refused **before** approval routing
> (`action_blocked{reason:"graph_write_rejected"}` is insufficient — add
> `reason:"target_quarantined"`)."

### S-05 · MUST FIX · A2-4.7 + A2-5.6 + A2-3.4 (vs ADR-0016 §1 "evidence, not opinion")
Content-hash dedup is **first-write-wins on provenance**, and it makes the
`verified` confidence grade unreachable.
(a) A2-4.7: a write whose `(engagement_id, kind, content_hash)` exists returns
the existing `gn_` and stores nothing — including nothing about the *new*
observation. The second observer's `event_id`, `task_id`, `agent_node_id` and
`confidence` are discarded. A compromised worker that writes a fabricated
`finding` first owns its provenance forever; every later legitimate
confirmation is absorbed into the attacker's node and the report cites the
attacker's event.
(b) A2-5.6 defines `verified` as "reproduced by a second, independent
observation" — but that second observation is a dedup hit, so it cannot raise
the grade (A2-5.6: "MUST NOT be raised by a later write"), and it cannot create
a revision either, because a revision with identical content has the same
`content_hash` and collapses (A2-4.7). `confidence:"verified"` is therefore
dead vocabulary, and Q2's "Finding carries confidence" is unimplemented.
Replacement:
> A2-4.7: a dedup collapse MUST NOT discard the observation. The platform MUST
> append the new provenance as an additional entry in the node's provenance set
> (`provenance` becomes a bounded list, ≤ 8 entries, each with its own
> `event_id`, `principal_kind`, `run_id`, `confidence`), MUST raise `confidence`
> to `verified` **only** when the new entry's `event_id` differs from every
> existing one and its `task_id`/`agent_node_id` differ (independence), and MUST
> emit the A1 `graph_node_written` event with a new field `dedup_hit:bool` so
> the collapse is visible in the chain. Provenance list order MUST be by
> `provenance[].event_id` `seq`, never by arrival, so the node's bytes are
> deterministic. `content_hash` MUST remain computed over content only
> (A2-4.6), so the collapse key is unchanged.

### S-06 · MUST FIX · A1-4.4 (bullet 3) + A1-3.3 `approval_requested.untrusted_context` (vs SPEC §6, ADR-0018 §4, adversarial A1/A9)
The anti-laundering rule has no computation and no test. A1-4.4 forbids a
platform subsystem from copying untrusted content into an unmarked field, but
the only enforcement named is "the contract test corpus asserts that claim per
kind" — no test id, no per-kind table of which fields may receive event text.
Worse, `approval_requested.untrusted_context` — the ADR-0018 §4 flag that tells
the operator "this request was shaped by injected content" — has **no
definition at all**: A1-4.2 does not constrain it, no clause says who computes
it or from what, and it is not in the `*`-marked list, so A1-4.4's
`untrusted` computation ignores it.
Attack: the approval service copies the orchestrator's `task_description`
(untrusted, 2048 B, model-authored) into `action_summary` and sets
`untrusted_context:false` (its zero value). The operator sees clean-looking
prose with no injection flag and approves. Nothing in the contract is violated
by any single clause — which is the defect.
Replacement:
> A1-4.2 `approval_requested`: `untrusted_context` is **platform-computed** and
> MUST be `true` iff the request event, or any event referenced transitively by
> `request_event_id` / `spawn_request_event_id`, carries `untrusted:true`, or
> iff any `*`-marked field of this payload is non-empty. It MUST NOT be a caller
> value. `action_summary` MUST be composed only from platform vocabulary and
> the A7 action spec's own fields; copying model or tool prose into it is a
> laundering violation of A1-4.4.
> A1-4.4 bullet 3 MUST name the tests: `TestUntrustedContextComputedNotSupplied`,
> `TestNoUntrustedTextInUnmarkedFields` (per-kind corpus: for every kind, every
> non-`*` string field is asserted to receive only platform vocabulary — a
> table-driven allowlist of source fields per target field, reviewed with the
> ingest mapping), `TestApprovalViewFlagsUntrustedContext`.

### S-07 · MUST FIX · A1-6.5 (validity window) + A1-6.4 (vs Q11 "never silent", ADR-0009 §5)
One override authorizes an unbounded stream of customer artifacts. The override
is valid "until the next `chain_verified` or `chain_break_detected`"; on a
permanently broken chain no further break event is appended (A1-6.3 forbids a
duplicate for the same `(break_seq, break_kind)`), so the window never closes.
A1-6.5's own remedy — "each released artifact MUST name the override event" —
is disclosure, not limitation, and the naming has no log-side record (see T-02).
Attack/failure: a single admin click (or one stolen admin session) releases
every subsequent report, findings export and evidence bundle for that
engagement from a chain known to be tampered with.
Replacement:
> An `integrity_override` with `scope:"export"` is **single-use** (mirroring
> Q10): it authorizes exactly one artifact release. The platform MUST bind the
> release to it atomically (uniqueness on
> `(engagement_id, override_event_id)` in the release record) and MUST compose
> an `artifact_released` event (T-02) naming `override_event_id`,
> `head_seq`, `head_hash` and the artifact's `evi_` id. A second release
> requires a second human decision → `conflict` + `integrity_failed`.

### S-08 · MUST FIX · A1-6.2 vs A1-6.4 (startup gate race)
A1-6.2 requires startup verification "before the platform serves any `/api/v1`
request for that engagement", but also that it "MUST NOT block startup of the
platform process", and A1-6.4 says internal views keep serving on a failed
chain. Nothing defines what a read returns while the walk is *in flight*
(`integrity_state:"unverified"`), and A1-6.4's flagging obligation is only
attached to `failed`.
Attack/failure: a long chain (10⁵ events) takes seconds to walk; during boot the
UI and `/api/v1` serve events with no integrity indication, and an operator
exporting in that window (or a report builder caching the read) acts on an
unverified chain. A1-6.6's `integrity_state` enum has only
`verified|failed_overridden` — an artifact built during the window has no legal
value to stamp.
Replacement:
> A1-6.2: a read served before that engagement's startup walk completes MUST
> carry `integrity_state:"unverified"` on the same carrier A4 uses for `failed`;
> no customer-facing artifact MUST be produced from an `unverified` chain
> (`integrity_failed` 409, A0-3.1). A1-6.6's `integrity_state` enum MUST stay
> `verified|failed_overridden` for exports, and the export path MUST refuse
> `unverified` outright. Test: `TestReadDuringStartupWalkIsFlaggedUnverified`.

### S-09 · SHOULD FIX · A1-4.2 (`expires_at` on the four approval kinds) (vs C6, ADR-0012 §7)
`expires_at` is required to be platform-computed on `approval_requested`, but
`approval_granted`, `approval_expired` and `approval_executed` each carry their
own `expires_at` with **no equality obligation** (A1-4.2 constrains only
`fingerprint_hash`). A defective or compromised approval component — or a
future code path that recomputes expiry from the engagement's *current*
timeout after an operator raised it via `engagement_policy_changed` — silently
extends the C6 2-hour window, and the chain shows four mutually inconsistent
expiry values with no rule violated.
Replacement:
> A1-4.2: `expires_at` on `approval_granted`, `approval_expired` and
> `approval_executed` MUST be byte-equal to the `approval_requested` value for
> the same `approval_id`. A divergence is a platform defect → `internal`, MUST
> abort the execution, and MUST be recorded as
> `action_blocked{reason:"approval_metadata_mismatch"}` (new enum value,
> additive). A timeout policy change MUST NOT affect an already-requested
> approval.

### S-10 · SHOULD FIX · A1-8.5 (SSE `Last-Event-ID`) (vs A1-8.2, C8/A12)
The stream resume is a bare decimal `seq` from a client-controlled header, with
no engagement resolution rule analogous to A1-8.2's cursor rule. `seq` is
per-engagement, so `Last-Event-ID: 412` from engagement A's stream is a valid
position in engagement B's stream.
Attack/failure: a viewer assigned to both A and B (legal) replays A's frame id
on B's stream and gets B's rows from an unrelated point — not a cross-engagement
leak of *A's* data, but a silent skip/replay inside B that defeats the
"gap-free, in order" guarantee a live monitoring UI relies on, and an easy way
for a careless operator to believe B is idle.
Replacement:
> A1-8.5: the resume position MUST be validated exactly like a cursor
> (A1-8.2): the platform MUST resolve the requested `seq` **in the stream's own
> engagement**, MUST resync from that engagement's earliest retained row when it
> cannot, and MUST send an explicit resync control frame rather than silently
> starting elsewhere. A `Last-Event-ID` that is not a decimal integer in
> `[0, head_seq]` MUST be ignored (restart from the head) and MUST NOT be
> echoed. Test: `TestSSEResumeFromForeignSeqNeverServesUnrelatedPosition`.

### S-11 · SHOULD FIX · A2-10.2 (13 steps) vs A2-3.4 / A2-4.7 / A2-8.10
The validation order never mentions the dedup collapse or `content_hash`
computation, and quarantine stamping is **step 13 — the last one**. Whether a
duplicate write is collapsed before or after step 13 is undefined.
Attack/failure: an agent re-observes a blacklisted host 500 times. If the
collapse short-circuits before step 13, only the first observation emits
A2-8.10's `graph_node_quarantined{blacklisted}` event, and the operator's
"the agent kept trying" signal — the most valuable line in the report per A2
§6.3 — is lost. If the collapse happens after, `summary_too_large`/`validation`
from steps 8–12 can reject a write that would have been a harmless no-op.
Replacement:
> A2-10.2 MUST insert "(10a) compute `content_hash` over the A2-4.6 document;
> (10b) evaluate policy and quarantine (current step 13); (10c) **then** attempt
> the dedup collapse of A2-3.4/A2-4.7" and MUST state that a collapse still
> emits the A1 `graph_node_written` (with `dedup_hit:true`, S-05) and still
> emits `graph_node_quarantined` when the recomputed quarantine state differs
> from the stored one.

### S-12 · SHOULD FIX · A1-8.8 (last bullet) vs A1-3.1 (closed, "complete on day one")
A1-8.8 mandates that evidence deletion "MUST itself be recorded as an
`evidence`-component event", but the closed 39-kind taxonomy contains no such
kind, and A1-4.1 forbids adding fields to `evidence_stored`.
Attack/failure: a careless operator or a retention job deletes the artifact that
`approval_executed.action_spec_evidence_id` (C-02) or a `finding`'s
`evidence_ids` points at; the contract requires an audit record that cannot be
written, so the implementer either skips it or invents an unregistered kind
(→ `validation`, A0-6.3). Evidence destruction becomes contractually invisible.
Replacement: add kind `evidence_removed` (platform · `evidence`) with
`evidence_id:string`, `removal_kind:enum{retention,operator,gdpr_request,
corrupted}`, `reason:string(512)*`, `referencing_event_count:int`, and require
it **before** any deletion mechanism ships. A dangling `evi_` remains a
non-break (A1-8.8), but the removal is chained.

### S-13 · SHOULD FIX · A1-7.12 (kill-path exception)
"retried until it lands" has no durability model. If the retry queue is
in-process (the natural stdlib reading, DESIGN §6), a platform crash between
the kill and the append loses the only record that a hard stop happened — and
`hard_stop_fired` is the event A1-4.2, A1-6.4 and the ADR-0005 §4 respawn
guarantee all hang off.
Replacement:
> A1-7.12: the pending kill record MUST be written to a durable outbox
> (append-only table, `REVOKE UPDATE, DELETE`, A1-7.9) **before** the kill is
> issued, and startup MUST drain the outbox into the chain. A kill whose event
> cannot be composed MUST surface in the UI as an unresolved integrity warning
> (A1-6.4's carrier), not only in `slog`.

### S-14 · NICE · A1-8.5 / A0-1.7 (`seq` publication)
Publishing dense per-engagement `seq` on every envelope and every SSE frame
lets any authorized reader measure append volume and infer activity bursts
(e.g. "the agent is running something now"). Accepted by A0-1.7's id-leak
posture, but A1-8.5's justification ("`seq` is already public") should record
the consequence explicitly so A5 can decide whether a `viewer` role sees `seq`
at all.

---

## 3. Stdlib-only feasibility (C1, ADR-0010)

### F-01 · MUST FIX · A0-2.5 / A0-2.6 / §4 `cjson.Canonical`
The mandated `json.Decoder.Token()` walk returns numbers as **`float64`** unless
`UseNumber()` is called. Consequences that break the contract's own rules:
`{"a":1e3}` and `{"a":1.0}` are indistinguishable from `{"a":1000}` and would be
**accepted** (A0-2.6 requires rejection); `{"a":-0}` becomes a signed-zero
float64 that Go re-emits as `-0`, so A0-2.6's rejection never fires and two
logical values get two digests; integers above 2⁵³ lose precision silently.
This is the single most likely way two builders produce two different chains.
Replacement:
> A0-2.5: "The stream walk MUST call `Decoder.UseNumber()` and MUST re-validate
> every number's **literal token text** against `^-?(0|[1-9][0-9]{0,15})$`
> before range-checking it with `strconv.ParseInt` against A0-2.6's
> `[-(2^53-1), 2^53-1]`. A token whose text is not already in that canonical
> form (`.`, `e`/`E`, `+`, leading zeros, `-0`) is `validation`. The
> canonicalizer MUST emit the validated literal text verbatim, never a
> re-formatted number. The walk MUST also call `Decoder.More()` after the
> top-level value to enforce A0-2.3's no-trailing-data rule, and MUST count
> nesting depth itself (`encoding/json` has no depth limit)."
> Add rejection vectors: `{"a":1E3}`, `{"a":-0.0}`, `{"a":9007199254740993}`,
> `{"a":10000000000000000000}`.

### F-02 · MUST FIX · A0-2.3 / A0-2.7 (invalid UTF-8 and lone surrogates)
`encoding/json` never errors on either case: a `\ud800` escape (unpaired
surrogate) decodes to U+FFFD, and invalid UTF-8 bytes inside a string literal
are either passed through by the decoder and then replaced with U+FFFD by the
encoder, or replaced on decode — both routes end at U+FFFD, and
`CanonicalValue`'s marshal step (`encoding/json`) always emits the substitution.
So the two rejections A0-2.3 and A0-2.7 mandate do not happen for free, and the
failure is a **collision**: `{"a":"\ud800"}` and `{"a":"\ufffd"}` canonicalize
to identical bytes and identical digests, and A1-4.9's secret scan sees U+FFFD
where the wire carried something else.
Replacement:
> A0-2.3: "Implementations MUST NOT rely on `encoding/json` for UTF-8
> validation: the decoder substitutes U+FFFD for invalid input. The
> canonicalizer MUST (a) reject the whole document when `utf8.Valid(doc)` is
> false, and (b) scan every `\u` escape during the walk and reject a surrogate
> code point (`D800`–`DFFF`) that is not followed by a complementary surrogate
> forming a valid pair. Both are `validation`."
> Add vectors: `{"a":"\ud800"}` (reject), `{"a":"\ud83d\ude00"}` (accept,
> literal 😀), `{"a":"\xff"}` (reject).

### F-03 · MUST FIX · A0-2.7 (U+2028 / U+2029)
A0-2.7 says "escape only `"`, `\`, and U+0000–U+001F … every other code point is
literal", but Go's `encoding/json` escapes U+2028 and U+2029 **unconditionally**,
even with `SetEscapeHTML(false)`. An implementer who emits strings with a
stdlib encoder and one who hand-rolls the literal rule produce different bytes
for the same document → two digests → a broken chain and broken ADR-0018
fingerprints. No vector covers it (V3 covers é/中/😀 and control chars only).
Replacement:
> A0-2.7: add "U+2028 and U+2029 are emitted **literally** (raw UTF-8), not as
> `\u2028`/`\u2029`; `encoding/json`'s encoder escapes them unconditionally and
> therefore MUST NOT be used as the canonical emitter — it MAY be used to
> produce the intermediate bytes that the canonicalizer re-parses
> (`CanonicalValue`), never the final ones. U+007F is literal."
> Add vector V7: `{"s":"a\u2028b\u2029c\u007fd"}` → canonical bytes
> `{"s":"a<e2 80 a8>b<e2 80 a9>c\u007fd"}` … i.e. the literal 3-byte sequences
> for U+2028/U+2029 and a literal `0x7f`, with its length and SHA-256 computed
> and published exactly as V1–V6 are.

### F-04 · MUST FIX · A0-5.3 (implementation note is wrong)
Go layout `2006-01-02T15:04:05.000Z07:00` **accepts** a numeric offset (`Z07:00`
matches both `Z` and `±hh:mm`), `time.Parse` **accepts** a `:60` leap second and
normalizes it into the next minute, and it accepts any 4-digit year. Three of
the five rejections A0-5.3 mandates therefore do not happen with the recipe the
clause itself recommends; a `+02:00` timestamp would be parsed, converted to
UTC and re-emitted with different bytes than the client sent — the
"reject, never normalize" philosophy (Q3) inverted.
Replacement:
> A0-5.3: "Parsing MUST be a two-stage check: (1) the byte-exact regex of
> A0-5.1 (`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`) — which by
> construction rejects offsets, lowercase `t`/`z`, space separators, other
> precisions and `:60` only if the seconds group is additionally range-checked
> `00`–`59`; (2) `time.Parse(TimeLayout, s)` for calendar validity, followed by
> an explicit year check against `[MinYear, MaxYear)`. `time.Parse` alone MUST
> NOT be used: `Z07:00` accepts numeric offsets and Go normalizes leap seconds.
> A parsed value MUST round-trip: `FormatTime(ParseTime(s)) == s` byte-exactly,
> else `validation`."
> Add the round-trip property as a contract test (`TestTimeParseRejectsNotNormalizes`,
> table-driven over the A0-5.3 rejection list).

### F-05 · SHOULD FIX · A1-5.4 (advisory-lock alternative)
`pg_advisory_xact_lock` takes `bigint` (or two `int4`), not text; "keyed on the
engagement id" is not implementable as written. The obvious bridge
(`hashtext(engagement_id)`) is an undocumented Postgres internal, is 32-bit,
and **collides across engagements** — engagement A's append would then block on
B's lock, coupling two customers' chains (C8-adjacent) and giving one
engagement's load a latency channel into another's.
Replacement: strike the advisory-lock option, or specify it exactly:
> "The store seam MUST serialize appends with `SELECT … FOR UPDATE` on the
> engagement's chain-head row (A1-5.6). If an advisory lock is used instead, its
> key MUST be derived from the chain-head row's `bigint` primary key — never
> from a hash of the engagement id text — and MUST be
> `pg_advisory_xact_lock(bigint)` so the lock dies with the transaction."

### F-06 · SHOULD FIX · A1-1.2 / A1-4.8 / A2-4.6 (nil slice and nil map → `null`)
Go marshals a nil `[]string` and a nil `map` as `null`, which A0-8.3, A1-1.2 and
A1-4.8 all forbid — and which changes the digest (`null` vs `[]` vs absent).
`Event.EvidenceRefs`, `TaskResultPayload.RevertEventIDs`,
`CleanupPlannedPayload.RevertEventIDs/NonRevertableEventIDs` and
`contentDoc.Attrs`/`EvidenceIDs`/`Addresses` are all exposed. This is the most
common way a stdlib-only canonicalizer produces a document that its own contract
rejects on re-parse.
Replacement:
> A0-2.14 (or A1-1.2): "Every slice and map field of a canonicalized type MUST
> be non-nil before marshaling; the constructors (`events.NewEvent`,
> `graph.NewNode`, `cjson.CanonicalValue`) MUST initialize them to empty. A
> canonical document containing `null` is a platform defect → `internal`, and
> `cjson.Canonical` MUST reject `null` in input (A0-2.3's rejection list)."
> Test: `TestNilCollectionsNeverCanonicalizeToNull` (reflection over every
> canonicalized type: zero-valued instance marshals with no `null` token).

### F-07 · SHOULD FIX · A2-6.1 / A2-6.2 / §4 `Attrs`, `AttrValue` (undefined wire shape inside a digest)
`AttrValue` is a 4-field struct with **no `json` tags** (A0-8.1 violation) and
no declared JSON representation. A2 §4.1's examples settle it by showing the
flat form (`"attrs": {"cvss_v3_x10": 88, "relay_tool": "ntlmrelayx"}`), but the
normative type sketch — which `contracts/README.md` makes "the source of truth
for field names and JSON shapes until the implementing package merges" — says
the opposite: default marshaling of that struct yields
`{"k":{"Type":"string","Str":"v","Num":0,"Bool":false}}`, depth 2, which
A2-6.1's "depth exactly one" then contradicts. `attrs` is part of
`content_hash` (A2-4.6), so the ambiguity is a fingerprint fork: the same node
written by two builds dedups against nothing and A2-4.8 verification fails.
Replacement:
> A2-6.1: "The wire and canonical form of `attrs` is
> `{"<key>":<JSON string|integer|boolean>}` — depth exactly one, no wrapper
> object. `AttrValue` MUST carry explicit `json:"-"` on its Go fields and a
> hand-written `MarshalJSON`/`UnmarshalJSON`; the type discriminator exists only
> in Go. A2 MUST publish a normative `content_hash` vector that includes a
> three-key `attrs` (one per type), exactly as A1 §4.3 does for the chain."

### F-08 · NICE · A0-4.4 `Cursor.K int64` vs A0-4.3
A0-4.3 permits an ordering key built from "append sequence **or id**", but the
`Cursor` type only carries an `int64`. A collection ordered by id text cannot
use the declared cursor shape.
Replacement: A0-4.3 add "every paginated collection MUST have an integer primary
ordering key so A0-4.4's `{k,id}` cursor is expressible; a text key is only ever
the tie-breaker carried in `id`."

### F-09 · NICE · A0-2.11 (depth/size limits)
Neither limit is provided by `encoding/json`; both are hand-rolled inside the
`Token()` walk, which AGENTS.md puts on the high-review list ("all
untrusted-input parsing"). Say so in A0-2.11 and require the review gate
explicitly, plus a 32-deep accept / 33-deep reject vector pair and a 1 MiB
boundary pair (the current rejection list has only "40-deep nesting").

---

## 4. Determinism traps

### D-01 · MUST FIX · A2-4.6 / A2-4.8 (no normative vector)
A0 has V1–V6, A1 has §4.3, A2 has **none** — yet `content_hash` decides node
identity (A2-4.7 dedup), revision legality (A2-4.3) and A2-4.8 verification.
Failure: the 20-key `contentDoc` is transcribed with one field missing or one
zero value differing (e.g. `status:""` vs absent, `port:0` vs omitted), and every
stored fingerprint silently forks between builds; A2-4.9 downgrades the symptom
to `internal`, so nobody sees it.
Replacement: add "### 4.2 Normative content fingerprint vector" to A2 with two
rows — a `finding` node (all 20 keys, `attrs` with one string/int/bool,
`evidence_ids` with two entries) and a `service` node (inapplicable keys at
their zero values) — each giving the exact canonical bytes, byte length and
SHA-256, and declare it normative like A0-2.17 and A1 §4.3.

### D-02 · MUST FIX · A0-2.14 vs A2-4.6 / §4 `Node` (two presence rules for one node)
A2-4.6 says the digest document uses fixed keys with zero values while "the API
representation of a node keeps A0-8.3 absence semantics" — and the Go sketch has
one `Node` type with `omitempty` tags plus a separate unexported `contentDoc`.
Nothing forbids computing the fingerprint from `Node`, and nothing tests that
`contentDoc` always emits all 20 keys.
Failure: a builder canonicalizes `Node` (it is exported, it has all the fields);
a node with `port` unset and one with `port:0` then hash identically or
differently depending on the tag, and a later added field changes every
historical fingerprint (A2-4.8's stated fear, realized).
Replacement:
> A2-4.6: "The fingerprint MUST be computed from `contentDoc` only; `Node` MUST
> NOT be passed to `cjson` for fingerprinting. A reflection test
> (`TestContentDocFixedKeySet`) MUST assert that the canonical bytes of a
> zero-valued `contentDoc` contain exactly the 20 keys of A2-4.6 in byte order,
> and `TestContentHashVector` MUST assert the §4.2 vector byte-exactly."

### D-03 · MUST FIX · A1-5.7 `Served()` (decode → re-canonicalize)
Producing the served envelope requires decoding the stored preimage and
re-canonicalizing with three extra keys. That round trip must preserve number
**literal text** (F-01), reject duplicates (A0-2.5) and preserve string
escaping (F-03) — and the §4.3 vector's largest integer is `48210`, so nothing
in the normative material exercises a 16-digit integer, a `\u2028`, or a
non-BMP key on this path. `TestServedBytesReproducible` compares two runs of
one implementation, not two implementations.
Replacement:
> A1-5.7: "`Served` MUST decode with `UseNumber()` and MUST re-emit every
> number's literal text verbatim; it MUST reject a preimage that fails A0-2
> (a stored preimage that does not re-parse is `preimage_mismatch`, A1-6.3)."
> A1 §4.3: extend the vector with a fourth event carrying a 16-digit integer
> (`size_bytes`), a `\u2028` inside a `reason`, and a payload key that is
> non-BMP, so preimage → served → preimage is locked against a second
> implementation. Add `TestServedRoundTripIsIdentity`.

### D-04 · SHOULD FIX · A1-7.10 step order (12 `untrusted` before 13 `evidence_refs`)
`untrusted` is computed at step 12 but `evidence_refs` is derived at step 13 and
is also inside the digest. Today neither depends on the other, so the order is
harmless — but the contract states the order as normative and any later field
derived at step 13 that is `*`-marked would make the flag stale and the digest
wrong.
Replacement: A1-7.10 add "(12) `untrusted` computation and (13) `evidence_refs`
derivation are order-independent because neither is a `*`-marked field; a future
amendment that makes a derived field untrusted MUST move step 12 last."

### D-05 · SHOULD FIX · A2-6.4 / A0-7.3 (`attrs` cap measurement)
"4096 B for the **serialized** `attrs` object" plus A0-7.3's "with the same
encoder settings used for the response" is undefined for a stored node: the
response representation (absent-when-empty, A0-8.3) and the canonical
representation (always present, A2-4.6) differ in bytes, so ingest and a later
re-check disagree and a node accepted at write time can fail validation on
read-back.
Replacement: A2-7.1/A2-6.4: "document-class caps in A2 are measured on the
**canonical JSON** (A0-2) bytes of the field's own object, computed once at
ingest and stored alongside `content_hash`; the API representation is never a
measurement input."

### D-06 · SHOULD FIX · A0-2.11 vs A1-1.5 / A1-4.8 / A2-6.1 (depth counting)
Four depth statements (32 / ≤3 / ≤3 / exactly 1) and no definition of what
counts as level 1. An implementation that counts the top-level object as 0 and
one that counts it as 1 disagree at the boundary; the rejection list only has
"40-deep nesting", which passes under both.
Replacement: A0-2.11: "Depth is counted with the top-level object as level 1;
an array element adds one level; a document at depth 32 MUST be accepted and at
33 MUST be rejected (`validation`). Vectors: a 32-deep accept and a 33-deep
reject."

### D-07 · NICE · A0-4.4 prose vs A0-4.4 example
The clause writes the cursor payload as `{"k":…,"id":…}`; canonical byte order
(A0-2.4) and the §4 example are `{"id":…,"k":…}`. A builder transcribing the
prose into a hand-written encoder produces a cursor that A0-4.8 rejects as
non-canonical. Fix the prose to the sorted order.

### D-08 · NICE · A0-1.9 (`COLLATE "C"`)
The clause does not say whether the collation is a column/DDL property or a
per-query `ORDER BY … COLLATE "C"`. The per-query form silently disables index
use on every paginated read (A0-4.6's `limit+1` scan becomes a sort), and two
builds can disagree on ordering for text keys while both "comply". State that
the collation is declared on the column and that queries MUST NOT re-specify it
(dependency on backlog session 6).

---

## 5. Taxonomy completeness (closed 39 kinds)

Coverage of SPEC §5 steps is claimed by A1-3.7 and is largely correct; the gaps
below are what the claim hides. Every addition listed here is additive under
A0-6.5/A1-3.5 and costs nothing before Freeze.

### T-01 · MUST FIX · A1-3.7 step 1 (vs SPEC §5.1, ADR-0012 §1)
Step 1 ("create an engagement: client, scopes, blacklist, assigned operators,
model-role config, rules of engagement") has no `engagement_created` kind.
`chain_genesis`'s payload is `chain_spec` only, so the chain never records that
the engagement exists, who the client is, what ROE was agreed, or who may
approve. There is likewise no `engagement_closed`. An auditor reading one
engagement chain cannot state the engagement's own identity or authority model
from the log — the report has to be trusted instead of verified.
Add: `engagement_created` (user · `scope`) with `client_ref:string(128)`,
`roe_evidence_id:string`, `policy_evidence_id:string`, `operator_count:int`;
`engagement_closed` (user · `runtime`) with `close_reason:enum{completed,
cancelled,abandoned}`, `report_evidence_id:string`. `chain_genesis` stays the
integrity anchor and MUST remain at `seq` 0, with `engagement_created` at
`seq` 1.

### T-02 · MUST FIX · A1-3.7 step 9 (vs ADR-0009 §5, Q7, A1-6.5/6.6)
No kind records the **release of a customer artifact**. A1-6.6 binds the
integrity block to the artifact, and A1-6.5 requires each released artifact to
name its override — but the log has no `artifact_released`, so neither the
release nor the override that authorized it is chained. Step 9 is currently
mapped to `chain_verified{pre_export}` + `integrity_override` +
`evidence_stored{report_artifact}`, none of which records *that a release
happened*, to whom, or from which head.
Add: `artifact_released` (user or platform · `integrity`) with
`artifact_kind:enum{report_html,report_pdf,findings_json,evidence_bundle,
verification_bundle}`, `artifact_evidence_id:string`, `head_seq:int`,
`head_hash:64hex`, `integrity_state:enum{verified,failed_overridden}`,
`override_event_id:string`, `recipient_ref:string(128)`. This is also the
enforcement point for S-07's single-use override.

### T-03 · MUST FIX · A1-3.3 approval/spawn payloads vs A7 (ADR-0018 §1–§2, Q14)
A1-4.1 forbids reshaping a kind's payload after Freeze, and A1-3.6 hands A7
"fingerprint content" — but the fields A7 *must* quote in the log are not
reserved, so A7 cannot be written without breaking A1. Missing at minimum: the
action-spec reference (C-02), the graph node the action targets (S-04), and a
stable argv digest distinct from the whole-spec fingerprint.
Reserve now on `approval_requested` **and** `approval_executed`:
`action_spec_evidence_id:string`, `target_graph_node_id:string` (`gn_`, `""`
when the target is not a graph node), `argv_hash:64hex`; and on
`spawn_requested`/`task_spawned`: `target_graph_node_id:string`. Declare in
A1-3.6 that A7 owns their *content*, A1 their *presence*.

### T-04 · SHOULD FIX · A1-3.3 `graph_node_written` / `graph_edge_written` vs A2-3.4 / A2-4.7
Neither kind can express a **dedup collapse**, which is the normal outcome of a
re-observation (and of every offline-node replay, ADR-0013). The log therefore
cannot distinguish "new evidence recorded" from "duplicate ignored", and the
second observation's provenance is invisible (S-05).
Add `dedup_hit:bool` and `content_hash:64hex` to `graph_node_written`; add
`dedup_hit:bool` to `graph_edge_written`.

### T-05 · SHOULD FIX · A1-3.1 vs four occurrences the contracts themselves mandate
The closed list cannot express: evidence deletion (A1-8.8 → S-12
`evidence_removed`), credential/token revocation and node pairing (Q8/Q9 → C-09),
engagement operator assignment (ADR-0012 §1 → C-08), and a clock anomaly
(A1-5.4's clamp firing — see below). A1-3.1's "closed and complete on day one"
is therefore false as written.
Add a fifth: `clock_anomaly` (platform · `runtime`) with
`direction:enum{backward,forward}`, `step_ms:int`, `clamped_event_count:int` —
required because A1-5.4's clamp is unbounded (next finding).

### T-06 · SHOULD FIX · A1-3.3 `quarantine_kind` vs A2-8.1 `quarantine_reason`
Two vocabularies for one fact with no mapping and no owner: A1 has
`out_of_scope_discovery|blacklist_match|operator_quarantine|operator_release`;
A2 stores `out_of_scope|blacklisted`. A1-3.6 declares that A2 owns
`node_kind`/`edge_kind` but says nothing about quarantine, so a builder will
copy strings across and drift.
Add to A1-3.6: "A2 owns the quarantine **state** vocabulary (`quarantine_reason`);
A1 owns the **occurrence** vocabulary (`quarantine_kind`). The mapping is
`out_of_scope_discovery→out_of_scope`, `blacklist_match→blacklisted`,
`operator_quarantine→(the reason already in force)`, `operator_release→
quarantined:false`. The stored reason MUST be derived by the platform from this
mapping, never copied from an event string." (And delete `operator_release`
per C-06.)

### T-07 · SHOULD FIX · A1-5.4's `recorded_at` clamp (abuse analysis, as requested)
The clamp cannot be driven by a client (`recorded_at` is platform-stamped,
A1-2.3, and `occurred_claimed_at` orders nothing, A1-8.1), so it is **not** an
ordering-forgery vector. It is a *timeline-integrity* vector: `max(clock_now,
prev_recorded_at)` is unbounded forward, so one bad clock read (an NTP step, a
restored VM snapshot, a careless operator setting the host clock) stamps
**every subsequent event of that engagement** with a future `recorded_at` until
wall time catches up — potentially days. `occurred_at` is *not* clamped, so the
two diverge inside one event, and the UI/report shows `occurred_at` while the
audit authority is the clamped one. Nothing records that the clamp fired.
Replacement: bound it and log it.
> A1-5.4: "The forward clamp MUST be bounded to 1000 ms. Beyond that the
> platform MUST use the true clock reading, MUST compose a `clock_anomaly`
> event (T-05) and MUST log at error level; `recorded_at` MAY then be
> non-monotone, and A1-8.1's rule that only `seq` orders anything is the
> compensating control. `occurred_at` MUST be clamped by the same rule so the
> two never diverge by more than the bound."

### T-08 · NICE · A1-3.7 step 2/8 (run lifecycle)
`run_started` has no counterpart for suspension, and `run_ended` is
platform-composed with `end_reason:cancelled` but no field naming who cancelled
(the `actor` is `platform`, so a user-initiated cancel is attributed to the
platform). Add `cancelled_by:string` (`usr_`) to `run_ended`, or make
`run_ended` user-composable for the cancel case.

---

## 6. Enforceability (AGENTS.md: every safety rule needs a positive **and** a negative test)

Named tests that exist today (A0: 0 · A1: 14 · A2: 8) cover the chain vectors,
the tamper matrix, cross-engagement reads, secret-free serialization and the
seam shape. The safety path is the gap: **no MUST on the write-authorization,
approval, kill or quarantine path has a named test at all.**

### E-01 · MUST FIX · A1-2.3, A1-2.6, A1-2.7, A1-3.4, A1-7.4 (write-path bypass)
The single most important negative test in A1 is unnamed: a machine principal
appending a platform-composed kind, or supplying any platform-stamped envelope
field. A1-7.4 says "a worker that could append `approval_granted` would own the
safety model" and then names no test.
Add to the shared suite (names are the contract):
`TestWorkerCannotAppendNonCKind` (each of the 36 non-C kinds → `forbidden` +
`action_blocked{append_not_permitted}`), `TestClientCannotSupplyEnvelopeFields`
(table over all 12 fields of A1-2.3 → `validation` naming the field),
`TestActorCannotBeForged` (a body-supplied `actor.type:"platform"` is rejected
and never influences the stored row), `TestOrchestratorCannotClaimCommandExecuted`
(A1-7.4's C-kind actor rule), `TestMachinePrincipalCannotReachIntegrityKinds`,
`TestUntrustedFlagCannotBeSupplied`.

### E-02 · MUST FIX · A2-8.3, A2-8.5, A2-8.8 (quarantine is the C5 control)
No test proves a quarantined node is not actionable, that `blacklisted` survives
a scope change and an approval, or that `report_excluded` is operator-only.
Add: `TestQuarantinedNodeNotTargetableWithValidApproval` (grant a real approval
for a quarantined node's target → refused before execution, ADR-0018 §2
re-validation), `TestBlacklistedNodeSurvivesScopeWidening`,
`TestQuarantinedNodeAbsentFromPlanningView`, `TestRetractedEdgeAbsentFromPlanningView`,
`TestMachinePrincipalCannotSetReportExcluded`, `TestQuarantineFlagCannotBeSuppliedOnWrite`.

### E-03 · MUST FIX · A1-6.4, A1-6.5, A1-7.12 (integrity + kill path)
Add: `TestExportBlockedOnFailedChain` (each of the four artifact classes →
`integrity_failed` 409 whose message names `break_seq`/`break_kind`),
`TestInternalViewOverrideDoesNotAuthorizeExport`, `TestOverrideDiesAtNextBreak`,
`TestSingleUseOverride` (S-07), `TestExportFromUnverifiedChainRefused` (S-08),
`TestHardStopProceedsWhenEventStoreUnavailable` (ADR-0005 §4: containers are
killed, respawn stays blocked, the event lands from the outbox after recovery —
S-13), `TestKillPathDoesNotBlockOnAppendLatency`.

### E-04 · MUST FIX · A1-5.3, A1-6.2, A1-7.11 (genesis and the append gate)
Add: `TestSecondGenesisIsRejected`, `TestAppendRefusedWithoutValidGenesis`,
`TestAppendRefusedBeforeStartupVerificationCompletes`,
`TestFailedAppendConsumesNoSeq` (rollback leaves no gap — A1-5.4),
`TestNoGlobalSequenceSharedAcrossEngagements` (a rolled-back append in A does
not create a gap in B).

### E-05 · SHOULD FIX · remaining MUSTs with no named test
A1-4.2 (`expires_at` platform-computed; fingerprint equality across the four
approval kinds) · A1-4.7 (array sort/dedup at composition; count caps →
`summary_too_large`) · A1-4.9 (`TestSecretScanNamesFieldNotValue`) ·
A1-7.6 (`TestIdempotencyKeyReuseWithDifferentPayloadIsConflict`,
`TestRetryCannotShiftClaimedTime`, `TestDedupHitWritesNoEvent`) ·
A1-8.4 (`TestWorkerAndNodeHaveNoReadScope`, `TestSSENotReachableByMachinePrincipal`) ·
A1-8.5 (`TestSSEEmitsAfterCommitInSeqOrder`) · A0-2.15 (`TestGatingComparisonsAreConstantTime`
— assert `subtle.ConstantTimeCompare` is the only comparison on the four gating
paths) · A0-4.5/A0-4.6 (`TestLimitAboveMaxRejectedNotClamped`,
`TestExactFullPageHasNoNextCursor`) · A0-7.5/A0-7.6 (`TestNoSilentTruncationInHashedRecords`) ·
A2-3.2 (`TestEdgeEndpointMatrixRejects`), A2-3.7, A2-4.3 (`TestSupersedeCycleRejected`,
`TestSupersedeNonCurrentRejected`), A2-6.3 (`TestReservedAttrsKeyRejected`),
A2-10.6 (see E-07).

### E-06 · SHOULD FIX · contracts/README.md "Merge gate" (vs AGENTS.md, DESIGN §8)
The merge gate lists six *categories* of test and never states AGENTS.md's rule
that every safety rule needs a positive **and** a negative test. Result: 27
MUSTs across A0–A2 have no test id, and a work package can satisfy the gate
while proving nothing about the bypass.
Replacement: add a seventh bullet and a per-clause obligation:
> "**Safety-path pairing** — every MUST / MUST NOT in A0–A8 that guards
> authorization, integrity, approval, quarantine, egress or the kill path
> carries a test id in the clause itself (`Tests: <Name>, <Name>`); one id MUST
> be a positive test of the rule and one MUST be a negative test of the bypass
> attempt named in the clause. A contract reaches `Frozen` only when every such
> clause has both. The suite fails if a safety-path clause has no test id."

### E-07 · NICE · A2-10.6 ("a `go vet`-visible exported-field audit fails the build")
`go vet` cannot express this; the sentence names a mechanism that does not
exist, so the constructor-only invariant is unenforced.
Replacement: "…enforced by `TestNodeFieldsUnexported` (reflection over
`graph.Node`/`graph.Edge`: every field except the four mutable flags of A2-1.3
and A2-3.9 is unexported) plus a `go build` check that no package outside
`internal/graph` constructs one (DESIGN §1 layering, review gate)."

### E-08 · NICE · A1 §4.2 (no JSON example for the approval path)
§4.2 gives examples for the append request, a served event, the break/override
pair, a paginated read, a single-event read, an SSE frame, the rejection
envelopes, the export integrity block and the chain-head state — but **none**
for `approval_requested` / `approval_granted` / `approval_executed`, the
highest-risk kinds in the document (ADR-0018, C6, Q10) and the ones whose
`fingerprint_hash`/`expires_at` equality obligations (A1-4.2, S-09) a builder
must get byte-exact. Add one three-event approval example (request → grant →
execution) with identical `fingerprint_hash` and `expires_at` across all three,
plus the `action_blocked{fingerprint_mismatch}` rejection that a substitution
attempt produces.

---

## Cross-cutting note (not a finding id)

Three documents each declare a freeze blocker on the same missing decision —
the `usr_` principal prefix (C-10) — and two each declare the A0-7.1 const-block
split (A1 AM-2, A2 AM-2) and the A0-7.10 cap contradiction (A0 §6.14, A2-12.3).
A0 should absorb all three before any of A0/A1/A2 is marked `Frozen`; otherwise
the first implementer will invent them locally, and A0-2.1's "one
implementation" rule is the one thing these contracts cannot survive losing.

## NOT COVERED (budget)

- A1 §4.2 JSON examples (lines 1849–2095) and A2 §4.1 examples (lines 894–1073):
  read only by targeted grep for `null`/float/`omitempty`/approval fields. No
  line-by-line cross-check of every example against its clause (e.g. the
  `chain_break_detected` / `integrity_override` example pair, the error-envelope
  examples, the A2 `finding` node example) — a mismatch between an example and a
  clause would not have been caught.
- A1-4.2's per-kind obligation table: checked for the integrity, approval,
  kill-path, graph and cleanup kinds; the remaining ~20 rows (e.g.
  `evidence_stored`, `llm_call`, `notification_sent`, `cleanup_*`) were read but
  not each traced to a test or an ADR clause.
- `adr/ADR-0006`, `0007`, `0008`, `0011`, `0014`, `0015` were skimmed only where
  A0/A1/A2 cite them; no independent conflict hunt against those six.
- `WORKFLOW.md`, `next_steps.md`, `docs/adversarial-review-2026-09-03.md`,
  `docs/enterprise-readiness.md`, `sessions/2026-09-07-contract-freeze.md`,
  `sessions/2026-09-11-contract-freeze-a1.md` were not read; findings A1/A9/A11/
  A12/A14/A15 are referenced only as the contracts characterize them.
- No verification of the A0-2.17 V1–V6 digests or of the A2/A1 example digests
  (no Go toolchain in this environment; `go` is not installed — the parent
  already reproduced A1 §4.3).
- Persistence/DDL feasibility (backlog session 6) beyond the two clauses that
  touch it (A1-7.9 `REVOKE`, A0-1.9 collation).
- A3–A8 seams: only the obligations A0/A1/A2 explicitly hand to them were
  checked for satisfiability.
