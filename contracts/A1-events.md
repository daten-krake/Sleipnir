# A1 — Event log

## 1. Header

| | |
|---|---|
| **Contract id** | A1 |
| **Status** | `Draft` (`contracts/README.md` lifecycle: Draft → Frozen → Implemented) |
| **Owner** | architect |
| **Gates** | A2 (graph provenance fields) · A3 (stage views read the log) · A4 (append/read/SSE endpoints) · A5 (scopes per principal) · A7 (fingerprint + spawn payload fields) · `internal/event` (store seam) · report builder · every component that composes an event |
| **Implements** | ADR-0009 §1–§4 · ADR-0012 §1/§2/§6/§7 · ADR-0016 §1/§2/§4 · ADR-0017 §2–§3 · ADR-0018 §1–§4 · ADR-0019 §3/§5 · ADR-0020 §2–§5 · ADR-0005 · ADR-0011 · ADR-0013 · SPEC §5 steps 1–10, §6, §7, §8, C5/C8/C9/C11 · Q6 (+ worker addendum), Q10, Q11, Q12, Q13, Q14 · adversarial A1/A9/A11/A12 |
| **Depends on** | A0 in full — ids (A0-1), canonical JSON (A0-2), error kinds (A0-3), pagination (A0-4), time (A0-5), versioning (A0-6), size caps (A0-7), field conventions (A0-8). A1 restates no A0 rule; it cites it. |
| **External refs** | FIPS 180-4 (SHA-256) · RFC 8785 via A0-2.2 · RFC 3339 via A0-5.1 |
| **Authority** | ADR > SPEC > DESIGN > contract. A clause here that contradicts an Accepted ADR is a defect in this document. |

## 2. Scope

**Fixed here:** the one event envelope (A1-1); typed actors and the platform's
identity stamping (A1-2, Q6); the **closed** event kind list and every payload
schema (A1-3); payload key-set, cap, untrusted-content and secret rules
(A1-4); the per-engagement hash chain — digest definition, exclusion list,
genesis, `seq` assignment, stored preimage bytes (A1-5, Q11); verification
triggers, failure behaviour, operator override, and the tamper matrix a
contract test must prove (A1-6); the append-only write path and the
`events:append` deduplication key (A1-7, A0-3.11); read ordering, filtering
and the guarantees an SSE consumer may rely on (A1-8, A0-4.3).

**Not fixed here:** endpoint paths, methods, request/response wrappers, SSE
framing, rate limits, per-endpoint body limits (A4) · token formats, scope
names, the Q6 exclusion list mechanics (A5) · node/edge kinds, `Attrs`,
quarantine semantics on the graph side, summary caps (A2 — A1 fixes only the
*event-side* field names A2 must reference, A1-3.6) · action fingerprint
content and the spawn-request schema (A7 — A1 fixes only the fields an event
quotes from them) · DDL, indexes, partitioning, retention, archival, and any
checkpoint/re-genesis mechanism (backlog session 6; see A1-7.9) · report and
export document composition (report session; A1 fixes only the integrity
metadata an export must carry, A1-6.6) · webhook payload schema (A4/ADR-0012
§3) · masking-scanner design (security hardening session; A1 fixes only the
`redacted` marker that must exist from day one, A1-4.6) · user/session audit
events (gap, §6.9).

**Design rule for this document (DESIGN §2):** the log is the audit spine,
not a general-purpose datastore. A new kind is added only when the occurrence
must be independently *filterable* (A1-8.3) and independently reportable;
everything else is a payload field or a `reason` enum value inside an existing
kind (A1-3.5).

## 3. Normative clauses

### A1-1 · Event envelope

- **A1-1.1** There is **one** envelope shape for every event, without
  exception. Its JSON key set is closed at 17 top-level keys. A kind adds no
  envelope field and removes none; kind-specific data lives only in `payload`
  (A1-4). _One shape is what makes the chain, the store, the paginated read
  path and the report builder single-code-path; per-kind envelopes would need
  one digest definition per kind._

  | # | Key | Type | Set by | In digest | Notes |
  |---|---|---|---|---|---|
  | 1 | `event_id` | string `evt_` | platform | yes | A0-1.2; unique per engagement chain |
  | 2 | `engagement_id` | string `eng_` | platform | yes | chain scope (A1-5.1); from token binding, never from the body (A1-2.5) |
  | 3 | `run_id` | string `run_` | platform | yes | `""` when the event is not run-scoped (e.g. `chain_genesis`, `scope_changed`) |
  | 4 | `job_id` | string `job_` | platform | yes | orchestrator container (A0-1.2); `""` when not applicable |
  | 5 | `task_id` | string `task_` | platform | yes | worker container; `""` when not applicable |
  | 6 | `node_id` | string `slp_node_` | platform | yes | remote agent node (Q9); `""` for platform-local execution. A **graph** node is never `node_id` (A0-3.6) |
  | 7 | `occurred_at` | timestamp | platform | yes | A0-5: platform-stamped time of the occurrence (A1-1.4) |
  | 8 | `occurred_claimed_at` | timestamp | client | yes | **untrusted** (A0-5.7, A0-8.2); `""` when the appender supplied none |
  | 9 | `recorded_at` | timestamp | platform | yes | authoritative ingest time (A0-5.6); the only client-time-vs-platform-time comparison an auditor needs |
  | 10 | `actor` | object | platform | yes | A1-2; three keys, fixed |
  | 11 | `kind` | enum string | platform-validated | yes | closed list (A1-3); unknown on write → `validation` (A0-6.3) |
  | 12 | `payload` | object | see A1-3 | yes | one flat type per kind (A1-4.1, A1-4.8) |
  | 13 | `evidence_refs` | array of string `evi_` | platform | yes | ADR-0009 §2; `[]` when none; ascending byte order (A1-4.7) |
  | 14 | `untrusted` | bool | platform | yes | A1-4.4: payload carries tool/model/target/client prose |
  | 15 | `seq` | integer ≥ 0 | platform | **no** | per-engagement position (A1-5.4); excluded per A0-2.12 |
  | 16 | `prev_hash` | 64-hex | platform | **no** | `hash` of `seq-1`, or the zero constant at genesis (A1-5.3/5.5) |
  | 17 | `hash` | 64-hex | platform | **no** | SHA-256 over the canonical preimage (A1-5.2) |

- **A1-1.2** The envelope is a canonicalized type: every one of the 17 keys
  appears on **every** event with its zero value (`""`, `0`, `false`, `[]`)
  when unset — no `omitempty`, no absence, no `null` (A0-2.14, A0-8.3). This
  overrides the general "unset optional field is absent" rule of A0-8.3 for
  events, because events are hashed. _`{"job_id":""}` and `{"job_id":…absent}`
  have different canonical bytes and therefore different digests; a producer
  and a verifier that disagree on presence would report every historical event
  as tampered._
- **A1-1.3** `seq`, `prev_hash` and `hash` are **top-level** envelope keys,
  not fields of a nested object, because A0-2.12 exclusion lists accept plain
  top-level names only (no path syntax, no wildcards). The Go `ChainBlock`
  type (§4) is a source-code grouping that is embedded so its three fields
  serialize at the envelope's top level; a nested `"chain":{…}` object would
  be unexcludable and is forbidden. _This is the single structural decision
  the exclusion list forces._
- **A1-1.4** Time semantics (A0-5.6, A0-5.7). `recorded_at` is the platform
  clock at commit. `occurred_at` is the platform's own stamp of when the
  occurrence happened: for a platform-composed event it is the instant the
  platform performed or observed the action; for a machine-appended event the
  platform sets it to the ingest instant and MUST NOT adopt the client's
  claim. A client's own time is accepted **only** into
  `occurred_claimed_at`, is stored next to `recorded_at` so divergence is
  visible (offline node buffering — ADR-0013, Q9), and MUST NOT drive
  ordering (A1-8.1), chain verification (A1-6), approval expiry (ADR-0012 §7)
  or single-use checks (Q10). All three fields are inside the digest: a
  claim cannot be edited afterwards without breaking the event's `hash`.
- **A1-1.5** `payload` is typed per kind and **flat**: a JSON object whose
  values are strings, integers, booleans, or arrays of strings/integers — no
  nested objects, no arrays of objects (A1-4.8). Depth of an event document is
  therefore ≤ 3, far inside A0-2.11's 32.
- **A1-1.6** An event MUST NOT contain a field that the platform intends to
  change after commit. Post-hoc annotation (verification status, report
  inclusion, quarantine, re-validation outcome) is a **new event** referencing
  the annotated one (A0-6.6, A1-7.2). _A mutable field inside a hashed
  document is a self-invalidating digest._
- **A1-1.7** `evidence_refs` and every `*_evidence_id` payload field carry
  `evi_` ids (A0-1.2, ADR-0009 §2). The event log stores **references**, never
  artifact bytes and never raw tool/model output (A1-4.3).
- **A1-1.8** The served representation of an event is the full 17-key envelope
  in canonical form (A0-2, keys sorted, no whitespace), so read bytes are
  reproducible across platforms and releases. Verification MUST nevertheless
  recompute from the stored preimage bytes (A0-2.16, A1-5.7) and MUST NOT use
  a served or re-marshaled body.

### A1-2 · Actor, identity and provenance stamping

- **A1-2.1** `actor` has exactly three keys, all always present (A0-2.14):
  `type` (closed enum: `user` | `orchestrator` | `worker` | `platform` |
  `node`), `principal_id` (string), `component` (string, `""` unless
  `type="platform"`).
- **A1-2.2** `principal_id` is an A0-1.2 identifier selected by `type`, and
  MUST be validated per A0-1.5 at composition:

  | `type` | `principal_id` | `component` |
  |---|---|---|
  | `user` | user principal id — A0-1.2 declares **no** user prefix; A1 requires `usr_` + A0-1.1 body (**A0 amendment request, §6.2**). A username or e-mail MUST NOT be used: both are mutable and both are personal data in a customer export | `""` |
  | `orchestrator` | the `job_` id of the orchestrator container | `""` |
  | `worker` | the `task_` id of the worker container | `""` |
  | `node` | the `slp_node_` id (Q9) | `""` |
  | `platform` | `""` | the platform subsystem that composed the event, from the closed list of A1-2.4 |

- **A1-2.3** Only the platform composes events. `actor` is stamped from the
  authenticated principal and the request context; a client MUST NOT supply
  `actor`, `event_id`, `engagement_id`, `run_id`, `job_id`, `task_id`,
  `node_id`, `seq`, `prev_hash`, `hash`, `recorded_at` or `occurred_at`. Any
  of them in an append body is an unknown field on a write → `validation`
  naming the field (A0-6.2). _Q3/Q6: never trust client discipline; a
  self-declared actor is the cheapest possible audit forgery._
- **A1-2.4** `component` is a closed enum (A0-8.5) naming the platform
  subsystem, so an audit reader can attribute a platform-composed event
  without a code search (ADR-0019 §3 correlation, C9): `event_store` ·
  `scope` · `approval` · `spawn_broker` · `llm_gateway` · `graph` · `evidence`
  · `cleanup` · `integrity` · `notify` · `runtime` · `api`. An unknown value
  is a platform defect → `validation` at composition, surfaced as `internal`
  to the caller (A0-3.1) and logged (A0-3.8).
- **A1-2.5** `engagement_id` is derived from the per-request engagement
  binding of the caller's token (Q6, third enforcement layer). A body-supplied
  engagement id is rejected (A1-2.3); a token bound to engagement A used
  against engagement B yields `notfound` (404), **never** `forbidden`
  (A0-3.9, adversarial A12) — the existence of another engagement's log MUST
  NOT be disclosed.
- **A1-2.6** Machine principals append only within their bound engagement and
  run: an orchestrator (`job_`) or worker (`task_`) append whose run binding
  does not match the token's is `forbidden` (403, principal-level, A0-3.1).
  A worker principal is **report-only** (Q6 addendum): `events:append`,
  `evidence:upload`, `task:result`, and no read scope (A1-8.4).
- **A1-2.7** `actor.type="platform"` MUST NOT be selectable or reachable
  through `/api/v1` by any machine principal; platform-composed kinds
  (A1-3.4) are unreachable for them by construction. A user principal can
  trigger a platform-composed event only through the endpoint that performs
  the action (hard stop, override, quarantine) — never by appending.
- **A1-2.8** Provenance for A2 (ADR-0016 §1) is the pair
  (`event_id`, `actor`) of the event that caused a graph write, carried into
  the graph row by the platform and echoed back in `graph_node_written` /
  `graph_edge_written` payloads as `source_event_id` (A1-3.6). Provenance is
  never client-supplied.

### A1-3 · Closed event taxonomy

- **A1-3.1** The event kind list is **closed** and complete on day one (37
  kinds, below). An unknown or malformed `kind` on a write is hard-rejected
  with `validation` (A0-6.3, Q3); on a read a client MUST preserve the raw
  string and skip what it cannot handle (A0-6.3). Kind spelling follows
  A0-8.5: `^[a-z][a-z0-9_]{0,31}$`, `<subject>_<past_participle>`.
- **A1-3.2** Notation in the payload column: `name:type`; `(N)` = capped at N
  UTF-8 bytes of the decoded value (A0-7.3), mechanism per A1-4.5; `*` =
  untrusted-content field (A1-4.4); `64hex` = A0-2.15 digest; `enum{…}` =
  closed list (A0-8.5); `array[string]` = ordered string array (A1-4.7).
  Every listed field is **always present** with its zero value (A0-2.14);
  a field not listed MUST NOT appear (`validation`, A0-6.2).
- **A1-3.3** "Composed by" names the principal whose action the event records
  (= `actor.type`, A1-2.1) and the platform subsystem that writes the row
  (= `actor.component` when `actor.type="platform"`). A platform subsystem
  recording a client's request stamps the **client** as actor: `actor` answers
  "who did this", never "who typed the row". Client-appendable kinds are
  marked **C** and are the only kinds reachable via `events:append` (A1-7.4).

**Integrity (Q11, A1-5/A1-6)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `chain_genesis` | platform · `event_store` | `chain_spec:string` | Q11, A1-5.3 |
| `chain_verified` | platform · `integrity` | `trigger:enum{startup,pre_export,on_demand}` `head_seq:int` `head_hash:64hex` `verified_count:int` `duration_ms:int` | Q11, A1-6.2 |
| `chain_break_detected` | platform · `integrity` | `break_kind:enum{preimage_mismatch,link_mismatch,seq_gap,seq_disorder,duplicate_event_id,genesis_invalid,row_count_mismatch}` `break_seq:int` `break_event_id:string` `expected_prev_hash:64hex` `actual_prev_hash:64hex` `verified_count:int` | A1-6.3, A11 |
| `integrity_override` | user (admin) · `integrity` | `reason:string(512)*` `break_event_id:string` `scope:enum{export,internal_view}` | Q11, A1-6.5 |

**Engagement and run lifecycle (SPEC §5 steps 1–2, 8)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `scope_changed` | user · `scope` | `change_kind:enum{allowlist_added,allowlist_removed,blacklist_added,blacklist_removed,roe_changed}` `entry:string(512)` `entry_hash:64hex` | C5, ADR-0005 |
| `engagement_policy_changed` | user · `scope` | `policy_kind:enum{llm_data_policy,approval_timeout,model_role_matrix,risk_tier,notification_channel}` `old_value:string(128)` `new_value:string(128)` `affected_tool_id:string` | ADR-0020 §1, ADR-0012 §7, ADR-0008 |
| `run_started` | user · `runtime` | `llm_data_policy:enum{local_only,cloud_masked,cloud_raw}` `approval_timeout_ms:int` `scope_snapshot_evidence_id:string` | SPEC §5.2, ADR-0020 §1 |
| `run_ended` | platform · `runtime` | `end_reason:enum{completed,failed,cancelled,hard_stop}` `detail:string(512)*` | SPEC §5 |
| `model_config_snapshotted` | platform · `llm_gateway` | `config_evidence_id:string` `matrix_hash:64hex` `role_count:int` | SPEC §7 (reproducibility) |
| `job_spawned` | platform · `spawn_broker` | `spawn_request_event_id:string` `image_digest:string(256)` `network_name:string(128)` | ADR-0017 §3 |
| `task_spawned` | platform · `spawn_broker` | `spawn_request_event_id:string` `tool_id:string` `tool_version:string(64)` `risk_tier:string(32)` `image_digest:string(256)` `approval_id:string` `network_name:string(128)` | Q14, ADR-0017 §2–§3 |
| `container_started` | platform · `runtime` | `subject:enum{job,task}` `container_ref:string(128)` `image_digest:string(256)` | ADR-0007/0017 |
| `container_killed` | platform · `runtime` | `subject:enum{job,task}` `container_ref:string(128)` `kill_reason:enum{completed,timeout,hard_stop,quota,node_lost,error,operator}` `exit_code:int` `duration_ms:int` | ADR-0005, ADR-0017 §3 |
| `hard_stop_fired` | user · `runtime` | `stop_scope:enum{run,engagement}` `reason:string(512)*` `containers_killed:int` `respawn_blocked:bool` | ADR-0005, ADR-0012 §2 |

**Spawn request (Q14, ADR-0017 §2)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `spawn_requested` | orchestrator · `spawn_broker` | `subject:enum{job,task}` `tool_id:string` `tool_version:string(64)` `task_description:string(2048)*` `cpu_millicores:int` `memory_bytes:int` `network:enum{run_isolated,target_only,none}` `risk_tier:string(32)` `image_digest:string(256)` `approval_id:string` | Q14, ADR-0017 §2 (platform derives digest + tier from the registry; `""` when validation failed before derivation) |

**Command execution, evidence, results (ADR-0009 §1–§3, SPEC §5 steps 3, 7)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `command_executed` **C** | worker / node | `command:string(2048)*` `target:string(256)*` `tool_id:string` `tool_version:string(64)` `exit_code:int` `duration_ms:int` `output_bytes:int` `output_evidence_id:string` `redacted:bool` | ADR-0009 §1: command, output **reference**, actor (envelope), target, timestamp (envelope) |
| `evidence_stored` | platform · `evidence` | `evidence_id:string` `evidence_kind:enum{command_output,screenshot,file_capture,memory_dump,packet_capture,browser_session,config_snapshot,report_artifact,other}` `media_type:string(128)` `size_bytes:int` `sha256:64hex` `source:enum{worker,node,orchestrator,browser,platform}` `redacted:bool` | ADR-0009 §2 |
| `task_result` **C** | worker / node | `status:enum{succeeded,failed,partial}` `result_summary:string(2048)*` `duration_ms:int` `command_count:int` `revert_event_ids:array[string]` `error_kind:string` | Q6 worker scope `task:result` |
| `revert_recorded` **C** | worker / node | `effect_kind:enum{account_created,account_modified,acl_changed,file_dropped,scheduled_task,service_installed,registry_edit,config_changed,credential_changed,persistence_added,other}` `target:string(256)*` `revert_action:string(2048)*` `revertable:bool` `tool_id:string` `state_change_evidence_id:string` | ADR-0009 §3. The revert record **is** the event: it is identified by its `event_id` (A0-1.2 declares no revert prefix and A1 adds none) |

**Approvals (Q10, ADR-0018 §1–§4, SPEC §5 steps 5–6)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `approval_requested` | orchestrator · `approval` | `approval_id:string` `fingerprint_hash:64hex` `tool_id:string` `tool_version:string(64)` `target:string(256)*` `action_summary:string(512)*` `risk_tier:string(32)` `expires_at:timestamp` `untrusted_context:bool` `request_event_id:string` | ADR-0018 §1 (fingerprint, prose *in addition*), §4 (`untrusted_context`), ADR-0012 §7 (`expires_at`) |
| `approval_granted` | user · `approval` | `approval_id:string` `fingerprint_hash:64hex` `expires_at:timestamp` `single_use:bool` `queue_wait_ms:int` | ADR-0012 §1 (approver identity = envelope actor + platform time), Q10 (`single_use` is `true` for every v1 approval) |
| `approval_denied` | user · `approval` | `approval_id:string` `fingerprint_hash:64hex` `reason:string(512)*` | ADR-0012 §1 |
| `approval_expired` | platform · `approval` | `approval_id:string` `fingerprint_hash:64hex` `expires_at:timestamp` `queue_wait_ms:int` | ADR-0012 §7: orchestrator replans on `approval_expired` (A0-3.1) |
| `approval_executed` | platform · `approval` | `approval_id:string` `fingerprint_hash:64hex` `revalidated:bool` `single_use_consumed:bool` `expires_at:timestamp` `tool_id:string` `tool_version:string(64)` `target:string(256)*` `action_summary:string(512)*` | ADR-0018 §2–§3 (execution-time re-validation, action + fingerprint recorded together), Q10 (consumption record) |

**LLM traffic (ADR-0020 §2–§5, SPEC §7)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `llm_call` | platform · `llm_gateway` | `model_role:enum{orchestrator,enumeration,research,summarizer}` `model_name:string(128)` `endpoint_name:string(128)` `egress_policy:enum{local_only,cloud_masked,cloud_raw}` `status:enum{ok,error}` `error_kind:string` `request_bytes:int` `response_bytes:int` `prompt_tokens:int` `completion_tokens:int` `masked_entity_count:int` `excluded_secret_count:int` `duration_ms:int` `request_evidence_id:string` `response_evidence_id:string` | ADR-0020 §2 (endpoint, model, payload size, policy applied — **metadata only**), §3 (`masked_entity_count`), §4 (`excluded_secret_count`), §5 (operator-visible egress log), ADR-0014 roles. `endpoint_name` is the configured name: a URL MUST NOT be stored (it may embed a credential, A0-3.7). `prompt_tokens`/`completion_tokens` are upstream-reported and untrusted-in-origin but are integers, so A1-4.4 does not apply; `egress_policy` values are ADR-0020's hyphenated names normalized to A0-8.5 |

**Graph mutation (ADR-0016 §1/§2/§4; the fields A2 must reference — A1-3.6)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `graph_node_written` | platform · `graph` | `graph_node_id:string` `node_kind:string(32)` `source_event_id:string` `supersedes_graph_node_id:string` `quarantined:bool` `summary_bytes:int` `attrs_count:int` | ADR-0016 §1 (provenance), §4 (`supersedes`, never overwrite), Q2 |
| `graph_edge_written` | platform · `graph` | `graph_edge_id:string` `edge_kind:string(32)` `from_graph_node_id:string` `to_graph_node_id:string` `source_event_id:string` `quarantined:bool` | ADR-0016 §1 |
| `graph_node_quarantined` | user or platform · `graph` | `graph_node_id:string` `quarantine_kind:enum{out_of_scope_discovery,blacklist_match,operator_quarantine,operator_release}` `reason:string(512)*` `source_event_id:string` | ADR-0016 §2, C5, Q5 |
| `report_inclusion_changed` | user · `graph` | `graph_node_id:string` `included:bool` `reason:string(512)*` | Q5 (operator removes a quarantined discovery from the report). The flag itself is mutable graph state (A2) and MUST never be an order key (A0-4.3) |

**Cleanup / revert execution (ADR-0009 §4, SPEC §5 step 10)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `cleanup_planned` | platform · `cleanup` | `plan_evidence_id:string` `revert_event_ids:array[string]` `non_revertable_event_ids:array[string]` `planned_action_count:int` `approval_id:string` | ADR-0009 §4 (plan from revert records, human approval), non-revertable effects documented |
| `cleanup_executed` | platform · `cleanup` | `revert_event_id:string` `status:enum{reverted,failed,skipped}` `command_event_id:string` `detail:string(512)*` | ADR-0009 §4 |
| `cleanup_verified` | platform · `cleanup` | `revert_event_id:string` `verified:bool` `verification_evidence_id:string` `detail:string(512)*` | ADR-0009 §4 (execute → verify) |

**Enforcement denials and agent errors (ADR-0005, C5/C9, ADR-0018 §2)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `scope_denied` | requesting principal · `scope` | `action_kind:enum{tool_exec,spawn,graph_read,llm_call,api}` `attempted_target:string(256)*` `tool_id:string` `request_event_id:string` | ADR-0005 (a denied action is auditable), C5 allowlist |
| `blacklist_denied` | requesting principal · `scope` | `action_kind:enum{tool_exec,spawn,graph_read,llm_call,api}` `attempted_target:string(256)*` `blacklist_entry:string(256)` `tool_id:string` `request_event_id:string` | C5: blacklist beats allowlist beats approval — separately filterable by design (A1-8.3) |
| `action_blocked` | requesting principal · varying | `reason:enum{approval_consumed,approval_expired_at_exec,fingerprint_mismatch,image_not_allowed,quota_exceeded,hard_stop_active,node_not_paired,llm_egress_blocked,secret_excluded,graph_write_rejected,append_not_permitted,unknown_event_kind,token_revoked}` `action_kind:enum{tool_exec,spawn,graph_read,graph_write,llm_call,api,events_append}` `attempted_target:string(256)*` `detail:string(512)*` `tool_id:string` `approval_id:string` `request_event_id:string` | Q10 (`approval_consumed`), ADR-0018 §2 (`fingerprint_mismatch`), ADR-0017 §2 (image/quota), ADR-0020 §4 (`secret_excluded`), Q3 (`graph_write_rejected`, `unknown_event_kind`), A1-7.5 (`append_not_permitted`) |
| `agent_error` | failing principal · `api` | `error_kind:string` `origin:string(128)` `message:string(512)*` `retryable:bool` | ADR-0012 §2 (`agent_error`), ADR-0019 §2–§3 (`origin` = `component.Function`, `message` already redacted per A0-3.7), C9 |

**Notification delivery (ADR-0012 §2–§6)**

| Kind | Composed by | Payload fields | Source |
|---|---|---|---|
| `notification_sent` | platform · `notify` | `notification_kind:enum{approval_required,scan_started,scan_finished,agent_error,hard_stop_fired,cleanup_proposed}` `channel:enum{webhook,sse}` `target_name:string(128)` `delivery_status:enum{delivered,failed}` `attempt:int` `http_status:int` `duration_ms:int` `related_event_id:string` | ADR-0012 §2 (the six notification kinds, spelled as the ADR spells them), §6 (retries + delivery log are themselves audited events). `target_name` is the configured channel name: a webhook URL MUST NOT be stored (A0-3.7). These are **notification** kinds, distinct from A1 event kinds; only `hard_stop_fired` and `agent_error` exist in both vocabularies |

- **A1-3.4** Kinds marked **C** (`command_executed`, `task_result`,
  `revert_recorded`) are the only client-appendable kinds. Every other kind is
  platform-composed and unreachable through `events:append` for any machine
  principal → `forbidden` (A0-3.1, A1-7.4). _A worker that could append
  `approval_granted` would own the safety model._
- **A1-3.5** A kind is added only for an occurrence that must be independently
  filterable (A1-8.3) and independently reportable. Everything else is a
  payload field or a `reason`/`status` enum value inside an existing kind —
  `action_blocked` is the designated home for enforcement denials that do not
  deserve their own filter. Adding a kind is additive (A0-6.5); **reshaping an
  existing kind's payload is not** (A0-6.6, A1-4.1) — a written event's
  canonical bytes are immutable.
- **A1-3.6** A2 coordination (binding on A2): the graph-provenance field names
  A2 MUST reference and MUST NOT rename are `graph_node_id`, `graph_edge_id`,
  `node_kind`, `edge_kind`, `source_event_id`, `supersedes_graph_node_id`,
  `from_graph_node_id`, `to_graph_node_id`, `quarantined`. A2 owns the closed
  `node_kind`/`edge_kind` value lists; A1 stores the values as capped strings
  (`string(32)`) so an A2 kind addition needs no A1 change (A0-6.5). A7
  coordination: A1 stores `fingerprint_hash` and `risk_tier` as opaque values;
  A7 owns their content and vocabulary.
- **A1-3.7** SPEC §5 step coverage (completeness check, no normative force of
  its own): 1 → `chain_genesis`, `scope_changed`, `engagement_policy_changed` ·
  2 → `run_started`, `model_config_snapshotted`, `job_spawned`,
  `container_started` · 3 → `spawn_requested`, `task_spawned`,
  `command_executed`, `evidence_stored`, `task_result`, `graph_*` · 4 →
  `llm_call`, `graph_*` · 5 → `approval_requested`, `notification_sent`,
  `scope_denied`, `blacklist_denied`, `action_blocked` · 6 →
  `approval_granted`/`_denied`/`_expired` · 7 → `command_executed`,
  `evidence_stored`, `revert_recorded`, `approval_executed` · 8 →
  `agent_error`, `llm_call`, `graph_*`, `container_killed` · 9 →
  `chain_verified{pre_export}`, `integrity_override`,
  `evidence_stored{report_artifact}` · 10 → `cleanup_planned`,
  `cleanup_executed`, `cleanup_verified`, `hard_stop_fired`.

### A1-4 · Payload rules

<!-- A1-4 -->

### A1-5 · Hash chain

<!-- A1-5 -->

### A1-6 · Verification and failure behaviour

<!-- A1-6 -->

### A1-7 · Write path

<!-- A1-7 -->

### A1-8 · Read and stream guarantees

<!-- A1-8 -->

## 4. Types

<!-- A1-types -->

## 5. Traceability

<!-- A1-trace -->

## 6. Open for product owner

<!-- A1-po -->
