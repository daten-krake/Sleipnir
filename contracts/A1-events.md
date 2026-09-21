# A1 — Event log

## 1. Header

| | |
|---|---|
| **Contract id** | A1 |
| **Status** | `Frozen` — product owner, 2026-09-21, PR #2 (`contracts/README.md` lifecycle: Draft → Frozen → Implemented) |
| **Owner** | architect |
| **Gates** | A2 (graph provenance fields) · A3 (stage views read the log) · A4 (append/read/SSE endpoints) · A5 (scopes per principal) · A7 (fingerprint + spawn payload fields) · `internal/events` (store seam) · report builder · every component that composes an event |
| **Implements** | ADR-0009 §1–§4 · ADR-0012 §1/§2/§6/§7 · ADR-0016 §1/§2/§4 · ADR-0017 §2–§3 · ADR-0018 §1–§4 · ADR-0019 §3/§5 · ADR-0020 §2–§5 · ADR-0005 · ADR-0011 · ADR-0013 · SPEC §5 steps 1–10, §6, §7, §8, C5/C8/C9/C11 · Q6 (+ worker addendum), Q10, Q11, Q12, Q13, Q14 · adversarial A1/A9/A11/A12 |
| **Depends on** | A0 in full — ids (A0-1), canonical JSON (A0-2), error kinds (A0-3), pagination (A0-4), time (A0-5), versioning (A0-6), size caps (A0-7), field conventions (A0-8). A1 restates no A0 rule; it cites it. |
| **External refs** | FIPS 180-4 (SHA-256) · RFC 8785 via A0-2.2 · RFC 3339 via A0-5.1 |
| **Authority** | ADR > SPEC > DESIGN > contract. A clause here that contradicts an Accepted ADR is a defect in this document. |
| **Review provenance** | Clause rationale may cite review finding ids — `P-nn` (buildability and cross-document consistency) and `C-nn`/`S-nn`/`F-nn`/`D-nn`/`T-nn`/`E-nn` (adversarial) — from `docs/reviews/2026-09-11-contract-review-principal.md` and `docs/reviews/2026-09-11-contract-review-adversarial.md`; the merged fix plan is `docs/reviews/2026-09-11-contract-fix-plan.md`. Those ids are audit trail, not normative references: no clause depends on them. |

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
  when unset — no `omitempty`, no absence, no `null` (A0-2.14, A0-8.3). Every
  slice and map field of a canonicalized type MUST be non-nil before marshaling;
  the constructors initialize them to empty (A0-2.14). A canonical event document
  containing `null` is a platform defect → `internal`. This
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
  `type="platform"`). A1-3.3's `actor (type, component)` column gives the
  literal pair for every kind.

  | A1 `actor.type` | A2 `principal_kind` | id shape (A0-1.2) |
  |---|---|---|
  | `platform` | `platform` | `""` |
  | `orchestrator` | `orchestrator` | `job_` |
  | `worker` | `worker` | `task_` |
  | `node` | `node` | `slp_node_` |
  | `user` | `user` | `usr_` (AM-1) |

  One vocabulary, two documents: A2-5.3 adopts A1-2.1's list verbatim; `operator` is
  renamed `user` and `operator_id` is renamed `user_id` platform-wide. The prose word
  "operator" (a human role, SPEC §3) is unaffected — only the enum value changes.
  `Tests: TestActorComponentIsEmptyForNonPlatform, TestActorVocabularyMatchesA2`.
- **A1-2.2** `principal_id` is an A0-1.2 identifier selected by `type`, and
  MUST be validated per A0-1.5 at composition:

  | `type` | `principal_id` | `component` |
  |---|---|---|
  | `user` | user principal id — A0-1.2 registers `usr_` (`^usr_B{26}$`, 30 B; **AM-1, resolved and confirmed — §6.2**). A username or e-mail MUST NOT be used: both are mutable and both are personal data in a customer export | `""` |
  | `orchestrator` | the `job_` id of the orchestrator container | `""` |
  | `worker` | the `task_` id of the worker container | `""` |
  | `node` | the `slp_node_` id (Q9) | `""` |
  | `platform` | `""` | the platform subsystem that composed the event, from the closed list of A1-2.4 |

  **AM-1 — resolved and confirmed at the Freeze (product owner decision
  2026-09-21, PR #2 item D3):** A0-1.2 registers the human-principal prefix
  `usr_` (`^usr_B{26}$`, 30 B) and A0 §4 adds `KindUser`. A0 owns id *shapes*;
  delegating the spelling to A5 would split A0-1.5 validation across two
  contracts. Every user-composed A1 kind (`actor.principal_id`, A1-2.2) and
  every A2 operator write (`user_id`, A2-5.3) validates against it. The prefix
  spelling is confirmed and is additive-only from here (A0-1.10). The identical
  sentence stands in A0 §6 and A2 §6.2.

- **A1-2.3** Only the platform composes events. `actor` is stamped from the
  authenticated principal and the request context; a client MUST NOT supply
  `actor`, `event_id`, `engagement_id`, `run_id`, `job_id`, `task_id`,
  `node_id`, `seq`, `prev_hash`, `hash`, `recorded_at` or `occurred_at`. Any
  of them in an append body is an unknown field on a write → `validation`
  naming the field (A0-6.2). _Q3/Q6: never trust client discipline; a
  self-declared actor is the cheapest possible audit forgery._
  `Tests: TestClientCannotSupplyEnvelopeFields` (table over all 12 fields named
  above → `validation` naming the field), `TestActorCannotBeForged`,
  `TestUntrustedFlagCannotBeSupplied`.
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
  `Tests: TestRunBindingMismatchIsForbidden, TestWorkerAndNodeHaveNoReadScope`.
- **A1-2.7** `actor.type="platform"` MUST NOT be selectable or reachable
  through `/api/v1` by any machine principal; platform-composed kinds
  (A1-3.4) are unreachable for them by construction. A user principal can
  trigger a platform-composed event only through the endpoint that performs
  the action (hard stop, override, quarantine) — never by appending.
  `Tests: TestMachinePrincipalCannotReachIntegrityKinds,`
  `TestPlatformActorNotSelectableViaAPI`.
- **A1-2.8** Provenance for A2 (ADR-0016 §1) is the pair
  (`event_id`, `actor`) of the event that caused a graph write, carried into
  the graph row by the platform and echoed back in `graph_node_written` /
  `graph_edge_written` payloads as `source_event_id` (A1-3.6). Provenance is
  never client-supplied.

### A1-3 · Closed event taxonomy

- **A1-3.1** The event kind list is **closed** and complete on day one (42
  kinds, below). An unknown or malformed `kind` on a write is hard-rejected
  with `validation` (A0-6.3, Q3); on a read a client MUST preserve the raw
  string and skip what it cannot handle (A0-6.3). Kind spelling follows
  A0-8.5: `^[a-z][a-z0-9_]{0,31}$`, `<subject>_<past_participle>`.
  `Tests: TestKindListIs42AndClosed, TestUnknownKindIsRejectedOnWrite`.
- **A1-3.2** Notation in the payload column: `name:type`; `(N)` = capped at N
  UTF-8 bytes of the decoded value (A0-7.3), mechanism per A1-4.5; `*` =
  untrusted-content field (A1-4.4); `64hex` = A0-2.15 digest; `enum{…}` =
  closed list (A0-8.5); `array[string]` = ordered string array (A1-4.7).
  Every listed field is **always present** with its zero value (A0-2.14);
  a field not listed MUST NOT appear (`validation`, A0-6.2).
- **A1-3.3** The `actor (type, component)` column gives, for every kind, the
  literal pair the platform stamps (A1-2.1): the principal whose action the
  event records is `actor.type`, and the platform subsystem that writes the row
  is `actor.component` — `""` unless `actor.type="platform"`. A platform
  subsystem recording a client's request stamps the **client** as actor:
  `actor` answers "who did this", never "who typed the row". Client-appendable
  kinds are marked **C** and are the only kinds reachable via `events:append`
  (A1-7.4).

**Integrity (Q11, A1-5/A1-6)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `chain_genesis` | `(platform, "event_store")` | `chain_spec:string` | Q11, A1-5.3 |
| `chain_verified` | `(platform, "integrity")` | `trigger:enum{startup,pre_export,on_demand}` `head_seq:int` `head_hash:64hex` `verified_count:int` `duration_ms:int` | Q11, A1-6.2 |
| `chain_break_detected` | `(platform, "integrity")` | `break_kind:enum{preimage_mismatch,link_mismatch,seq_gap,seq_disorder,duplicate_event_id,genesis_invalid,row_count_mismatch,engagement_mismatch,head_regression}` `break_seq:int` `break_event_id:string` `expected_prev_hash:64hex` `actual_prev_hash:64hex` `verified_count:int` | A1-6.3, A11, A1-5.8 (`head_regression`) |
| `integrity_override` | `(user, "")` — admin role only (A1-6.5) | `reason:string(512)*` `break_event_id:string` `scope:enum{export,internal_view}` | Q11, A1-6.5 |
| `artifact_released` | `(user, "")` **or** `(platform, "integrity")` | `artifact_kind:enum{report_html,report_pdf,findings_json,evidence_bundle,verification_bundle}` `artifact_evidence_id:string` `head_seq:int` `head_hash:64hex` `integrity_state:enum{verified,failed_overridden}` `override_event_id:string` `recipient_ref:string(128)` | Q11, A1-6.6 (the export stamp as a chained event), A1-6.5 (the `scope:"export"` override it names, ADR-0021) |

**Engagement and run lifecycle (SPEC §5 steps 1–2, 8)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `engagement_created` | `(user, "")` | `client_ref:string(128)` `roe_evidence_id:string` `policy_evidence_id:string` `operator_count:int` | SPEC §5.1, ADR-0005. Composed at `seq` **1** of the engagement chain — `chain_genesis` stays the integrity anchor at `seq` 0 (A1-5.3) and carries no engagement metadata |
| `scope_changed` | `(user, "")` | `change_kind:enum{allowlist_added,allowlist_removed,blacklist_added,blacklist_removed,roe_changed}` `entry:string(512)` `entry_hash:64hex` | C5, ADR-0005. A **global**-blacklist change is composed into every affected engagement chain (A1-4.2, adversarial C-04) |
| `engagement_policy_changed` | `(user, "")` | `policy_kind:enum{llm_data_policy,approval_timeout,model_role_matrix,risk_tier,notification_channel}` `old_value:string(128)` `new_value:string(128)` `affected_tool_id:string` | ADR-0020 §1, ADR-0012 §7, ADR-0008 |
| `run_started` | `(user, "")` | `llm_data_policy:enum{local_only,cloud_masked,cloud_raw}` `approval_timeout_ms:int` `scope_snapshot_evidence_id:string` | SPEC §5.2, ADR-0020 §1 |
| `run_ended` | `(platform, "runtime")` | `end_reason:enum{completed,failed,cancelled,hard_stop}` `detail:string(512)*` | SPEC §5 |
| `model_config_snapshotted` | `(platform, "llm_gateway")` | `config_evidence_id:string` `matrix_hash:64hex` `role_count:int` | SPEC §7 (reproducibility) |
| `job_spawned` | `(platform, "spawn_broker")` | `spawn_request_event_id:string` `image_digest:string(256)` `network_name:string(128)` | ADR-0017 §3 |
| `task_spawned` | `(platform, "spawn_broker")` | `spawn_request_event_id:string` `tool_id:string` `tool_version:string(64)` `risk_tier:string(32)` `image_digest:string(256)` `approval_id:string` `network_name:string(128)` `target_graph_node_id:string` | Q14, ADR-0017 §2–§3 |
| `container_started` | `(platform, "runtime")` | `subject:enum{job,task}` `container_ref:string(128)` `image_digest:string(256)` | ADR-0007/0017 |
| `container_killed` | `(platform, "runtime")` | `subject:enum{job,task}` `container_ref:string(128)` `kill_reason:enum{completed,timeout,hard_stop,quota,node_lost,error,operator}` `exit_code:int` `duration_ms:int` `stop_event_id:string` | ADR-0005, ADR-0017 §3 |
| `hard_stop_fired` | `(user, "")` | `stop_scope:enum{run,engagement}` `reason:string(512)*` `containers_killed:int` `respawn_blocked:bool` | ADR-0005, ADR-0012 §2 |
| `engagement_closed` | `(user, "")` | `close_reason:enum{completed,cancelled,abandoned}` `report_evidence_id:string` | SPEC §5 step 10 (the engagement ends after cleanup); the last user-composed fact of a chain |

**Spawn request (Q14, ADR-0017 §2)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `spawn_requested` | `(orchestrator, "")` | `subject:enum{job,task}` `tool_id:string` `tool_version:string(64)` `task_description:string(2048)*` `cpu_millicores:int` `memory_bytes:int` `network:enum{run_isolated,target_only,none}` `risk_tier:string(32)` `image_digest:string(256)` `approval_id:string` `target_graph_node_id:string` | Q14, ADR-0017 §2 (platform derives digest + tier from the registry; `""` when validation failed before derivation) |

**Command execution, evidence, results (ADR-0009 §1–§3, SPEC §5 steps 3, 7)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `command_executed` **C** | `(worker, "")` **or** `(node, "")` | `command:string(2048)*` `target:string(256)*` `tool_id:string` `tool_version:string(64)` `exit_code:int` `duration_ms:int` `output_bytes:int` `output_evidence_id:string` `redacted:bool` | ADR-0009 §1: command, output **reference**, actor (envelope), target, timestamp (envelope) |
| `evidence_stored` | `(platform, "evidence")` | `evidence_id:string` `evidence_kind:enum{command_output,screenshot,file_capture,memory_dump,packet_capture,browser_session,config_snapshot,report_artifact,other}` `media_type:string(128)` `size_bytes:int` `sha256:64hex` `source:enum{worker,node,orchestrator,browser,platform}` `redacted:bool` | ADR-0009 §2 |
| `task_result` **C** | `(worker, "")` **or** `(node, "")` | `status:enum{succeeded,failed,partial}` `result_summary:string(2048)*` `duration_ms:int` `command_count:int` `revert_event_ids:array[string]` `error_kind:string` | Q6 worker scope `task:result` |
| `revert_recorded` **C** | `(worker, "")` **or** `(node, "")` | `effect_kind:enum{account_created,account_modified,acl_changed,file_dropped,scheduled_task,service_installed,registry_edit,config_changed,credential_changed,persistence_added,other}` `target:string(256)*` `revert_action:string(2048)*` `revertable:bool` `tool_id:string` `state_change_evidence_id:string` | ADR-0009 §3. The revert record **is** the event: it is identified by its `event_id` (A0-1.2 declares no revert prefix and A1 adds none) |

**Approvals (Q10, ADR-0018 §1–§4, SPEC §5 steps 5–6)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `approval_requested` | `(orchestrator, "")` | `approval_id:string` `fingerprint_hash:64hex` `action_spec_evidence_id:string` `target_graph_node_id:string` `argv_hash:64hex` `tool_id:string` `tool_version:string(64)` `target:string(256)*` `action_summary:string(512)*` `risk_tier:string(32)` `expires_at:timestamp` `untrusted_context:bool` `request_event_id:string` | ADR-0018 §1 (fingerprint, prose *in addition*), §4 (`untrusted_context`), ADR-0012 §7 (`expires_at`) |
| `approval_granted` | `(user, "")` | `approval_id:string` `fingerprint_hash:64hex` `expires_at:timestamp` `single_use:bool` `queue_wait_ms:int` | ADR-0012 §1 (approver identity = envelope actor + platform time), Q10 (`single_use` is `true` for every v1 approval) |
| `approval_denied` | `(user, "")` | `approval_id:string` `fingerprint_hash:64hex` `reason:string(512)*` | ADR-0012 §1 |
| `approval_expired` | `(platform, "approval")` | `approval_id:string` `fingerprint_hash:64hex` `expires_at:timestamp` `queue_wait_ms:int` | ADR-0012 §7: orchestrator replans on `approval_expired` (A0-3.1) |
| `approval_executed` | `(platform, "approval")` | `approval_id:string` `fingerprint_hash:64hex` `action_spec_evidence_id:string` `target_graph_node_id:string` `argv_hash:64hex` `revalidated:bool` `single_use_consumed:bool` `expires_at:timestamp` `tool_id:string` `tool_version:string(64)` `target:string(256)*` `action_summary:string(512)*` | ADR-0018 §2–§3 (execution-time re-validation, action + fingerprint recorded together), Q10 (consumption record) |

**LLM traffic (ADR-0020 §2–§5, SPEC §7)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `llm_call` | `(platform, "llm_gateway")` | `model_role:enum{orchestrator,enumeration,research,summarizer}` `model_name:string(128)` `endpoint_name:string(128)` `egress_policy:enum{local_only,cloud_masked,cloud_raw}` `status:enum{ok,error}` `error_kind:string` `request_bytes:int` `response_bytes:int` `prompt_tokens:int` `completion_tokens:int` `masked_entity_count:int` `excluded_secret_count:int` `duration_ms:int` `request_evidence_id:string` `response_evidence_id:string` | ADR-0020 §2 (endpoint, model, payload size, policy applied — **metadata only**), §3 (`masked_entity_count`), §4 (`excluded_secret_count`), §5 (operator-visible egress log), ADR-0014 roles. `endpoint_name` is the configured name: a URL MUST NOT be stored (it may embed a credential, A0-3.7). `prompt_tokens`/`completion_tokens` are upstream-reported and untrusted-in-origin but are integers, so A1-4.4 does not apply; `egress_policy` values are ADR-0020's hyphenated names normalized to A0-8.5 |

**Graph mutation (ADR-0016 §1/§2/§4; the fields A2 must reference — A1-3.6)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `graph_node_written` | `(platform, "graph")` | `graph_node_id:string` `node_kind:string(32)` `source_event_id:string` `supersedes_graph_node_id:string` `quarantined:bool` `content_hash:64hex` `dedup_hit:bool` `summary_bytes:int` `attrs_count:int` | ADR-0016 §1 (provenance), §4 (`supersedes`, never overwrite), Q2, A2-4.6 (`content_hash`), A2-4.7 (`dedup_hit`) |
| `graph_edge_written` | `(platform, "graph")` | `graph_edge_id:string` `edge_kind:string(32)` `from_graph_node_id:string` `to_graph_node_id:string` `source_event_id:string` `quarantined:bool` `dedup_hit:bool` | ADR-0016 §1, A2-4.7 (`dedup_hit`) |
| `graph_node_quarantined` | `(user, "")` **or** `(platform, "graph")` | `graph_node_id:string` `quarantine_kind:enum{out_of_scope_discovery,blacklist_match,operator_quarantine}` `reason:string(512)*` `source_event_id:string` | ADR-0016 §2, C5, Q5. There is **no** `operator_release` value (A1-3.6, §6 item 14) |
| `quarantine_recomputed` | `(platform, "graph")` | `trigger:enum{scope_changed,blacklist_changed,node_written,edge_written}` `trigger_event_id:string` `nodes_evaluated:int` `nodes_quarantined:int` `nodes_released:int` `duration_ms:int` | A2-8.5 (a scope/blacklist change recomputes quarantine and the recomputation is itself recorded), ADR-0016 §2, C5. Per-node effects are separate `graph_node_quarantined` events; this event records the batch and its trigger |
| `graph_edge_retracted` | `(user, "")` **or** `(platform, "graph")` | `graph_edge_id:string` `edge_kind:string(32)` `from_graph_node_id:string` `to_graph_node_id:string` `reason:string(512)*` `source_event_id:string` | A2-3.9 (the only mutable edge field is always accompanied by an A1 event carrying the reason), ADR-0016 §4 (self-correction). `source_event_id` is the observation that contradicted the edge, `""` when operator-initiated |
| `report_inclusion_changed` | `(user, "")` | `graph_node_id:string` `included:bool` `reason:string(512)*` | Q5 (operator removes a quarantined discovery from the report). The flag itself is mutable graph state (A2) and MUST never be an order key (A0-4.3) |

**Cleanup / revert execution (ADR-0009 §4, SPEC §5 step 10)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `cleanup_planned` | `(platform, "cleanup")` | `plan_evidence_id:string` `revert_event_ids:array[string]` `non_revertable_event_ids:array[string]` `planned_action_count:int` `approval_id:string` | ADR-0009 §4 (plan from revert records, human approval), non-revertable effects documented |
| `cleanup_executed` | `(platform, "cleanup")` | `revert_event_id:string` `status:enum{reverted,failed,skipped}` `command_event_id:string` `detail:string(512)*` | ADR-0009 §4 |
| `cleanup_verified` | `(platform, "cleanup")` | `revert_event_id:string` `verified:bool` `verification_evidence_id:string` `detail:string(512)*` | ADR-0009 §4 (execute → verify) |

**Enforcement denials and agent errors (ADR-0005, C5/C9, ADR-0018 §2)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `scope_denied` | `(requesting principal's type, "")` | `action_kind:enum{tool_exec,spawn,graph_read,llm_call,api}` `attempted_target:string(256)*` `tool_id:string` `request_event_id:string` | ADR-0005 (a denied action is auditable), C5 allowlist |
| `blacklist_denied` | `(requesting principal's type, "")` | `action_kind:enum{tool_exec,spawn,graph_read,llm_call,api}` `attempted_target:string(256)*` `blacklist_entry:string(256)` `tool_id:string` `request_event_id:string` | C5: blacklist beats allowlist beats approval — separately filterable by design (A1-8.3) |
| `action_blocked` | `(requesting principal's type, "")` | `reason:enum{approval_consumed,approval_expired_at_exec,approval_metadata_mismatch,fingerprint_mismatch,target_quarantined,image_not_allowed,quota_exceeded,hard_stop_active,node_not_paired,llm_egress_blocked,secret_excluded,graph_write_rejected,append_not_permitted,append_rejected,unknown_event_kind,token_revoked}` `action_kind:enum{tool_exec,spawn,graph_read,graph_write,llm_call,api,events_append}` `attempted_target:string(256)*` `detail:string(512)*` `tool_id:string` `approval_id:string` `request_event_id:string` | Q10 (`approval_consumed`), ADR-0018 §2 (`fingerprint_mismatch`), ADR-0017 §2 (image/quota), ADR-0020 §4 (`secret_excluded`), Q3 (`graph_write_rejected`, `unknown_event_kind`), A1-7.5 (`append_not_permitted`; `append_rejected` = a refused `events:append` — schema, cap or secret scan — stays observable, adversarial A1) |
| `agent_error` | `(failing principal's type, "")` | `error_kind:string` `origin:string(128)` `message:string(512)*` `retryable:bool` | ADR-0012 §2 (`agent_error`), ADR-0019 §2–§3 (`origin` = `component.Function`, `message` already redacted per A0-3.7), C9 |

**Notification delivery (ADR-0012 §2–§6)**

| Kind | `actor (type, component)` | Payload fields | Source |
|---|---|---|---|
| `notification_sent` | `(platform, "notify")` | `notification_kind:enum{approval_required,scan_started,scan_finished,agent_error,hard_stop_fired,cleanup_proposed,chain_head_anchor}` `channel:enum{webhook,sse}` `target_name:string(128)` `delivery_status:enum{delivered,failed}` `attempt:int` `http_status:int` `duration_ms:int` `related_event_id:string` | ADR-0012 §2 (its six notification kinds, spelled as the ADR spells them; A1 adds the seventh, `chain_head_anchor`), §6 (retries + delivery log are themselves audited events). `target_name` is the configured channel name: a webhook URL MUST NOT be stored (A0-3.7). These are **notification** kinds, distinct from A1 event kinds; only `hard_stop_fired` and `agent_error` exist in both vocabularies. chain_head_anchor is the A1-5.8 (4) out-of-band head anchor (product owner decision 2026-09-21, PR #2 item D5) — an A1 addition to the ADR-0012 §2 vocabulary, recorded here because §2's list predates the anchoring decision |

The subsystem named in the Source column is the code path that writes the row;
it appears in `actor.component` **only** when `actor.type="platform"` (A1-2.1).
For every non-platform row `actor.component` is `""`. `actor` is inside the
digest (A1-5.2).
`Tests: TestActorComponentIsEmptyForNonPlatform`

`action_blocked.reason` gains `target_quarantined` (additive): a spawn or action
request cited a `gn_` id whose node is quarantined; refused **before** approval
routing (A2-8.3, ADR-0018 §2). `quarantine_recomputed.trigger` is the closed list
`scope_changed · blacklist_changed · node_written · edge_written` — the last two are
the recomputations a new node or a new edge touching a quarantined node causes
(A2-8.2).

`graph_node_written` carries `content_hash:64hex` (the A2-4.6 fingerprint of the
written node) and `dedup_hit:bool`; `graph_edge_written` carries `dedup_hit:bool`
(edges have no fingerprint). A dedup collapse (A2-4.7) MUST still emit the event with
`dedup_hit:true`, so the chain distinguishes "new evidence recorded" from "duplicate
absorbed" — including every offline-node replay (ADR-0013).

- **A1-3.4** Kinds marked **C** (`command_executed`, `task_result`,
  `revert_recorded`) are the only client-appendable kinds. Every other kind is
  platform-composed and unreachable through `events:append` for any machine
  principal → `forbidden` (A0-3.1, A1-7.4). _A worker that could append
  `approval_granted` would own the safety model._
  `Tests: TestWorkerCannotAppendNonCKind` (each non-C kind → `forbidden` +
  `action_blocked{append_not_permitted}`), `TestOrchestratorCannotClaimCommandExecuted`.
- **A1-3.5** A kind is added only for an occurrence that must be independently
  filterable (A1-8.3) and independently reportable. Everything else is a
  payload field or a `reason`/`status` enum value inside an existing kind —
  `action_blocked` is the designated home for enforcement denials that do not
  deserve their own filter. Adding a kind is additive (A0-6.5); **reshaping an
  existing kind's payload is not** (A0-6.6, A1-4.1) — a written event's
  canonical bytes are immutable. The three kinds this revision added under this
  rule are `engagement_created`, `engagement_closed` and `artifact_released`
  (T-01/T-02): the engagement envelope and every released artifact must be
  independently filterable and independently reportable, and `artifact_released`
  is the enforcement point of A1-6.5's override attribution (ADR-0021).
- **A1-3.6** A2 coordination (binding on A2): the graph-provenance field names
  A2 MUST reference and MUST NOT rename are `graph_node_id`, `graph_edge_id`,
  `node_kind`, `edge_kind`, `source_event_id`, `supersedes_graph_node_id`,
  `from_graph_node_id`, `to_graph_node_id`, `quarantined`, `content_hash`,
  `dedup_hit`. A2 owns the closed `node_kind`/`edge_kind` value lists; A1 stores
  the values as capped strings (`string(32)`) so an A2 kind addition needs no A1
  change (A0-6.5). A7 coordination: A1 stores `fingerprint_hash`, `argv_hash`,
  `action_spec_evidence_id` and `risk_tier` as opaque values — **A7 owns the
  *content* of these fields, A1 their *presence*.**

  Quarantine vocabulary: A2 owns the **state** vocabulary (`quarantine_reason`:
  `out_of_scope`, `blacklisted`); A1 owns the **occurrence** vocabulary
  (`quarantine_kind`: `out_of_scope_discovery`, `blacklist_match`,
  `operator_quarantine`). Mapping — `out_of_scope_discovery → out_of_scope` ·
  `blacklist_match → blacklisted` · `operator_quarantine → (the reason already in
  force)`. There is no `operator_release` value: a release is
  `quarantine_recomputed{scope_changed}` plus the per-node recomputation, stored as
  `quarantined:false` with `quarantine_reason` absent. The stored reason MUST be
  derived by the platform from this mapping, never copied from an event string.
  `Tests: TestQuarantineReasonIsDerivedFromKindMapping,`
  `TestOperatorReleaseRejectedAsUnknownEnumValue`.

  _A2-1.6 owns this table; it is reproduced here so the A1↔A2 field mapping reads
  the same from either document._
  One value, one name platform-wide (A0-3.6): a graph node is `graph_node_id` and a
  graph edge is `graph_edge_id` in every JSON document, payload, error body and log
  attribute; `node_id` remains the remote agent node (`slp_node_`, Q9). A2's served
  field `id` is renamed accordingly (pre-Freeze, additive-only afterwards).

  | A1 event payload field | A2 graph field (Go / JSON) |
  |---|---|
  | `graph_node_written.graph_node_id` | `Node.ID` / `graph_node_id` |
  | `graph_edge_written.graph_edge_id` | `Edge.ID` / `graph_edge_id` |
  | `from_graph_node_id` | `Edge.SourceID` / `source_id` |
  | `to_graph_node_id` | `Edge.TargetID` / `target_id` |
  | `supersedes_graph_node_id` | `Node.SupersedesID` / `supersedes_id` |

  The originating worker/orchestrator of a graph write is **not** the event's
  `actor`: `graph_node_written.actor` is `(platform, "graph")` and the producer is
  carried by the node's `provenance.principal_kind` (A2-5.3). The two are expected
  to differ — copying one into the other is a defect.
  `Tests: TestFieldNameMappingIsTotal, TestPrincipalKindIsNotCopiedFromActor`.
- **A1-3.7** SPEC §5 step coverage (completeness check, no normative force of
  its own): 1 → `engagement_created`, `chain_genesis`, `scope_changed`,
  `engagement_policy_changed` ·
  2 → `run_started`, `model_config_snapshotted`, `job_spawned`,
  `container_started` · 3 → `spawn_requested`, `task_spawned`,
  `command_executed`, `evidence_stored`, `task_result`, `graph_*` · 4 →
  `llm_call`, `graph_*` · 5 → `approval_requested`, `notification_sent`,
  `scope_denied`, `blacklist_denied`, `action_blocked` · 6 →
  `approval_granted`/`_denied`/`_expired` · 7 → `command_executed`,
  `evidence_stored`, `revert_recorded`, `approval_executed` · 8 →
  `agent_error`, `llm_call`, `graph_*`, `container_killed` · 9 →
  `chain_verified{pre_export}`, `integrity_override`, `artifact_released`,
  `evidence_stored{report_artifact}` · 10 → `cleanup_planned`,
  `cleanup_executed`, `cleanup_verified`, `hard_stop_fired`. Step 1 also owns
  `quarantine_recomputed{scope_changed}`, step 8 `graph_edge_retracted`;
  `engagement_closed` follows step 10 — the engagement ends after its cleanup is
  verified.

- **A1-3.8** A2 cross-contract answers (binding on A2; A2 §6.9 and A2's
  "Cross-contract requests (not A0)"). Each request is confirmed or corrected
  here — A1 owns the envelope and the taxonomy; the additions are listed for
  product-owner confirmation at §6.14:

  | A2 request | A1 answer |
  |---|---|
  | envelope field names `event_id`, `kind`, `recorded_at`, `seq`, `engagement_id`, `run_id`, `job_id` | **Confirmed** — all seven are spelled exactly as A2-5.4 assumes (A1-1.1). A2 MAY also consume `task_id`, `node_id`, `occurred_at`, `occurred_claimed_at` and `actor`; `occurred_claimed_at` is the source of A2-5.3's `observed_claimed_at` and inherits its untrusted status (A0-5.7, A1-1.4) |
  | kind for *node created* | **Confirmed**: `graph_node_written` |
  | kind for *node quarantined* | **Confirmed**: `graph_node_quarantined` |
  | kind for *quarantine recomputed* | **Added**: `quarantine_recomputed` (additive, A0-6.5). Discharges A2-8.5 |
  | kind for *node report-excluded* | **Confirmed**: `report_inclusion_changed` with `included:false`. Discharges A2-8.7 |
  | kind for *node superseded* | **Corrected — no new kind.** A supersession is already recorded twice: `graph_node_written` with a non-empty `supersedes_graph_node_id`, and `graph_edge_written` with `edge_kind:"supersedes"`. A third kind would be a third source of truth for one fact (A2-2.6's own argument against a second encoding). A2 MUST derive supersession from those two payloads (A2-4.1/4.2) |
  | kind for *edge retracted* | **Added**: `graph_edge_retracted` (additive). Discharges A2-3.9 |
  | kind for *write rejected* | **Confirmed**: `action_blocked` with `reason:"graph_write_rejected"` and `action_kind:"graph_write"`. Discharges A2-10.8 — a rejected graph write stores nothing and still leaves a chained trace |
  | ingest dedup key for idempotent replay (A2-4.7) | **Answered by A1-7.6/A1-7.7**: the append dedup key is `(engagement_id, actor.principal_id, kind, idempotency_key)`. The guarantee A2-4.7 needs follows from it — a replayed append returns the **same** `event_id`, so ingest sees one event identity and A2's own `(engagement_id, kind, content_hash)` dedup collapses the replay |
  | field names A1 must not rename | **Confirmed** by A1-3.6; `graph_edge_retracted` reuses `graph_edge_id`, `edge_kind`, `from_graph_node_id`, `to_graph_node_id`, so A2's naming stays single-sourced |

### A1-4 · Payload rules

- **A1-4.1** Each kind has exactly **one** payload type: a flat Go struct with
  an explicit `json:"…"` tag on every field (A0-8.1), whose JSON key set is
  **closed** and listed in A1-3.3. There is no shared payload base type, no
  per-kind `map[string]any`, no embedded "common" object, and no payload field
  that is optional in the sense of *absent* — every listed key is present on
  every instance with its zero value when unset (A0-2.14, A1-1.2). A kind's key
  set MUST NOT be reshaped inside `/api/v1`: no rename, no type change, no
  added or removed field, because the canonical bytes of every written event
  are immutable (A0-6.6, A0-2.16). Evolution happens by **adding a kind**
  (A1-3.5). _One flat type per kind is what makes the fixed key set checkable
  by reflection in the shared suite and the digest reproducible ten releases
  later; a `map[string]any` payload has no key set to fix._

- **A1-4.2** Per-kind obligations beyond "all listed fields present". Each is
  checked at composition in the A1-7.10 order; a violation is `validation`
  (A1-7.5) unless the clause names another kind.

  | Kind | Obligation |
  |---|---|
  | `chain_genesis` | `chain_spec` MUST equal the A1-5.3 constant byte-exactly; exactly one per engagement at `seq` 0 (A1-5.3) |
  | `chain_verified` | `head_seq`/`head_hash` MUST equal the chain-head state at the end of the walk (A1-5.6); `verified_count` = events walked; MUST NOT be written for a chain with a known break (A1-6.2) |
  | `chain_break_detected` | `break_seq` is the **first** break in `seq` order (A1-6.3); `expected_prev_hash`/`actual_prev_hash` MUST both be present, equal when the break is not a link break |
  | `integrity_override` | `reason` MUST be non-empty — an override without a stated reason is `validation` (Q11: never silent); `break_event_id` MUST resolve in this engagement (A1-4.11) |
  | `engagement_created` | composed at `seq` **1** of the chain (A1-5.3 keeps `chain_genesis` at `seq` 0 as the integrity anchor); `roe_evidence_id` and `policy_evidence_id` MUST be non-empty — an engagement whose rules of engagement are not an artifact is not reproducible (SPEC §7); `client_ref` is a customer-chosen label, never a person's name (A0-3.7) |
  | `engagement_closed` | `report_evidence_id` MUST be non-empty when `close_reason:"completed"` and MAY be `""` for `cancelled`/`abandoned`; `close_reason` MUST be one of the three enum values (A1-4.12) |
  | `artifact_released` | composed by the export path for **every** released customer-facing artifact (A1-6.6) and only after that path's `pre_export` walk; `head_seq`/`head_hash` MUST equal the values carried in the artifact's A1-6.6 metadata block (the last row the walk verified, A1-6.6); `integrity_state` is the A1-6.6 export enum, derived per A1-6.6's mapping table; `override_event_id` MUST be non-empty iff `integrity_state:"failed_overridden"` and MUST name the live override that released this artifact (A1-6.5, ADR-0021); `artifact_evidence_id` MUST resolve in this engagement (A1-4.11) |
  | `scope_changed` | `entry_hash` MUST be the A0-2.15 digest of `entry`, so a scope entry is comparable without echoing it into every report (ADR-0005 §2–§3). A **global**-blacklist change MUST also be composed as `scope_changed` into **every affected engagement chain**, with `change_kind:"blacklist_added"`/`"blacklist_removed"` and the `entry`/`entry_hash` of the global entry: the composing subsystem is `scope`, `actor` is the admin user (`(user, "")`, `usr_` id per §6.2) and `run_id` is `""`. Each engagement's `quarantine_recomputed{blacklist_changed}` MUST reference **that engagement's own** `scope_changed` event as `trigger_event_id`. `Tests: TestGlobalBlacklistChangeIsChainedPerEngagement, TestQuarantineRecomputedTriggerResolvesInEngagement` |
  | `run_started` | `llm_data_policy` MUST equal the engagement policy in force at start (ADR-0020 §1); `scope_snapshot_evidence_id` MUST be non-empty — a run without a scope snapshot is not reproducible (SPEC §7) |
  | `job_spawned` / `task_spawned` | `image_digest` MUST be the registry-derived digest, never an orchestrator-supplied string (Q14, ADR-0017 §2); `spawn_request_event_id` MUST be non-empty; `task_spawned.target_graph_node_id` MUST equal the `spawn_requested` value for the same spawn (C-02/T-03: the target is cited by `gn_` id, never by string, A2-8.3) |
  | `spawn_requested` | `image_digest` and `risk_tier` are platform-derived from the registry (Q14) and are `""` only when validation failed before derivation; `task_description` is untrusted and capped (A1-4.4/4.5); `target_graph_node_id` is `""` or a `gn_` id resolving in this engagement — the target of a spawn is **never** taken from a graph field, only cited by id so the quarantine check is on the id (A2-8.3) |
  | `container_killed` | `exit_code` is `-1` when no exit status exists (killed, timeout, node lost). `container_killed{kill_reason:"hard_stop"}` MUST carry `stop_event_id:string` (`evt_`, the `hard_stop_fired` event it answers). The platform MUST commit `hard_stop_fired` **before** it issues any kill, in its own transaction, and MUST NOT block the kill on that commit (A1-7.12): if the commit fails the platform MUST kill anyway and MUST retry the append until it lands. The obligation is correlation by `stop_event_id`, not `seq` order; a `container_killed{hard_stop}` whose `stop_event_id` is empty or unresolvable is a platform defect → `internal` (A0-3.1) and MUST NOT delay the kill. The validator reads the run's hard-stop state held by `internal/policy`, never the chain (A1-7.12). `Tests: TestHardStopKillCarriesStopEventID, TestKillWithoutStopEventIDIsInternal` |
  | `command_executed` | `command` is the exact argv as executed, never a paraphrase (ADR-0009 §1 reproducibility); `output_evidence_id` MUST be non-empty when `output_bytes > 0`; `exit_code` ∈ `[-1, 255]`, `-1` = no exit status; `duration_ms` ≥ 0 |
  | `evidence_stored` | `evidence_id` MUST be the id returned by `evidence:upload` (ADR-0009 §2); `sha256` is the artifact-integrity digest and MUST be non-empty; `size_bytes` > 0 |
  | `task_result` | `status:"failed"` → `error_kind` non-empty and equal to an A0-3.1 kind (A1-4.12); `revert_event_ids` lists the `revert_recorded` events this task produced (ADR-0009 §3) and MAY be empty; `command_count` is the worker's own claim — the platform's count is derivable from the chain and a divergence is reportable, not a rejection |
  | `revert_recorded` | `revertable:false` → `revert_action` MAY be empty, and the effect MUST be carried forward into `cleanup_planned.non_revertable_event_ids` (ADR-0009 §4: non-revertable effects are documented, not dropped) |
  | `approval_requested` | `fingerprint_hash` MUST be the A7 digest of the action (A1-3.6); `action_spec_evidence_id` MUST be non-empty — the platform stores the exact canonical bytes (A0-2) of the A7 action spec as a write-once evidence artifact at request time and records its id here and in `evidence_refs` (A1-7.3); `argv_hash` MUST be the A0-2.15 digest of the exact argv the spec will produce; `expires_at` is **platform-computed** at request time from the engagement's timeout (ADR-0012 §7) and MUST NOT be a caller value; `untrusted_context` is **platform-computed** (A1-4.4) and MUST NOT be a caller value; `action_summary` is prose *in addition to* the fingerprint, never instead of it (ADR-0018 §1), and MUST be composed only from platform vocabulary and the A7 action spec's own fields — copying model or tool prose into it is a laundering violation of A1-4.4. `Tests: TestApprovalRequestStoresActionSpecArtifact, TestUntrustedContextComputedNotSupplied` |
  | `approval_granted` / `_denied` / `_expired` / `_executed` | `fingerprint_hash` MUST equal the one on the `approval_requested` event with the same `approval_id` — the platform MUST refuse to compose a decision event whose fingerprint diverged (ADR-0018 §2); `single_use` is `true` for every v1 approval (Q10) |
  | `approval_executed` | `revalidated` records the execution-time re-validation outcome and MUST be `true` for a successful execution (ADR-0018 §2–§3); `single_use_consumed` is the Q10 consumption record; `action_spec_evidence_id` MUST equal the `approval_requested` value for the same `approval_id`, and `fingerprint_hash` MUST be the A0-2.15 digest of **exactly those artifact bytes** (A1-3.6, C-02); **`expires_at` on `approval_granted`, `approval_expired` and `approval_executed` MUST be byte-equal to the `approval_requested` value for the same `approval_id`** — divergence is a platform defect → `internal`, the execution is aborted, and `action_blocked{reason:"approval_metadata_mismatch"}` is recorded (a timeout-policy change MUST NOT affect an already-requested approval, ADR-0012 §7). A fingerprint that diverged at execution MUST be refused with **`conflict`** (409) naming `approval_id` and the first 8 hex characters of both digests, and MUST record `action_blocked{fingerprint_mismatch}`. The platform MUST consume the approval and authorize the execution **in one transaction** (or under the same per-engagement lock as A1-5.4's append), guarded by a uniqueness constraint on `(engagement_id, approval_id)` in the consumption table; a second attempt MUST fail **before** any container is created, with `conflict` (A0-3.1) and `action_blocked{reason:"approval_consumed"}`, and MUST NOT compose a second `approval_executed`. `Tests: TestApprovalFingerprintAndExpiryAreEqual, TestApprovalMetadataMismatchIsConflict, TestActionSpecArtifactIsChained, TestSingleUseApprovalRaceConsumesOnce` (N concurrent executions → exactly one `approval_executed`, N−1 `action_blocked{approval_consumed}`), `TestConsumedApprovalCannotSpawnAgain` |
  | `llm_call` | metadata only (ADR-0020 §2): the platform MUST NOT compose this kind with prompt or completion text in any field; `endpoint_name`/`model_name` are configured names, never a URL or a credential (A0-3.7); `status:"error"` → `error_kind` non-empty |
  | `graph_node_written` / `graph_edge_written` | `source_event_id` MUST be non-empty — provenance is mandatory on the graph side (A2-5.1/5.4) and the event that carries it is its anchor. `content_hash` MUST equal the A2-4.6 fingerprint of the written node (edges have no fingerprint — A2-4.6 is node-only); `dedup_hit` is `true` when the write collapsed into an existing row under A2-4.7, and a collapse MUST still emit the event. `node_kind` MUST be an A2-2.1 kind and `edge_kind` an A2-3.1 kind; because only the platform composes these kinds after a successful graph write, a violation is a platform defect → `internal`. `Tests: TestContentHashMatchesGraphFingerprint, TestDedupCollapseStillEmitsEvent` |
  | `graph_node_quarantined` | `quarantine_kind:"blacklist_match"` is not releasable (A2-8.5); `operator_quarantine` MUST carry a non-empty `reason`. There is no `operator_release` value (A1-3.6) |
  | `quarantine_recomputed` | `trigger_event_id` MUST resolve, in this engagement, to the event that caused the recomputation: the `scope_changed` event for `scope_changed`/`blacklist_changed` (A2-8.5, incl. the per-engagement copy of a global-blacklist change), or the `graph_node_written`/`graph_edge_written` event for `node_written`/`edge_written` (A2-8.2) |
  | `graph_edge_retracted` | `graph_edge_id` MUST resolve in this engagement; `reason` MUST be non-empty when the retraction is operator-initiated (A2-3.9) |
  | `cleanup_planned` | `approval_id` MUST be non-empty — a cleanup plan executes only after human approval (ADR-0009 §4) |
  | `cleanup_executed` / `cleanup_verified` | `revert_event_id` MUST resolve to a `revert_recorded` event; `status:"failed"` → `detail` non-empty |
  | `scope_denied` / `blacklist_denied` | `attempted_target` is the target **as requested** (untrusted, never normalized into scope vocabulary); `blacklist_entry` on `blacklist_denied` MUST be the matched entry, and blacklist beats allowlist beats approval (ADR-0005 §3, C5) |
  | `action_blocked` | `reason` and `action_kind` MUST both be set; `detail` carries the human prose (A0-3.4) and is never parsed |
  | `agent_error` | `error_kind` is an A0-3.1 kind; `origin` is `component.Function` per ADR-0019 §2–§3; `message` is already redacted (A0-3.7, A1-4.9) |
  | `notification_sent` | `notification_kind` is the ADR-0012 §2 vocabulary, not an A1 kind (A1-3.3); `target_name` is the configured channel name, never a webhook URL (A0-3.7); `attempt` ≥ 1; for `chain_head_anchor`, `related_event_id` names the event whose append crossed the emission point (or the `chain_verified` event) |

  Every `error_kind` field (`task_result`, `agent_error`, `llm_call`) MUST be
  `""` or a byte-exact A0-3.1 kind; the per-kind non-empty obligations above are
  additional; an unknown value is `validation` (A0-6.3).

- **A1-4.3** A payload carries **references, never content**: `evi_` ids for
  artifacts (A1-1.7, ADR-0009 §2) and `evt_` ids for other events (A1-4.11).
  A payload MUST NOT contain artifact bytes, a base64 blob of an artifact, raw
  tool output, a model prompt or completion, a screenshot, a packet capture or
  a file body — not truncated, not encoded, not "just the first line". Bulk
  content belongs in the evidence store behind `evidence:upload`
  (ADR-0009 §2). _Two reasons, both structural: a hashed row must stay small
  enough to re-verify a whole chain at startup (A1-6.1), and every byte of
  captured output is a candidate secret carrier (A1-4.9, ADR-0020 §4) whose
  proper home is the write-once evidence store, not the audit spine._

- **A1-4.4** Untrusted content (SPEC §6, adversarial A1). The `*` marker in
  A1-3.3 names a field whose **content originates outside the platform's own
  vocabulary**: tool output, target-supplied data (hostnames, banners, DNS
  answers), model prose, client prose, or free text a human typed into a
  dialog. The envelope's `untrusted` boolean is **platform-computed** and MUST
  be `true` iff at least one `*`-marked field of that kind's payload is
  non-empty (non-`""`, non-`[]`), `false` otherwise; it MUST NOT be accepted
  from a client (A1-2.3) and MUST be deterministic, because it is inside the
  digest (A1-5.2).
  - Consumers MUST treat a `*` value, and any value on an event with
    `untrusted:true`, as hostile input: rendered only through
    `html/template` autoescaping (C2), flagged in approval views
    (ADR-0018 §4), passed to a model only through the gateway with the
    engagement's policy applied (ADR-0020 §1–§2, C11), and **never**
    interpreted as configuration, a path, a command, an argv element, a scope
    or blacklist entry, an id, an enum value or a control instruction
    (SPEC §6, A2-6.7).
  - `untrusted:false` is **not** a licence to interpret a string: no event
    field of any kind is ever configuration. The flag narrows *provenance* for
    approval views and model-input handling; it never widens trust.
  - A platform subsystem MUST NOT launder untrusted content into an unmarked
    field (copying a target-supplied hostname into `entry`, a model sentence
    into `origin`, tool text into `model_name`): an unmarked field is a claim
    that the value is platform vocabulary, and the contract test corpus
    asserts that claim per kind.
  - Free text a human typed (`integrity_override.reason`,
    `report_inclusion_changed.reason`, `hard_stop_fired.reason`) is untrusted
    **for rendering** even where A1-3.3 does not mark it: the UI MUST
    autoescape every string field of every event. _A stored-XSS payload in an
    override reason is the same bug as one in a tool banner, and the platform
    has no interest in distinguishing them._
  - `approval_requested.untrusted_context` is **platform-computed** on the same
    rule and MUST NOT be a caller value: it MUST be `true` iff this payload has
    any non-empty `*`-marked field, **or** iff the event referenced by
    `request_event_id`/`spawn_request_event_id` (transitively) carries
    `untrusted:true`. An approval whose context descends from injected prose is
    flagged even when the action spec itself is pure platform vocabulary
    (ADR-0018 §4).
  - `action_summary` MUST be composed only from platform vocabulary and the A7
    action spec's own fields; copying model or tool prose into it is a
    laundering violation of this clause.
  - The `*`-marked fields per kind, transcribed from A1-3.3 (normative: this
    table is what `UntrustedFields` returns, §4.1):

    | Kind | `*` fields |
    |---|---|
    | `integrity_override` | `reason` |
    | `run_ended` | `detail` |
    | `hard_stop_fired` | `reason` |
    | `spawn_requested` | `task_description` |
    | `command_executed` | `command`, `target` |
    | `task_result` | `result_summary` |
    | `revert_recorded` | `target`, `revert_action` |
    | `approval_requested` | `target`, `action_summary` |
    | `approval_denied` | `reason` |
    | `approval_executed` | `target`, `action_summary` |
    | `graph_node_quarantined` | `reason` |
    | `graph_edge_retracted` | `reason` |
    | `report_inclusion_changed` | `reason` |
    | `cleanup_executed` | `detail` |
    | `cleanup_verified` | `detail` |
    | `scope_denied` | `attempted_target` |
    | `blacklist_denied` | `attempted_target` |
    | `action_blocked` | `attempted_target`, `detail` |
    | `agent_error` | `message` |
    | every other kind (23 of 42) | none |

  `Tests: TestUntrustedFlagMatchesStarredFields, TestNoUntrustedTextInUnmarkedFields`
  (per-kind source→target allowlist table), `TestApprovalViewFlagsUntrustedContext`,
  `TestUntrustedFlagCannotBeSupplied`.

- **A1-4.5** A0-7.7 obligation discharged: **every** capped A1 field class, its
  cap and its mechanism. A1 assigns **mechanism R (reject, A0-7.6) to every
  class and mechanism T (truncate, A0-7.5) to none**.

  | Field class | Cap | Fields | Mech |
  |---|---|---|---|
  | Prose, long — untrusted `*` fields | 2048 B | `command`, `task_description`, `result_summary`, `revert_action` | **R** |
  | Prose, medium — `*` fields plus operator-typed `entry` | 512 B | `reason` (all kinds), `detail`, `action_summary`, `message`, `entry` | **R** |
  | Target text — untrusted `*` fields | 256 B | `target`, `attempted_target`, `blacklist_entry` | **R** |
  | Platform/config labels | 128 B | `origin`, `container_ref`, `network_name`, `media_type`, `model_name`, `endpoint_name`, `target_name`, `old_value`, `new_value`, `client_ref`, `recipient_ref` | **R** |
  | Registry-derived strings | 64 B / 32 B / 256 B | `tool_version` (64), `node_kind`/`edge_kind`/`risk_tier` (32), `image_digest` (256) | **R** |
  | Identifiers | A0-1.2 total length | every `*_id` field | regex **reject** → `validation` (A0-1.5), not a byte cap |
  | Digests | 64 chars | `*_hash`, `sha256` | A0-2.15/8.7 **reject** → `validation` |
  | Closed enums | A0-8.5's 32 chars | every `enum{…}` field, `kind`, `error_kind` | **reject** → `validation` (A0-6.3) |
  | Arrays | counts below | A1-4.7 | **R** |
  | Canonical event document | 32768 B | the whole event | platform invariant, A1-4.5 note |

  Rationale for R everywhere: every A1 cap guards client-, agent- or
  target-originated text that becomes evidence, and mechanism T would silently
  shorten an audit record — the exact failure A0-7.6 exists to prevent
  (ADR-0016 §1, ADR-0018 §1). Two structural reasons reinforce it: T requires
  a sibling `<field>_truncated` boolean (A0-7.5) that A1-3.3's closed key sets
  do not declare, so T would reshape every kind's payload (A1-4.1 forbids it);
  and the unbounded thing an agent actually produces — tool output — is not in
  the payload at all, it is an `evi_` reference (A1-4.3), so rejecting an
  over-cap summary costs a retry, never evidence. A0-7.3 measurement
  (decoded UTF-8 bytes of the value) and A0-7.8 (enforcement at platform
  ingest, never in the producer) apply unchanged. **Decided** (product owner,
  2026-09-21, PR #2 §6.3).
  The **values** of every cap above live in the A0-7.1 registry (AM-2):
  A1 cites the registry names — `ProseLongMaxBytes`, `ProseMediumMaxBytes`,
  `TargetMaxBytes`, `LabelMaxBytes`, `ToolVersionMaxBytes`, `KindNameMaxBytes`,
  `DigestMaxBytes`, `EvidenceRefsMax`, `EventRefsMax`, `ExitCodeMin`,
  `ExitCodeMax`, `EventMaxCanonicalBytes`, `IdempotencyKeyMaxBytes` — and declares
  no constant of its own (§4.1); this table is the mechanism assignment A0-7.7
  requires of the owning contract.
  A kind's **maximal payload** sets every string field to exactly its cap length
  in bytes of U+0001 (worst case `\u0001` = 6 canonical bytes per input byte,
  A0-2.7), every integer to its declared maximum, every array to its count cap
  filled with maximum-length ids, every bool to `true`.
  `TestMaximalPayloadFitsCanonicalBound` asserts
  `len(Preimage(e)) ≤ EventMaxCanonicalBytes` for all **42** kinds under that
  construction. `Tests: TestMaximalPayloadFitsCanonicalBound,`
  `TestCapsRejectWithSummaryTooLarge`.
  _Note on the last row: with every field capped, the canonical bytes of any
  event are bounded by 6 × the largest decoded prose cap (A0-2.7's worst-case
  `\u00xx` expansion of a 2048 B control-character string = 12288 B) plus
  envelope overhead, i.e. < 32768 B. The platform MUST assert this bound at
  composition; exceeding it means a cap was not enforced and is a platform
  defect → `internal` (A0-3.1), never a client error. The shared suite asserts
  that a maximal payload for every kind fits._

- **A1-4.6** The `redacted` marker (`command_executed`, `evidence_stored`) MUST
  exist from day one even though the masking/redaction scanner is not designed
  yet (A1 §2, security-hardening session). `Tests: TestRedactedIsPlatformSetOnly,
  TestRedactedNotUsedForTruncation`. Semantics: `redacted:true` means a
  **platform** redaction or masking step (ADR-0020 §3–§4) altered or withheld
  part of the referenced artifact or of the metadata recorded about it. It is
  platform-set only (A1-2.3), it is inside the digest, and it MUST NOT be used
  to mean "truncated" (that is A0-7.5's `*_truncated`, which A1 does not use —
  A1-4.5) or "a secret was stripped from this payload" (an append carrying a
  secret is rejected outright, A1-4.9). `redacted:false` is a value, not an
  absence (A0-8.3): it asserts that no redaction step fired, which is only
  meaningful once the scanner exists, so until then the platform MUST write
  `false` and MUST NOT claim a scan happened. _A marker that has to be added
  later cannot be added later: the key set is frozen at Freeze (A1-4.1) and
  every historical row would keep the old digest._

- **A1-4.7** Arrays. Every A1 payload array is an ordered array of strings
  (`array[string]` in A1-3.2); the platform MUST sort it
  **ascending by unsigned byte value** (A0-2.4's order, `COLLATE "C"` for
  stored text — A0-1.9) and MUST remove duplicates **at composition**, so two
  producers listing the same set yield identical canonical bytes. Clients MUST
  NOT rely on an input order being preserved: an append whose array is
  unsorted is normalized, not rejected. Count caps (mechanism R):
  `evidence_refs` ≤ `EvidenceRefsMax` (A0-7.1 registry)
  · `revert_event_ids` and `non_revertable_event_ids` ≤ `EventRefsMax`
  (A0-7.1 registry). Exceeding a count → `summary_too_large` (A0-7.6) naming
  the field, the cap and the actual count. _Sorting at composition is a
  determinism rule, not a convenience: `evidence_refs` is inside the digest
  (A1-1.1), so producer-dependent order would make the same logical event hash
  differently on two code paths._
  `Tests: TestArraysSortedDedupedAtComposition, TestEvidenceRefsDerivation`
  (one subtest per kind, incl. `evidence_stored`).

- **A1-4.8** Flatness (A1-1.5). A payload is a JSON object whose values are
  strings, integers, booleans, or arrays of strings/integers. Nested objects,
  arrays of objects, `null` (A0-8.3) and floats (A0-2.6) MUST NOT appear —
  `validation`. Document depth is therefore ≤ 3 (envelope → `payload`/`actor`
  → value), far inside A0-2.11's 32. A structure that seems to need nesting is
  either several flat fields (`from_graph_node_id`, `to_graph_node_id`) or
  belongs in the evidence store (A1-4.3). _Flat is what lets the shared suite
  check a kind's key set and value types by reflection over one struct, and
  what keeps a hashed document readable by a human auditor in a report._
  Every slice and map field of a canonicalized type MUST be non-nil before
  marshaling; the constructors initialize them to empty (A0-2.14). A canonical
  event document containing `null` is a platform defect → `internal`.
  `Tests: TestPayloadsAreFlat, TestNilCollectionNeverSerializesAsNull`.

- **A1-4.9** No secret values (ADR-0019 §5, ADR-0020 §3–§4, A0-3.7). No event
  field — payload, `actor`, `evidence_refs`, or a value inside an array — MAY
  contain secret material: a plaintext credential, a captured
  password/NTLM/Kerberos hash, a private key, a ticket, a token, a session
  cookie, a TOTP seed, a webhook URL or model-endpoint URL with embedded
  credentials, a per-run masking-mapping value or pseudonymization-table entry
  (ADR-0020 §3), or a raw model/tool payload. Events carry `evi_` references to
  captured material (A1-4.3) and opaque labels for it — the graph-side rule is
  identical and is A2-9; A1 and A2 MUST be read together, and A2-9.6 is the
  worker-side procedure (upload, reference, never inline).
  - Ingest scanning: the platform MUST run the `internal/secretscan` rules of
    A2-9.4 over every string field and every array element of an append payload,
    and MUST reject a match with `validation` naming the **field** and the
    **rule id** — never echoing the value, a prefix of it, or its digest
    (A2-9.5). The rejection stays observable as
    `action_blocked{reason:"append_rejected"}` (A1-7.5). Reject-vs-redact is
    **ruled**: reject, never redact (§6.4).

    The pattern set is A2-9.4's rule table and nothing else: rule ids live in exactly
    one document, and an error MUST name the field and the rule id, never the value, a
    prefix of it, or a digest of it (A0-3.4, A2-9.5). The scan is a **filter, not a
    guarantee** — a hostile worker can encode, split or re-format a secret past any
    pattern set. Egress exclusion (ADR-0020 §4) is the enforcement point and MUST be
    applied independently at the gateway to every string that leaves the platform; for an
    engagement whose policy is not `local_only` the gateway MUST exclude by **kind** (no
    `credential` node, no node with `credential_kind` set, no `attrs` of such a node) and
    `llm_call` MUST record the exclusion in `excluded_secret_count`.
  - `evidence_stored.sha256` is the **artifact-integrity** digest required by
    ADR-0009 §2 and is not "a captured hash" in A0-3.7's sense; a digest *of a
    credential value* MUST NOT be stored in any event field.
  - Cloud egress: the gateway's exclusion (ADR-0020 §4) is the enforcement
    point and is recorded as `llm_call.excluded_secret_count`; the scan above is
    a filter, not a guarantee, so no read path, stage view (A3),
    report or SSE frame may be relied on to be secret-free by construction.
  - Negative tests (shared suite, "secret-free serialization"):
    `TestEventSecretFreeSerialization` — for a corpus of appends with planted
    secrets (an NTLM-shaped hash, a PEM private key, a cloud access-key id, a
    bearer token, a `https://user:pass@…` webhook URL) in every `*` field of
    representative kinds: the append is rejected, **and** no planted value
    appears in the stored preimage bytes, the served envelope, the error body,
    the `action_blocked` event, or the captured `slog` output.
    `TestNoMaskingMappingInEvents` — no event field contains an ADR-0020 §3
    mapping value or its inverse. `TestEventCarriesReferencesOnly` — over the
    same corpus, no payload field carries artifact bytes (no base64 blob, no
    embedded file body) and every `evi_` reference resolves to an uploaded
    artifact (A1-4.3).
    `Tests: TestEventSecretFreeSerialization, TestSecretScanNamesFieldNotValue,`
    `TestNoSecretValueOrDigestInError`.

- **A1-4.10** Unknown fields, both directions (A0-6.1/6.2/6.3, A0-6.6).
  `Tests: TestServedEventIgnoresUnknownFields, TestAppendRejectsUnknownField`.
  - **Write** (`events:append` and every platform composition path): an
    unknown key in the request body or inside `payload` MUST be rejected with
    `validation` naming the offending field (`json.Decoder
    .DisallowUnknownFields`, A0-6.2). The closed key set of A1-4.1 is what
    "unknown" means; there is no lenient mode for internal callers
    (A1-7.10).
  - **Read**: a client decoding a served event MUST ignore unknown fields and
    MUST NOT enable strict decoding on responses (A0-6.1); an unknown `kind`
    MUST be preserved verbatim and the item skipped without aborting the page
    (A0-6.3). Tolerance on read is **not** a reason to relax the write side —
    the asymmetry is the point (A0-6 table).
  - **Evolution**: because an event is canonicalized and hashed, adding a
    field to an existing kind's payload changes that kind's canonical bytes and
    is therefore **breaking** (A0-6.5/6.6) — it requires `/api/v2` plus an ADR,
    and it would invalidate verification of every historical event of that
    kind. The 17-key envelope (A1-1.1) is likewise frozen for `/api/v1`: an
    18th envelope key is a v2 change, not an additive one. New information is
    carried by a **new kind** (A1-3.5) or by a new evidence artifact
    (A1-4.3). _This is the second-order consequence of hashing a document: the
    usual "just add an optional field" escape hatch does not exist here, so the
    taxonomy has to be complete on day one and additions have to be kinds._

- **A1-4.11** Identifiers and event references inside a payload. Every `*_id`
  field carries an A0-1.2 identifier and MUST be validated byte-exact per
  A0-1.5 at composition (no Crockford normalization, no case folding). `""` is
  the zero value meaning "not applicable" (A0-8.4) and is legal exactly where
  A1-3.3/A1-4.2 do not require a value. A `*_event_id` field MUST reference an
  event **in the same engagement** (C8, A1-8.6): a well-formed `evt_` that
  does not resolve in this engagement is `notfound` (404) when it came from a
  client, never `forbidden` (A0-3.9, adversarial A12), and `internal` when the
  platform derived it — a platform subsystem that cannot resolve its own
  correlation id has a defect. The same rule applies to `evi_`, `gn_`, `ge_`,
  `apr_` and `tool_` references. Cross-engagement references are not
  expressible in a stored event: the check runs before the row is written
  (A1-7.10).
  A client MUST upload its artifacts (`evidence:upload`) **before** appending the
  event that references them; an append whose `*_evidence_id` does not resolve in
  this engagement is `notfound` (404) and MUST be retried by the buffering client
  after the upload (ADR-0013). Write-time resolution is what makes A1-8.8's
  dangling reference an archival state, never a normal one.
  `Tests: TestEvidenceIDResolvesAtWriteTime, TestDanglingEvidenceIDIsNotFound`.

- **A1-4.12** Enums, timestamps and integers inside a payload.
  - **Enums** spelled `enum{…}` in A1-3.3 are closed lists owned by A1
    (A0-8.5 spelling, byte-exact comparison, no synonyms). An unknown value on
    a write → `validation` (A0-6.3); on a read a client MUST preserve the raw
    string and skip what it cannot handle. Five payload fields across three
    groups are **not** A1 enums: `node_kind`/`edge_kind` are A2's closed lists,
    stored here as capped strings so an A2 addition needs no A1 change
    (A1-3.6); `error_kind` is A0-3.1's closed list, stored as a string for the
    same reason; `risk_tier` and `tool_version` are A7/registry vocabulary
    (A1-3.6).
  - **Timestamps**: a `timestamp` field is a Go **`string`** holding an A0-5.1
    value, in the envelope and in every payload. `time.Time` MUST NOT appear in
    any canonicalized type: its `MarshalJSON` drops trailing zero fractional
    digits and silently changes digests. Payload timestamps (`expires_at`)
    follow A0-5.1 exactly (three fractional digits, `Z`) and are
    platform-computed (A1-4.2): a caller MUST NOT supply one, and no payload
    timestamp is ever a `*_claimed_at` field — the only client-supplied time in
    an event is the envelope's `occurred_claimed_at` (A1-1.4, A0-5.7).
    `Tests: TestNoTimeTimeInCanonicalizedTypes, TestTimestampFieldsAreStrings`.
  - **Integers** are within A0-2.6's `[-(2^53-1), 2^53-1]`, with the suffix
    semantics of A0-8.2 (`*_ms`, `*_bytes`) and the sign rules of A1-4.2
    (`exit_code` may be `-1`; counts, sizes and durations may not be
    negative). Transcription rule for `:int` — **`int64`** for every `*_ms`,
    `*_bytes`, count and `seq` field; **`int`** only where A1-4.2 declares a
    range (`exit_code`). A float in a payload is `validation` (A1-4.8).
    Upstream-reported counters (`llm_call.prompt_tokens`, `completion_tokens`)
    are integers of untrusted *origin* but are not prose, so A1-4.4's marking
    does not apply to them; a consumer MUST NOT treat them as billing truth.

### A1-5 · Hash chain

- **A1-5.1** Chain scope is the **engagement** (Q11): exactly one chain per
  engagement, identified by `engagement_id`, containing every event of every
  run, job, task and node of that engagement in one `seq` space. There is no
  per-run chain, no per-day chain, no global chain, and no chain that spans
  engagements (C8, ADR-0016 §3). _Per-engagement is the only scope that makes
  the chain a customer-deliverable artifact: a report is per engagement, so the
  integrity statement attached to it must be too. Second-order consequence: a
  run-scoped read (A1-8.3) is **not** a contiguous `seq` range and cannot be
  verified on its own — verification is always whole-chain (A1-6.1), and a
  run-scoped integrity claim is a claim about the engagement chain filtered,
  never about a sub-chain._

- **A1-5.2** The digest input is the event document minus a **fixed exclusion
  list of exactly three top-level names** (A0-2.12): `hash`, `prev_hash`,
  `seq`. The **chainable field set** is therefore exactly these 14 keys, and
  nothing else is hashed:

  | # | Chainable key | Notes |
  |---|---|---|
  | 1 | `actor` | the whole object, canonicalized recursively: `component`, `principal_id`, `type` (A0-2.4) |
  | 2 | `engagement_id` | binds the event to its chain — the anti-splicing property (A1-5.9) |
  | 3 | `event_id` | |
  | 4 | `evidence_refs` | sorted, deduplicated (A1-4.7) |
  | 5 | `job_id` | `""` when not applicable |
  | 6 | `kind` | |
  | 7 | `node_id` | remote agent node (Q9) |
  | 8 | `occurred_at` | platform stamp (A1-1.4) |
  | 9 | `occurred_claimed_at` | untrusted, but hashed: a claim cannot be edited afterwards |
  | 10 | `payload` | the whole per-kind object, canonicalized recursively |
  | 11 | `recorded_at` | authoritative ingest time (A0-5.6) |
  | 12 | `run_id` | `""` when not run-scoped |
  | 13 | `task_id` | `""` when not applicable |
  | 14 | `untrusted` | platform-computed (A1-4.4) |

  The hash input is the **canonical JSON** of that document (A0-2 in full:
  UTF-8 byte key order, no whitespace, integers only, literal UTF-8,
  duplicate-key rejection, `internal/cjson` as the single implementation) and
  the digest is **SHA-256** written as 64 lowercase hex characters (A0-2.15,
  FIPS 180-4). Because A0-2.12 exclusion lists accept plain top-level names
  only, `seq`/`prev_hash`/`hash` are top-level envelope keys (A1-1.3) and no
  payload field is excludable — every payload byte is evidence and is hashed.
  **Decided** (product owner, 2026-09-21, PR #2 §6.1).
  The exclusion list is part of the digest's definition and MUST NOT change
  after Freeze except by a new ADR (A0-2.13, A1-5.10). _Excluding `seq` is what
  lets one logical event be re-verified without knowing its position;
  excluding `prev_hash` and `hash` is what makes the chain computable at all.
  Everything else stays in, including the untrusted claimed time: an auditor's
  question "was this timestamp edited later?" must have a cryptographic
  answer._

- **A1-5.3** Genesis. The first event of every chain is a `chain_genesis`
  event at `seq` 0, composed by the platform (`actor.type:"platform"`,
  `actor.component:"event_store"`) **in the same transaction that creates the
  engagement**, with: `run_id`, `job_id`, `task_id`, `node_id`,
  `occurred_claimed_at` all `""`; `evidence_refs` `[]`; `untrusted` `false`;
  `occurred_at` = `recorded_at` = the engagement-creation instant;
  `prev_hash` = `ChainZero` (A1-5.5); payload
  `{"chain_spec":"sleipnir/chain/v1"}` — the fixed 17-byte chain
  specification constant naming this digest definition (SHA-256 over A0-2
  canonical bytes with the A1-5.2 exclusion list), so a future chain format is
  explicit per chain instead of inferred from a release version. Any other
  `chain_spec` value is a platform defect → `validation` at composition,
  surfaced as `internal` to the caller (A0-3.1, A1-2.4). An engagement whose
  chain has no valid genesis MUST NOT accept appends (A1-7.11) and MUST fail
  verification with `genesis_invalid` (A1-6.3). Exactly one genesis per chain:
  a second `chain_genesis` event is a platform defect and a break
  (`duplicate_event_id` if it reuses an id, `genesis_invalid` if it does not).
  `Tests: TestSecondGenesisIsRejected, TestAppendRefusedWithoutValidGenesis`.

- **A1-5.4** `seq` and ordering under concurrent writes. `seq` is a
  per-engagement integer, `0` at genesis, strictly increasing by exactly `1`,
  assigned by the platform **inside the transaction that inserts the row** —
  never before it, never outside it, never by the client (A1-2.3). The store
  seam MUST serialize appends per engagement: two acceptable implementations
  are a `SELECT … FOR UPDATE` on the chain-head row (A1-5.6) or a
  transaction-scoped advisory lock keyed on the engagement id
  (`pg_advisory_xact_lock`); the exact primitive is the store seam's (A1-7.9).
  Consequences, all normative:
  - `seq` order **is** commit order **is** chain order (A1-8.1). Two concurrent
    appends produce two rows with consecutive `seq`, never the same `seq`,
    never a fork, never an interleaved partial chain.
  - A global sequence shared across engagements MUST NOT be used: a rolled-back
    or failed append would burn a value and leave a `seq` gap in one
    engagement's chain, which verification reports as a break (A1-6.3) — a
    false positive caused by an unrelated failure.
  - A failed append MUST NOT consume a `seq` and MUST NOT leave a row; the
    chain is written by one transaction or not at all (A1-7.8).
    `Tests: TestFailedAppendConsumesNoSeq, TestNoGlobalSequenceSharedAcrossEngagements`.
  - `recorded_at` MUST be read from the injected clock (A0-5.4) **after** the
    per-engagement lock is taken, and MUST be clamped forward to
    `max(clock_now, prev_recorded_at)`, so `recorded_at` is non-decreasing
    along `seq` by construction and a backwards host clock step can never
    produce a time regression in a chain. The forward clamp compares A0-5.1
    strings **byte-wise** — for a fixed-format UTC millisecond timestamp, byte
    order equals time order — and is bounded to 1000 ms by A0-5.4: beyond that
    bound the platform uses the true clock reading, logs at error level with
    the engagement/run correlation attributes (ADR-0019 §3), and `recorded_at`
    MAY then be non-monotone; A1-8.1's rule that only `seq` orders anything is
    the compensating control. The clamp input is `ChainHead.LastRecordedAt`
    (A1-5.6). _A monotone `recorded_at` is what makes "the log shows X before Y"
    a statement an auditor can rely on; the bound is what keeps a stuck clock
    from being laundered into a plausible timeline (A0-5.4)._
    `Tests: TestRecordedAtClampIsMonotone, TestRecordedAtClampBeyondBoundIsLogged`.
  - `occurred_at` and `occurred_claimed_at` are **not** clamped and are not
    monotone: they describe the occurrence, not the commit (A1-1.4).

- **A1-5.5** Linking formula. With `preimage(e)` = the A0-2 canonical bytes of
  event `e` under the A1-5.2 exclusion list, `SHA256Hex` = A0-2.15's 64-char
  lowercase hex, and `ChainZero` = 64 ASCII `0` characters:

  ```
  preimage(e)     = cjson.Canonical(e, "hash", "prev_hash", "seq")
  hash(e)         = SHA256Hex(preimage(e))
  prev_hash(e@N)  = hash(e@(N-1))      for N > 0
  prev_hash(e@0)  = ChainZero          (genesis, A1-5.3)
  ```

  Each event's `hash` covers its own content only; the link to its predecessor
  is carried by `prev_hash`, which is **not** inside its own preimage (A1-5.2)
  — the standard chain construction, and the reason a verifier walks in `seq`
  order and can stop at the first break (A1-6.3). Digest comparisons that gate
  an action (verification, export metadata) MUST use
  `crypto/subtle.ConstantTimeCompare` on decoded bytes (A0-2.15). `ChainZero`
  is a constant, not a digest of anything: a genesis whose `prev_hash` is
  anything else is `genesis_invalid`.

- **A1-5.6** The platform MUST keep per-engagement **chain-head state** outside
  the event rows: `head_seq`, `head_hash`, `event_count`, `integrity_state`
  (`unverified` · `verified` · `failed` · `overridden`, A1-6.4),
  `verified_at`, `verified_head_hash`, and the chain's `chain_spec`, plus the
  four bookkeeping fields the append transaction writes and reads:
  `LastRecordedAt string` (A1-5.4's clamp input), `LastBreakSeq int64`,
  `LastBreakKind BreakKind` and `LastBreakEventID string` (A1-6.3's break-dedup
  state). It is updated in the same transaction as every append (A1-5.4) and
  every verification (A1-6). This state is **mutable platform bookkeeping, not
  event content**: it MUST NOT appear in an event document (A1-1.6) and it is
  never hashed. The **row** is internal; the **projection**
  `(head_seq, head_hash, integrity_state, verified_at)` is served to authorized
  readers and to every export, with a wire shape A4 owns (A1-8.9, A1-6.4). Its
  columns are the store seam's business (A1-7.9); its existence and semantics
  are A1's. _Without a head row, every append would have to scan for the
  maximum `seq` and every verification would have no cheap starting point;
  with it, `row_count_mismatch` (A1-6.3) becomes a one-row check._
  `Tests: TestChainHeadProjectionIsServedRowIsNot, TestChainHeadFieldsNeverHashed`.

- **A1-5.7** Stored preimage bytes (A0-2.16 — the obligation A0 assigns to A1).
  The store MUST persist, per event row: the **exact canonical preimage bytes**
  `hash` was computed over (a dedicated `bytea`/text column), plus `seq`,
  `prev_hash` and `hash` as separate indexed columns. Verification MUST
  recompute SHA-256 from those stored bytes and MUST NOT re-serialize a decoded
  struct, a served body or a re-marshaled value (A0-2.16, A1-1.8). The platform
  MUST NOT store a second copy of the envelope JSON: the served representation
  (A1-1.8) MUST be produced by decoding the stored preimage, adding the three
  chain fields with their stored values, and re-canonicalizing (A0-2) — one
  source of bytes, no possibility of the stored copy and the served copy
  drifting. `Served(preimage, c)` =
  `cjson.With(preimage, map[string]any{"seq": c.Seq, "prev_hash": c.PrevHash,`
  `"hash": c.Hash})` (A0 §4), operating on the generic document only;
  it MUST decode with `UseNumber()` and re-emit every number's literal text
  verbatim (A0-2.5, F-01), and MUST reject a preimage that fails A0-2 with
  `preimage_mismatch` (A1-6.3). U+2028/U+2029 are literal in the served bytes
  (A0-2.7) — see §4.3 row 3. _A later Go version, a new struct field or a
  different encoder setting would otherwise silently invalidate every
  historical hash, and the failure would look like tampering._
  `Tests: TestServedBytesReproducible, TestServedRoundTripIsIdentity`.

- **A1-5.8** What the chain proves, and what it does not (declared limits,
  not defects):
  - **Proves:** order and immutability **within** one engagement chain — any
    edit, deletion, insertion, reordering or re-labelling of a stored event
    changes bytes that a verifier recomputes, and any consistent-looking chain
    that was not built by appending in `seq` order fails the link check.
  - **Does not prove wall-clock truth:** the platform clock is the only time
    anchor in v1 (A0-5.6/5.8, Q11: no external timestamp anchoring). Whoever
    controls the platform host controls `recorded_at`.
  - **Does not prove absence of tail truncation:** deleting the last *k* rows
    and updating the chain-head state to match leaves a self-consistent chain.
    No hash chain can detect this from inside itself; it needs a record of the
    head **outside** the store being verified. v1 mitigations, in force from
    day one: (1) the store seam exposes no update or delete (A1-7.2); (2) the
    platform MUST emit `(engagement_id, head_seq, head_hash, integrity_state)`
    to the application log (`slog`, ADR-0019 §3) **and** MUST write the same
    tuple to a store the event-store role cannot UPDATE or DELETE — a separate
    append-only `chain_head_trail` table created with
    `REVOKE UPDATE, DELETE` from the event-store role (A1-7.9's rule) — at
    every verification (A1-6.2), at every append that crosses a
    **100**-`seq` boundary (`HeadLogIntervalSeq`, §4.1) and at every
    `run_ended`/`hard_stop_fired`, so the trail — a different store, with
    different access — contradicts a truncated chain; (3) every delivered
    export carries the head hash (A1-6.6), so a truncated chain contradicts the
    report already in the customer's hands; (4) the platform MUST deliver the
    head tuple of (2), extended with `chain_spec`, to a recipient **outside the
    deployment** over the ADR-0012
    §3 signed webhook, at exactly the emission points of (2) — every
    verification (A1-6.2), every append that crosses a `HeadLogIntervalSeq`
    boundary and every `run_ended`/`hard_stop_fired` — so an independent copy of
    the chain head exists on a host the platform operator does not control
    (product owner decision 2026-09-21, PR #2 item D5; this is head-**hash**
    anchoring, not the external *timestamp* anchoring Q11 excluded). The webhook
    payload MUST carry `engagement_id`, `head_seq`, `head_hash`,
    `integrity_state` and `chain_spec`, signed per ADR-0012 §3, and MUST NOT
    carry event payloads, evidence or secret material (ADR-0019 §5, ADR-0020
    §5); the URL is never stored in an event (A0-3.7), only the channel's
    `target_name`. Each delivery attempt composes `notification_sent` with
    `notification_kind:"chain_head_anchor"` and `channel:"webhook"` (ADR-0012
    §6: the delivery log is itself audited), so the set of anchors the platform
    claims to have sent is reconstructable from the chain it anchors. Delivery
    is best-effort and MUST NOT gate the safety path: a failed or unacknowledged
    delivery MUST NOT block the append, the verification or the export, MUST be
    recorded as `delivery_status:"failed"` with its `attempt`, and MUST be
    retried per ADR-0012 §6 — no webhook outage may stall the event log.
    Composing `notification_sent` for an anchor MUST NOT itself trigger another
    anchor (the emission points of (2) are the only triggers). Recipient-side
    comparison is out of platform scope; the platform's own startup comparison
    remains the `chain_head_trail` row. Startup verification (A1-6.2) MUST
    compare the walked head against the highest `chain_head_trail` row for that
    engagement and MUST report `head_regression` as a break with A1-6.4's
    consequences when the stored head `(head_seq, head_hash)` is lower or
    different from the highest head this platform recorded out-of-band.
    `Tests: TestHeadRegressionDetected, TestHeadTrailIsAppendOnly,`
    `TestHeadAnchorWebhookCarriesTuple,`
    `TestWebhookAnchorFailureNeverBlocksAppend,`
    `TestAnchorDeliveryNeverTriggersAnotherAnchor`.
    _The out-of-band webhook anchor of (4) was **approved** by the product owner
    on 2026-09-21 (PR #2 item D5; §6 item 7 and §6 item 16). The truncation
    residual in A1-6.6 is narrowed accordingly, not closed: forgery is now
    detectable by a recipient that retains its anchors, and detection depends on
    it retaining and comparing them._
  - **Does not survive unrestricted write access to the store plus the log:**
    an attacker who can rewrite both can forge a consistent history. That is
    host compromise, which is outside a hash chain's threat model
    (adversarial A11's residual, A0-5.8's analogue). Since mitigation (4) it is
    nonetheless **detectable from outside**: the forgery contradicts the anchors
    an independent recipient already holds, and suppressing those deliveries is
    itself visible as a head that stops advancing. Prevention is out of scope;
    detection is the recipient's, and A1-6.6 requires that a customer running no
    such recipient be told the case is open for them.

- **A1-5.9** Tamper matrix — the cases the shared contract tests MUST prove
  (`contracts/README.md` merge gate). Each row names the mutation, the expected
  detection, and the `break_kind` (A1-6.3) a verifier MUST report.

  | # | Mutation applied to a clean 3-event chain (§4.3 vector) | Detection | `break_kind` |
  |---|---|---|---|
  | T1 | flip one byte of a stored `payload` value (`exit_code` 0 → 1) | recomputed digest differs from the stored `hash` | `preimage_mismatch` |
  | T2 | rewrite `actor.principal_id` or `actor.type` in the stored bytes | as T1 — the actor is chainable | `preimage_mismatch` |
  | T3 | move `recorded_at`/`occurred_at`/`occurred_claimed_at` by 1 ms | as T1 — all three timestamps are chainable | `preimage_mismatch` |
  | T4 | flip `untrusted` `true` → `false` (untrusted-content laundering, adversarial A1) | as T1 | `preimage_mismatch` |
  | T5 | rewrite `engagement_id` in a stored row | as T1 — the engagement is chainable, so a stolen event cannot be re-labelled | `preimage_mismatch` |
  | T6 | **cross-engagement splicing (A12):** copy engagement A's rows verbatim into engagement B's chain | the row's `engagement_id` is not B | `engagement_mismatch` |
  | T7 | **cross-engagement splicing with re-labelling:** copy A's row into B and rewrite `engagement_id` to B (hash left alone) | digest differs from the stored `hash` | `preimage_mismatch` |
  | T8 | **cross-engagement re-hash:** copy A's rows into B, rewrite `engagement_id`, recompute every `hash`/`prev_hash` consistently | verification **passes** (declared limit, A1-5.8) — the test MUST assert that the out-of-band head trail (log emission / export metadata) no longer matches, and that B's own genesis is still at `seq` 0 (so the splice is a fork, not an extension) | none — out-of-band |
  | T9 | delete a middle row | `seq` no longer consecutive | `seq_gap` |
  | T10 | delete the last *k* rows **and** update chain-head state | verification passes; head hash differs from the logged/exported head hash | none — out-of-band (A1-5.8) |
  | T11 | swap two adjacent rows' `seq` values | `prev_hash` no longer equals the predecessor's `hash` | `link_mismatch` (+ `seq_disorder` when `seq` is not the walk order) |
  | T12 | re-insert a row with an already-used `event_id` at a new `seq` | id uniqueness inside the chain | `duplicate_event_id` |
  | T13 | alter or delete the genesis row, or place a non-`chain_genesis` kind at `seq` 0 | genesis invariant | `genesis_invalid` |
  | T14 | delete rows and leave chain-head `event_count`/`head_seq` untouched | head state disagrees with the actual row count | `row_count_mismatch` |
  | T15 | replace a stored `hash`/`prev_hash` column while leaving the preimage bytes | stored link disagrees with the recomputed digest | `link_mismatch` |
  | T16 | re-serialize the preimage bytes (whitespace, key order) with identical decoded content | byte-exact recomputation fails | `preimage_mismatch` |
  | T17 | canonicalize with a different exclusion set (e.g. also excluding `recorded_at`) | digest differs from the stored one | `preimage_mismatch` (+ the A0-2.13 exclusion-list test) |

  Plus, from A1-5.4/A1-5.7: `TestConcurrentAppendsProduceContiguousSeq` (N
  goroutines appending to one engagement yield `seq` 1…N with no gap, no
  duplicate and a verifying chain) · `TestConcurrentAppendsAcrossEngagements`
  (two engagements' chains are independent: A's appends never change B's head)
  · `TestServedBytesReproducible` (the served canonical envelope is
  byte-identical across two runs and equals canonical(stored preimage + chain
  fields)) · `TestChainVectorDigests` (the §4.3 normative vector, byte-exact).

- **A1-5.10** The chain specification is versioned by `chain_spec` (A1-5.3).
  Changing the digest algorithm, the canonical form, the exclusion list, the
  chainable field set, the linking formula or the envelope key set invalidates
  verification of every historical chain and is therefore **breaking**:
  `/api/v2` plus a new ADR (A0-2.13, A0-6.5), plus a migration and
  checkpoint/re-genesis design that does not exist in v1 (A1-7.9). Adding a new
  event **kind** needs none of that — a new kind is a new payload key set
  inside the same envelope (A0-6.5, A1-3.5).

### A1-6 · Verification and failure behaviour

- **A1-6.1** Verification is a **full walk** of one engagement chain from
  `seq` 0 to `head_seq`, in `seq` order, per event: recompute SHA-256 over the
  stored preimage bytes (A1-5.7) and compare with the stored `hash`; compare
  the stored `prev_hash` with the previous event's stored `hash` (or
  `ChainZero` at `seq` 0); assert `engagement_id`, `event_id` uniqueness and
  the genesis invariant; then compare `head_seq`/`event_count` against the rows
  actually present. Verification MUST NOT sample, MUST NOT trust a stored
  `hash` without recomputing it, MUST NOT stop early because a
  `chain_verified` event exists in the chain, and MUST NOT verify more than one
  engagement per walk (C8). Cost is one SHA-256 per event over bytes that are
  bounded at 32768 B (A1-4.5) — linear, no joins, no re-serialization, and
  cheap enough to run at every startup. _Sampling would turn the chain into a
  spot check, and a spot check is what a tamperer optimizes around._

- **A1-6.2** Verification runs on exactly three triggers — the closed
  `chain_verified.trigger` enum of A1-3.3 (Q11):
  - **`startup`** — before the platform serves any `/api/v1` request for that
    engagement, for every engagement chain it hosts. A chain whose startup
    verification has not completed MUST NOT accept appends (A1-7.11) and MUST
    NOT be exported. Startup verification runs asynchronously, one goroutine
    per engagement chain, owned and cancellable per DESIGN §6; it MUST NOT
    block startup of the platform process. While a walk is incomplete
    `integrity_state` is `unverified`: a read served before the walk completes
    MUST carry `integrity_state:"unverified"` on the same carrier A4 uses for
    `failed` (A1-6.4), appends return `timeout` (A1-7.5), and **no
    customer-facing artifact MUST be produced from an `unverified` chain**
    (`integrity_failed` 409, A0-3.1); other engagements are unaffected.
    `Tests: TestReadDuringStartupWalkIsFlaggedUnverified,
    TestExportFromUnverifiedChainRefused, TestAppendRefusedBeforeStartupVerificationCompletes`.
  - **`pre_export`** — immediately before any customer-facing artifact is
    produced or released: the HTML report, a PDF export, the Q7
    `GET /engagements/{id}/findings` JSON export, and any evidence bundle or
    webhook delivery that carries report content. A cached verification result
    MUST NOT be reused: the walk MUST cover events appended since the last one,
    because the artifact is only as trustworthy as the chain at release time.
  - **`on_demand`** — an explicit operator/admin request (A1-6.8).

  On success the platform MUST append a `chain_verified` event
  (`trigger`, `head_seq`, `head_hash`, `verified_count`, `duration_ms`) to the
  chain it just verified, set `integrity_state:"verified"` with `verified_at`
  and `verified_head_hash` (A1-5.6), and emit the head to the application log
  (A1-5.8). The `chain_verified` event is itself appended **after** the walk and
  is therefore not covered by it; the next walk covers it. _Writing the proof
  of verification into the thing verified is the only way an auditor can tell
  "verified at 14:03 up to seq 4711" from "never verified" without trusting a
  mutable flag._

- **A1-6.3** On failure the platform MUST append exactly **one**
  `chain_break_detected` event per verification run, describing the **first**
  break in `seq` order, and MUST stop the walk there: after a break the
  remaining links are meaningless, and one event per run keeps a startup loop
  from flooding a chain with thousands of break rows. `break_seq` is the first
  failing `seq`, `break_event_id` its `event_id` (`""` when the row is missing
  or unreadable), `expected_prev_hash` the value the link requires,
  `actual_prev_hash` the value found (both equal when the break is not a link
  break), `verified_count` the number of events that passed before the break.
  `break_kind` is the closed enum of A1-3.3, assigned as follows:

  | `break_kind` | Condition |
  |---|---|
  | `preimage_mismatch` | SHA-256 of the stored preimage bytes is not the stored `hash` (any content edit, A1-5.9 T1–T7, T16, T17) |
  | `link_mismatch` | stored `prev_hash` is not the predecessor's stored `hash` (reorder, column-only edit: T11, T15) |
  | `seq_gap` | `seq` values are not consecutive from 0 (deleted middle row: T9) |
  | `seq_disorder` | rows are not returned in ascending `seq` order by the store, or a row's `seq` disagrees with its chain-head position (T11) |
  | `duplicate_event_id` | the same `event_id` appears twice in one chain (T12) |
  | `genesis_invalid` | `seq` 0 is missing, is not `chain_genesis`, has a `prev_hash` other than `ChainZero`, has a `chain_spec` other than the A1-5.3 constant, or a second genesis exists (T13) |
  | `row_count_mismatch` | chain-head `event_count`/`head_seq` disagrees with the rows present (T14) |
  | `engagement_mismatch` | a stored row's `engagement_id` differs from the chain's engagement — the cross-engagement splicing case (A12, T6) |
  | `head_regression` | the stored head `(head_seq, head_hash)` is lower or different from the highest head this platform recorded out-of-band for that engagement in `chain_head_trail` (A1-5.8, T10) |

  The platform MUST set `integrity_state:"failed"`, MUST log the break at error
  level with the correlation attrs (ADR-0019 §3, A0-3.8), and MUST NOT append a
  duplicate `chain_break_detected` for the same `(break_seq, break_kind)` on a
  later run — a repeat is logged, not re-chained, so the state stays visible
  without growing the log. The break-dedup state is
  `ChainHead.LastBreakSeq`/`LastBreakKind` (A1-5.6), written in the same
  transaction as the break event, so two concurrent walks cannot each append
  one. A break MUST NOT be reported as an HTTP 5xx: it is
  a state, not a platform crash (A0-3.1 keeps `integrity_failed` at 409 so it
  is never retried or alerted as a bug).
  `Tests: TestOneBreakEventPerRun, TestBreakDedupStateWrittenInSameTransaction`.

- **A1-6.4** Consequences of `integrity_state:"failed"` (Q11, verbatim
  mapping):
  - **Customer-facing exports are blocked.** Every endpoint that produces or
    releases a customer-facing artifact (report, PDF, findings JSON export,
    evidence bundle) MUST refuse with `integrity_failed` (409, A0-3.1) whose
    `message` follows ADR-0019 §2 and names the engagement, the `break_seq`,
    the `break_kind` and the `chain_break_detected` event id, and whose `attrs`
    carry `engagement_id` (A0-3.6). The remedy is an override (A1-6.5), never a
    retry (A0-3.11).
  - **Internal views are flagged, not blocked.** The UI and `/api/v1` reads
    continue to serve events and graph data, and every such view MUST carry the
    integrity state so the reader sees it: the machine-readable carrier is
    A4's (a response field or header — A4 decides), the human carrier is the
    report/UI wording "integrity verification failed" (Q11). A1 fixes the
    semantics and the metadata (A1-6.6), not the wire shape.
  - **Appends continue.** `events:append` and platform composition MUST keep
    working on a failed chain: stopping evidence capture because the audit
    trail is damaged destroys the very data an investigator needs, and a
    tamperer's cheapest attack would then be "break one row, blind the
    platform". New events link from the current head (A1-5.5); the break stays
    reported at its own `seq` on every later walk. **Decided** (product owner,
    2026-09-21, PR #2 §6.5); the alternative was fail-closed refusal of
    appends.
  - **Blast radius is one engagement.** A failed chain MUST NOT block exports,
    reads or appends of any other engagement, and MUST NOT stop the platform
    process.
  - `integrity_state:"overridden"` (A1-6.5) has the same read/append
    behaviour as `failed` and additionally unblocks exports under the
    conditions of A1-6.5.

  `Tests: TestExportBlockedOnFailedChain` (each of the four artifact classes —
  report HTML, report PDF, findings JSON export, evidence bundle — →
  `integrity_failed` 409 naming `break_seq`/`break_kind`),
  `TestInternalViewOverrideDoesNotAuthorizeExport`,
  `TestAppendsContinueOnFailedChain`.

- **A1-6.5** Operator override — explicit, attributed, single-purpose, never
  silent (Q11). An override MUST be recorded as an `integrity_override` event
  **before** the blocked operation proceeds, carrying:
  - `reason` — non-empty free text (A1-4.2): an override without a stated
    reason is `validation`. The event's `actor` is the human who decided
    (A1-2.1/2.2, `usr_` id per §6.2), `recorded_at` is platform time, and both
    are inside the digest — the override is itself immutable evidence.
  - `break_event_id` — the `chain_break_detected` event the override answers.
  - `scope` — `export` or `internal_view`, and the two are not interchangeable:
    `export` authorizes customer-facing artifact release; `internal_view`
    records acknowledgement of the break for internal use and MUST NOT
    authorize any export. An `export` override MUST NOT be inferred from an
    `internal_view` one.

  Validity: an override applies from its own `seq` until the next
  `chain_verified` or `chain_break_detected` event on that chain — a later
  `chain_verified` supersedes it (the chain is sound again, no override needed)
  and a later `chain_break_detected` invalidates it (a **new** break needs a
  **new** human decision). It is not a
  standing waiver and it has no expiry of its own; each released artifact MUST
  additionally name the override event that released it (A1-6.6), so every
  release under one override is individually attributable — that attribution is
  the compensating control for the risk accepted in ADR-0021.

  Only a `user` principal holding the **admin** role may compose
  `integrity_override` (SPEC §3 places integrity-class controls next to the
  hard stop; Q11's "operator" reads as "human", and this clause narrows it —
  §6 item 15). An operator-scoped user — including one assigned to the
  engagement — MUST NOT override any break. A5 MUST gate the endpoint on the
  admin role and MUST return `forbidden` (A0-3.1, principal-level, naming no
  object) to an operator. A machine principal can never override (Q6, A1-2.7 —
  `integrity_override` is not client-appendable, A1-3.4).
  `Tests: TestOverrideIsAdminOnly, TestOperatorCannotOverrideAnyBreak.`

  Lifetime (ADR-0021, product owner decision 2026-09-11): an
  `integrity_override` with `scope:"export"` is **not single-use**. It stays
  valid until it dies under the rule above and MAY authorize more than one
  artifact release. The platform MUST compose an `artifact_released` event for
  **every** release, naming `override_event_id`, `head_seq`, `head_hash`, the
  artifact's `evi_` id and `recipient_ref`, so the complete set of artifacts
  released under one human decision is reconstructable from the chain alone. A
  release attempted with no live `export` override MUST fail with
  `integrity_failed` (A0-3.1) and MUST NOT compose `artifact_released`.
  _Rationale: a per-artifact rule does not survive the long-term service vision,
  where one release decision legitimately covers a report, its findings export
  and its evidence bundle. The residual risk — one admin decision authorizing
  several customer artifacts from a chain known to be broken — is **accepted and
  transferred to the service owner** who runs the deployment; the platform's
  side of the bargain is that the attribution above is complete, immutable and
  stamped into the export (A1-6.6). An owner who needs a stricter rule
  compensates with their own review and logging of `artifact_released`._
  `Tests: TestEveryReleaseUnderOneOverrideIsChained,`
  `TestReleaseWithoutLiveOverrideRefused, TestOverrideDiesAtNextBreak,`
  `TestArtifactReleasedNamesItsOverride.`

- **A1-6.6** Integrity metadata an export MUST carry. Every customer-facing
  artifact (report HTML/PDF, findings JSON export, evidence bundle) MUST embed
  this exact field set, produced by the platform at release time; the artifact
  format and its wire/document shape belong to the report session and A4, the
  **fields and their meaning** belong to A1 (**Decided** — product owner,
  2026-09-21, PR #2 §6.13):

  | Field | Type | Meaning |
  |---|---|---|
  | `chain_spec` | string | the A1-5.3 constant the chain was built under |
  | `engagement_id` | `eng_` | the chain the artifact was built from |
  | `head_seq` | int | the `seq` of the **last row the walk verified** — identical to the `head_seq` payload of the event named by `chain_verified_event_id`, not the head after that event was appended; `head_seq` therefore equals the post-append `ChainHead.HeadSeq` minus one (A1-5.6) |
  | `head_hash` | 64hex | the digest of that same last verified row — the value an auditor re-checks the log against |
  | `integrity_state` | enum `verified` · `failed_overridden` | machine-readable verdict; `failed` and `unverified` are never exportable (A1-6.4) |
  | `verified_at` | timestamp | when the `pre_export` walk that produced this block finished (A0-5.1) — for `failed_overridden` that is the **failing** walk's completion time |
  | `chain_verified_event_id` | `evt_` | the `chain_verified` event of that walk |
  | `integrity_override_event_id` | `evt_` | the override that released this artifact; `""` when `integrity_state:"verified"` |

  When `integrity_state` is `failed_overridden`, the artifact MUST additionally
  render the literal words **"integrity verification failed"** (Q11) together
  with the override's `reason` and the overriding user's id, in a place a reader
  cannot miss. Prose is prose and MUST NOT be parsed (A0-3.4); the
  machine-readable carrier is `integrity_state`. _The point of the stamp is
  that a customer holding a report can tell a clean chain from an overridden
  one without asking us, and can see who decided — the report is a legal
  artifact (adversarial A11), so "never silent" has to survive into the
  document itself._

  The export `integrity_state` is a **distinct** enum
  (`verified` · `failed_overridden`) derived from A1-5.6's
  (`unverified` · `verified` · `failed` · `overridden`). The two lists MUST NOT
  be used interchangeably:

  | A1-5.6 `integrity_state` | exportable? | A1-6.6 `integrity_state` |
  |---|---|---|
  | `unverified` | no — `integrity_failed` (409, A1-6.2/A1-6.4) | not emitted |
  | `verified` | yes | `verified` |
  | `failed` | no — `integrity_failed` (409, A1-6.4) | not emitted |
  | `overridden` | yes, while the override is live (A1-6.5) | `failed_overridden` |

  The export path MUST compose an `artifact_released` event (A1-3.3, A1-4.2)
  for every released customer-facing artifact, naming `artifact_kind`,
  `artifact_evidence_id`, this block's `head_seq`/`head_hash`,
  `integrity_state`, `override_event_id` (`""` when `verified`) and
  `recipient_ref`. It is the compensating control for the override lifetime of
  A1-6.5 (ADR-0021): the override is not single-use, so this event — one per
  artifact, chained, attributed to the override and to a recipient — is what
  makes the full set of releases under one human decision reconstructable.
  `Tests: TestExportComposesArtifactReleased,`
  `TestExportIntegrityStateIsNotTheChainStateEnum`.

  _Out-of-band anchor (product owner decision 2026-09-21, PR #2 item D5): the
  head tuple is additionally delivered over the ADR-0012 §3 signed webhook
  (A1-5.8 (4)), so a recipient outside the deployment holds an independent copy
  of the chain head. **Residual risk, narrowed but not eliminated:** an attacker
  holding both store-write and log-write access on the same host — the default
  Docker deployment, SPEC §10 — can still forge history and this metadata block
  with it, and can suppress webhook delivery. What the anchor buys is that
  suppression is visible to the recipient as a head that stops advancing, and
  that a forged history contradicts the anchors the recipient already holds.
  Detection therefore depends on the recipient retaining and comparing those
  anchors against this block's `head_seq`/`head_hash`. A customer who runs no
  such recipient MUST be told that the truncation case is open for them._

- **A1-6.7** Third-party re-verification seam. The platform MUST be able to
  produce a **verification bundle** for one engagement: every event's stored
  preimage bytes plus its `seq`, `prev_hash` and `hash`, in `seq` order,
  together with the A1-6.6 metadata block. That bundle is sufficient for an
  independent party to recompute the whole chain with SHA-256 and a canonical
  JSON implementation (A0-2), with no access to Sleipnir. The bundle's wire
  format, transport and access control are A4/report-session; A1 fixes only
  that the stored preimage bytes (A1-5.7) make it possible and that no
  platform-private state is needed. _A tamper-evident log that only the vendor
  can verify is an assertion, not evidence._

- **A1-6.8** Who may trigger verification. `startup` and `pre_export` are
  platform-internal and automatic. `on_demand` MUST be reachable only by a
  **user** principal (admin or an operator assigned to the engagement, SPEC §3,
  A5 decides the scope name) and MUST NOT be reachable by any machine
  principal: a full walk is O(events) and an agent-triggerable walk is a
  denial-of-service lever (adversarial A14), so A4 MUST rate-limit it
  (A0-3.1 `rate_limited`). Verification MUST NOT write any event other than
  `chain_verified` and `chain_break_detected`, MUST NOT mutate any event row
  (A1-7.2), and MUST NOT read another engagement's rows (A1-8.6). A
  verification run MUST be idempotent: running it twice on an unchanged chain
  yields two `chain_verified` events and no state change beyond them.

- **A1-6.9** No in-place repair in v1. There is no "fix the chain" operation,
  no re-hash, no re-genesis, no compaction and no checkpoint: once a break is
  recorded it is permanent for that engagement, and the only forward paths are
  the override (A1-6.5) and continued appending (A1-6.4). A repair or
  re-genesis mechanism would have to preserve the audit meaning of the damaged
  region and is deferred with the rest of the persistence design (A1-7.9,
  backlog session 6) and requires its own ADR. _Anything that can rewrite a
  chain can also hide that it did; v1 chooses the boring option — detect,
  record, flag, and let a human decide._

### A1-7 · Write path

- **A1-7.1** Only the platform composes events (Q6, ADR-0017 §1–§2,
  SPEC §6/C5). Every row in the chain is written by platform code inside the
  platform process; `actor` is stamped from the authenticated principal and the
  request context (A1-2.3), and a platform subsystem recording a client's
  request stamps the **client** as actor (A1-3.3). Orchestrators and workers are
  **report-only**: an orchestrator never touches the container runtime or the
  store (ADR-0017) and a worker's whole surface is `events:append`,
  `evidence:upload`, `task:result` (Q6 worker addendum). Enforcement is the
  three layers of Q6: scopes exclude the verbs (A5) → middleware hard-blocks
  the endpoints for any machine principal → per-request engagement/run binding
  (A1-2.5/2.6). Nothing here is delegated to agent discipline, agent images or
  the store driver (ADR-0005 §5). _The log is the platform's own claim about
  what happened; a component that can write arbitrary rows into it can write
  any history it likes._

- **A1-7.2** Append-only, absolutely. There is no update path and no delete
  path for an event row: the store seam MUST NOT expose an update, delete,
  upsert, truncate or bulk-mutation method for events, and no HTTP route may
  mutate a stored event (A4 route-table audit, mirroring A2-11.1's
  `TestNoBulkReadSpansEngagements`). Post-hoc annotation — a verification
  outcome, a report-inclusion decision, a quarantine change, a retraction, a
  re-validation result, a correction — is always a **new event** that
  references the annotated one through a `*_event_id` payload field (A1-1.6,
  A1-4.11): `approval_executed` annotates `approval_requested`,
  `cleanup_executed` annotates `revert_recorded`, `graph_edge_retracted`
  annotates `graph_edge_written`, `integrity_override` annotates
  `chain_break_detected`. A mutable field inside a hashed document is a
  self-invalidating digest (A1-1.6), and a delete is unauditable by
  construction. The contract suite asserts the seam's shape
  (`TestNoUpdateOrDeleteOnEventSeam`, A1-5.9).

- **A1-7.3** The append request (`events:append`) has a **closed** key set of
  four keys; anything else is an unknown field on a write → `validation` naming
  the field (A0-6.2), including every platform-stamped envelope key of A1-2.3.

  | Key | Type | Req | Rule |
  |---|---|---|---|
  | `kind` | string | yes | MUST be one of the three client-appendable kinds (A1-3.4); unknown → `validation` + `action_blocked{unknown_event_kind}` |
  | `payload` | object | yes | the kind's closed flat key set (A1-4.1/4.2); unknown or missing field → `validation` |
  | `occurred_claimed_at` | timestamp | no | untrusted client time (A0-5.7, A1-1.4); absent → `""` in the envelope; malformed → `validation` (A0-5.3) |
  | `idempotency_key` | string | yes for machine principals | the A1-7.6 dedup key; ≤ 64 chars of `^[A-Za-z0-9_-]{1,64}$`; absent from a machine principal → `validation` |

  Everything else in the envelope is **derived** by the platform: ids
  (`event_id` from `crypto/rand` via A0-1.4), the binding ids
  (`engagement_id`, `run_id`, `job_id`, `task_id`, `node_id`) from the token
  (A1-2.5/2.6), `actor` from the principal (A1-2.1–2.3), `occurred_at` and
  `recorded_at` from the injected clock (A1-1.4, A0-5.4), `untrusted` from the
  payload's `*` fields (A1-4.4), the chain fields from A1-5.4/5.5, and
  `evidence_refs` as the ascending-byte-order deduplicated union of every
  non-empty payload field whose name ends in `_evidence_id` (A1-1.7, A1-4.7) —
  a client MUST NOT supply `evidence_refs` and the platform MUST NOT invent one
  that the payload does not reference.

- **A1-7.4** Which principal may write which kinds — the seam A5 turns into
  scopes and an exclusion list. "Compose" means the platform writes the row;
  "append" means the row is written through `events:append` on the caller's own
  request.

  | Principal kind | `events:append` | Kinds it may cause to be written |
  |---|---|---|
  | `user` (admin/operator/viewer, SPEC §3) | MUST NOT hold it | only via the endpoint that performs the action (A1-2.7): `scope_changed`, `engagement_policy_changed`, `run_started`, `hard_stop_fired`, `approval_granted`/`_denied`, `graph_node_quarantined`, `report_inclusion_changed`, `graph_edge_retracted`, `integrity_override` (admin only, A1-6.5), `cleanup_planned` approval; plus the refusal records of the note below when one of their requests is denied |
  | `orchestrator` (`job_`) | SHOULD NOT hold it in v1 (**Decided** — product owner, 2026-09-21, PR #2 §6.8) | `spawn_requested` (through the spawn broker, ADR-0017 §2), `approval_requested` (through the approval service), `agent_error`, `scope_denied`, `blacklist_denied`, `action_blocked` — all platform-composed with the orchestrator as `actor` |
  | `worker` (`task_`) | MUST hold it (Q6 addendum) | the three **C** kinds only: `command_executed`, `task_result`, `revert_recorded` (A1-3.4) |
  | `node` (`slp_node_`, Q9) | MUST hold it (ADR-0013 offline buffering) | the same three **C** kinds |
  | `platform` | not applicable | every kind, including the integrity kinds (A1-3.3), which no machine principal can reach by construction (A1-2.7) |

  The "kinds it may cause" column lists what a principal's own request asks
  for. Independently of it, the platform composes `action_blocked`,
  `scope_denied`, `blacklist_denied` and `agent_error` with **any** requesting
  principal as `actor` when that principal's request is refused or fails
  (A1-3.3, A1-7.5): those kinds are records of refusal, not privileges, and no
  principal can suppress one about itself (adversarial A1).

  An append of a non-**C** kind by any machine principal is `forbidden` (403,
  A0-3.1, A1-3.4) and MUST be recorded as
  `action_blocked{reason:"append_not_permitted", action_kind:"events_append"}`
  (A1-7.5). A **C** kind whose `actor.type` would not be `worker` or `node` is
  likewise `forbidden`: an orchestrator that never executes a command MUST NOT
  be able to claim one (ADR-0007/0017). A5 MUST list `events:append` as a
  worker/node scope, MUST NOT grant any read scope to a worker (A1-8.4), and
  MUST NOT grant any machine principal a kind outside this table.

- **A1-7.5** Rejection semantics, expressed only in A0-3 error kinds (never
  invented here). Every rejection that is not a pure transport failure MUST
  also be **observable in the chain** where the table says so, because a
  refused write that leaves no trace is how a hallucination loop hides
  (A2-10.8, adversarial A1).

  | Condition | Error kind (HTTP) | Chained record |
  |---|---|---|
  | unknown/malformed `kind` | `validation` (400) | `action_blocked{unknown_event_kind, events_append}` |
  | kind not appendable by this principal (A1-7.4) | `forbidden` (403) | `action_blocked{append_not_permitted, events_append}` |
  | unknown field in body or payload; platform-only field supplied (A1-2.3, A1-7.3) | `validation` (400) naming the field | `action_blocked{append_rejected, events_append}` |
  | missing required payload field, wrong value type, nested object, float, `null` (A1-4.8) | `validation` (400) | `action_blocked{append_rejected, events_append}` |
  | malformed id, unresolvable `*_event_id`/`*_evidence_id` supplied by the client (A1-4.11) | `validation` (400) / `notfound` (404, A0-3.9) | `action_blocked{append_rejected, events_append}` |
  | unknown enum value in a payload (A0-6.3) | `validation` (400) | `action_blocked{append_rejected, events_append}` |
  | malformed timestamp (A0-5.3), including `occurred_claimed_at` | `validation` (400) | `action_blocked{append_rejected, events_append}` |
  | capped field over budget (A1-4.5, mechanism R) | `summary_too_large` (413) naming field, cap, actual bytes | `action_blocked{append_rejected, events_append}` |
  | secret-pattern match (A1-4.9) | `validation` (400) naming field + rule id, never the value | `action_blocked{append_rejected, events_append}` |
  | missing `idempotency_key` from a machine principal (A1-7.6) | `validation` (400) | none — nothing was attempted |
  | `idempotency_key` reused with a **different** payload hash | `conflict` (409, A0-3.1 duplicate write) naming the original `event_id` | none |
  | token bound to engagement A used against B (A1-2.5) | `notfound` (404), **never** `forbidden` (A0-3.9, A12) | none — no existence disclosure |
  | run binding mismatch (A1-2.6) | `forbidden` (403) | `action_blocked{append_not_permitted, events_append}` |
  | token revoked / expired (Q8, hard stop) | `auth` (401) | none |
  | chain not verified at startup (A1-6.2 walk in flight) | `timeout` (504, retryable per A0-3.11 and A1-7.11) | none — nothing was appended |
  | no valid genesis (A1-5.3) | `internal` (500) — platform defect, fail closed (A1-7.11) | none |
  | canonicalization or store failure, uniqueness race (A0-1.4) | `internal` (500) / `timeout` (504, retryable with the same key) | none; logged once (A0-3.8) |

  `internal` (500) is reserved for the no-valid-genesis case (A1-5.3), which is
  a platform defect (A0-3.1) and fail-closed (A1-7.11). An append refused
  because the chain's startup verification has not completed returns
  **`timeout`** (504, retryable per A0-3.11): the walk finishing is a normal
  outcome, not a defect, and the client retries with the same
  `idempotency_key`.
  `Tests: TestAppendRefusedBeforeStartupVerificationCompletes,`
  `TestAppendRefusedWithoutValidGenesis`.

  Every message follows ADR-0019 §2 / A0-3.4
  (`component.Function: what was attempted: key identifiers: cause`), is
  self-contained, and MUST NOT contain SQL, a stack trace, a request body, a
  canonical-JSON dump or a secret value (A0-3.5/3.7, A1-4.9). The
  `action_blocked` record's `detail` carries the same prose — it is untrusted
  content and is marked `*` (A1-4.4). A rejected append stores **nothing**
  except that `action_blocked` event: no partial row, no orphaned evidence
  reference, no consumed `seq` (A1-5.4).

- **A1-7.6** `events:append` deduplication key (the A0-3.11 obligation).
  - **Key:** `(engagement_id, actor.principal_id, kind, idempotency_key)`
    (**Decided** — product owner, 2026-09-21, PR #2 §6.12).
    `idempotency_key` is a client-generated opaque string (≤ 64 chars,
    `^[A-Za-z0-9_-]{1,64}$`) that the client MUST reuse **verbatim** on every
    retry of the same logical append and MUST NOT reuse for a different one.
    It is a request field and a stored dedup column; it is **not** an envelope
    key (A1-1.1's 17 are closed) and is therefore **not** in the digest
    (A1-5.2) — a dedup artifact is not evidence.
  - **Enforcement:** a uniqueness constraint at the store seam, so a race
    cannot create two rows (A0-1.4: a violation surfaces as `internal`, never a
    silent retry loop). The stored dedup record keeps the key, the SHA-256 of
    the accepted canonical payload, and the resulting `event_id`. The compared
    digest covers `kind` + `payload` only: a retry that differs solely in
    `occurred_claimed_at` is a **hit**, and the originally recorded claim
    stands — a retry MUST NOT be able to shift a stored timestamp (A1-1.4,
    A0-5.7).

    `PayloadHash` = `cjson.SHA256Hex(cjson.CanonicalValue(struct{Kind Kind`
    `` `json:"kind"`; Payload Payload `json:"payload"` `` `}{…}))`, computed
    **after** the A1-7.10 normalization steps, so a retry whose array order
    differs is a dedup hit. No other field participates.

    Worked vector (§4.3 row 1's `kind` and `payload`) — normative, like §4.3:
    **len 293 · SHA-256
    `19d976a83cc9d36ac160313a20b80c0745fff805526b7f43f88d05e33c7be5e5`** over

    ```
    {"kind":"command_executed","payload":{"command":"nmap -sV -p 445 10.20.0.14","duration_ms":48210,"exit_code":0,"output_bytes":18432,"output_evidence_id":"evi_01m1y2whfh3ca875z2x8v8h7qt","redacted":false,"target":"10.20.0.14","tool_id":"tool_01m1y2whfhfjdvwqp9pfxqekmf","tool_version":"1.4.2"}}
    ```

    _This is the preimage only — no envelope field, so it is not the §4.3 row 1
    event digest. `Tests: TestPayloadHashVector,`
    `TestIdempotencyKeyReuseWithDifferentPayloadIsConflict,`
    `TestDedupHitWritesNoEvent, TestRetryCannotShiftClaimedTime`._
  - **Semantics:** first write wins. A repeat with an **identical** payload
    hash is a **dedup hit**: the platform writes nothing, consumes no `seq`,
    and returns the stored envelope of the original event with the same success
    status it originally returned (idempotent replay, A1-7.7). A repeat with a
    **different** payload hash is `conflict` (409) naming the original
    `event_id` — key reuse across different content is a client bug and MUST
    NOT silently pick a winner. A dedup hit MUST be logged (`slog`) and MUST
    NOT itself be recorded as an event: a transport retry is not an occurrence
    in the client environment, and chaining retries would let a flaky network
    rewrite the apparent volume of activity.
  - **Why a client key and not a content hash:** content-derived dedup would
    collapse two genuinely identical occurrences — the same command, same
    target, same exit code, run twice — into one event, which is silent
    evidence loss and strictly worse than a duplicate. The client key makes the
    retry decision explicit; a worker/node that buffers offline (ADR-0013, Q9)
    generates it once per buffered event and replays with it.
  - Platform-composed events use the same mechanism internally. The A1-7.6
    uniqueness constraint applies to rows appended through `events:append` only;
    a platform-composed row MUST carry a non-empty deterministic key:
    `approval_*` → `approval_id` + decision · `evidence_stored` → `evidence_id`
    · `job_spawned`/`task_spawned`/`container_started` →
    `spawn_request_event_id` · `graph_*` → the written `gn_`/`ge_` id ·
    `chain_verified` → `trigger` + `head_seq` · `chain_break_detected` →
    `break_seq` + `break_kind` · `cleanup_*` → `revert_event_id` ·
    `notification_sent` → `related_event_id` + `attempt` · `llm_call` → the
    gateway's per-call id · otherwise the composing subsystem's operation id.
    `""` MUST NOT be used, so an internal retry of a composition is idempotent
    too. `Tests: TestPlatformDedupKeyIsNonEmptyPerKind,`
    `TestPlatformDedupKeyEmptyRejected`.

- **A1-7.7** Idempotent replay, and what A2's ingest may rely on (A2-4.7,
  A1-3.8). Guarantees:
  1. `event_id` is unique inside an engagement chain and is **stable across
     retries** — a replayed append returns the original `event_id` (A1-7.6), so
     an ingest consumer that keys on `event_id` sees one identity per
     occurrence.
  2. `(engagement_id, seq)` is unique and permanent: no row is ever renumbered
     (A1-7.2), so `seq` is a safe ingest watermark.
  3. A committed event never changes (A1-1.6, A1-7.2), so an ingest that
     replays from an earlier `seq` re-reads identical bytes.
  4. Duplicate *delivery* to a consumer is possible and harmless: the consumer
     MUST deduplicate by `event_id` and rely on its own content-hash dedup
     (A2-4.7, A2-3.4) to make the replay a no-op.

  Where the ingest consumer stores its watermark is **not** fixed here
  (A1-7.9, backlog session 6); A1 fixes only that `seq` and `event_id` are
  stable enough to watermark on, and that reads are gap-free in `seq` order
  (A1-8.1/8.2).

- **A1-7.8** One event per append request in v1, and one transaction per event.
  The append transaction MUST: take the per-engagement chain lock (A1-5.4),
  assign `seq` and `prev_hash`, canonicalize and hash (A1-5.5), insert the row
  with its preimage bytes (A1-5.7), update the chain-head state (A1-5.6), and
  commit — all or nothing. A batch append endpoint MAY be added additively
  (A0-6.5, A4 owns it) and MUST then be a single transaction that assigns
  consecutive `seq` values in the declared order. An occurrence that logically
  spans two events (a command and its revert record) is two appends and two
  rows, correlated by `task_id` and by the `*_event_id` references
  (A1-4.11) — never one row with two meanings.

- **A1-7.9** Not fixed here (A1 §2): DDL, columns, indexes, partitioning,
  tablespace and archive layout · retention periods and archival (A1-8.8) ·
  the checkpoint / re-genesis mechanism a retention policy would need
  (A1-6.9) · the exact locking primitive and dedup-table shape behind the
  store seam (A1-5.4, A1-7.6) · the ingest watermark store (A1-7.7). All of it
  is backlog session 6 (persistence schema), it MUST keep every A1 guarantee
  above, and one recommendation travels with it: enforce append-only **in the
  database too** (`REVOKE UPDATE, DELETE ON events FROM <app role>`, or a
  rule/trigger that rejects them), so A1-7.2 holds even against a compromised
  application role — defence in depth for adversarial A11 that costs one line
  of DDL.

- **A1-7.10** Composition order and the store seam. Every event — client
  append or platform composition — passes the **same** validator, one pass,
  fail fast, in this order, so error messages are deterministic and testable
  (Q3, A2-10.2's pattern):
  (1) engagement/run binding (A1-2.5/2.6) · (2) request body size bound
  (A0-8.9, A4 sets the number) · (3) unknown fields (A1-7.3, A0-6.2) ·
  (4) `kind` in the closed list and appendable by this principal (A1-3.1/3.4,
  A1-7.4) · (5) payload key set exactly equals the kind's (A1-4.1) ·
  (6) id syntax (A0-1.5) and reference resolution in this engagement
  (A1-4.11) · (7) value types, flatness, enums, timestamps, integer ranges
  (A1-4.8/4.12) · (8) per-kind obligations (A1-4.2) · (9) size caps
  (A1-4.5) · (10) array sort/dedup/count (A1-4.7) · (11) secret scan
  (A1-4.9) · (12) `untrusted` computation (A1-4.4) · (13) `evidence_refs`
  derivation (A1-7.3) · (14) dedup check (A1-7.6) · then chain assignment and
  insert (A1-7.8). A write reports **exactly one** error: the first violated
  rule. Validation MUST NOT be relaxed for internal callers — ingest, the
  broker, the gateway, the approval service and the report builder all pass
  through it, because the platform does not trust discipline, including its own
  (A2-10.7, ADR-0005 §5). Every store-seam write method MUST take an
  engagement id; no unfiltered or multi-engagement write method MAY exist
  (A1-8.6, adversarial A12).

- **A1-7.11** Fail-closed durability: an event is never dropped. The platform
  MUST NOT offer a best-effort, fire-and-forget or "log if the store is up"
  append path. If the event cannot be committed, the operation that would have
  produced it fails (`timeout` 504 / `internal` 500, A0-3.1) and the caller
  retries with the same `idempotency_key` (A1-7.6); a worker or node whose
  append cannot reach the platform buffers locally and replays (ADR-0013, Q9).
  A retry MUST reuse the same `idempotency_key` (A1-7.6); a kind declared
  terminal by A0-3.11 MUST NOT be retried except where its owning contract
  declares the write retryable under `internal` with a deduplication key — A1
  declares exactly that one write retryable, `events:append` under A1-7.6
  (A0-3.11).
  Two preconditions are hard gates, both fail-closed: a chain with no valid
  genesis MUST NOT accept appends (A1-5.3) and a chain whose startup
  verification has not completed MUST NOT accept appends (A1-6.2). The first
  returns `internal` and logs at error level, because it is a platform defect,
  not a client error; the second returns `timeout` (A1-7.5), because the walk
  completing is a normal outcome. _The audit spine is the product's
  differentiator (ADR-0009 Consequences); a silently missing row is
  indistinguishable from a deleted one, and both destroy the report's legal
  value (adversarial A11)._
  `Tests: TestAppendRefusedWithoutValidGenesis,`
  `TestAppendRefusedBeforeStartupVerificationCompletes,`
  `TestRetryReusesIdempotencyKey`.

- **A1-7.12** Commit before success — with one deliberate exception. The
  platform MUST NOT report an action as completed to a caller before the event
  that records it is committed: an action whose event failed to commit is a
  failed action (`timeout`/`internal`), retried idempotently or aborted. There
  is no unaudited success. **Exception — kill paths:** `hard_stop_fired`,
  `container_killed` and token revocation (Q8) MUST NOT block on the event
  store. The kill happens first; its event is written after and retried until
  it lands, and a failure to write it MUST be logged at error level and
  surfaced in the UI. The pending kill record MUST be written to a durable
  **outbox** — an append-only table created with `REVOKE UPDATE, DELETE` from
  the event-store role (A1-7.9) — **before** the kill is issued, and startup
  MUST drain the outbox into the chain; a kill whose event cannot be composed
  MUST surface in the UI as an unresolved integrity warning on A1-6.4's
  carrier, not only in `slog`. _Availability of the kill switch beats durability
  of its record (ADR-0005 §4): a platform that cannot stop a container because
  its database is slow has failed at the only thing that must never fail. The
  gap is bounded and visible, which is the most an audit trail can do about it —
  and the outbox is what makes "bounded" true across a restart (E-03, S-13)._
  `Tests: TestKillPathDoesNotBlockOnEventStore,`
  `TestHardStopProceedsWhenEventStoreUnavailable,`
  `TestKillPathDoesNotBlockOnAppendLatency, TestKillOutboxIsDrainedAtStartup,
  TestKillOutboxIsAppendOnly`.

### A1-8 · Read and stream guarantees

- **A1-8.1** The one order (A0-4.3 obligation discharged). Every event
  collection — paginated read, single-run read, SSE stream, export, verification
  walk — is ordered by **`seq` ascending within one engagement**. `seq` is
  immutable, unique, platform-assigned and dense (A1-5.4), so it is a total,
  stable order built from an immutable unique key. Nothing else may order
  anything: not `recorded_at` (equal values are possible and it is clamped,
  A1-5.4), not `occurred_at`, never `occurred_claimed_at` (untrusted client
  time MUST NOT drive ordering — A0-5.7, A1-1.4), not `event_id` text (A0-1.6:
  id order is a storage convenience, not an API guarantee), and not any payload
  field. A **descending** read (`seq` descending, newest first, for the live UI)
  MAY be offered as a second *declared* order; the direction is an explicit A4
  request parameter and defaults to ascending. Direction is a **request
  parameter**, not a property of a cursor: a cursor issued for the other
  direction is interpreted in the requested direction and yields a
  well-defined (possibly empty) page. The cursor shape is **not** extended
  (§5 conflict 9): its integrity problem is solved by A0-4.4's resolved-row
  seek rule, not by a MAC (A1-8.2). Filtered reads preserve the same order
  (A1-8.3): a filter selects rows, it never reorders them.

- **A1-8.2** Cursors and pagination (A0-4 in full: one `{"items":[…],
  "next_cursor":"…"}` envelope, no offset paging, no total count, `limit`
  default 100 / hard max 1000 hard-rejected not clamped, `limit+1` has-more
  detection). A1 fixes the collection-specific parts:
  - The ordering tuple is `k` = `seq` of the last returned row, `id` = its
    `event_id` (A0-4.4), canonicalized then base64url-unpadded. Cursors are
    opaque; a client replays them byte-for-byte.
  - **A cursor must resolve inside the requested engagement.** The platform
    MUST resolve the cursor's `id` **in the requested engagement** and MUST
    derive the seek position from that row's `seq`, ignoring `k` for seeking; a
    cursor whose `k` disagrees with the resolved row's `seq`, or whose `id` does
    not resolve in this collection, is `validation` (400) telling the client to
    restart from the first page (A0-4.4, A0-4.8). One oracle: A1-8.2 and A2-11.4
    both say `validation`, never "an empty page or `validation`" (P-71, AM-4).
    `Tests: TestCursorFromEngagementARejectedInB,`
    `TestCursorWithInconsistentKAndIDRejected`. _Because `seq` is per engagement
    (A1-5.1), engagement A's cursor `{"k":412,…}` is a perfectly valid position
    in engagement B — replaying it would silently return B's rows from an
    unrelated point. Resolving the id turns a confusing cross-engagement replay
    into a loud error, and makes the A2-11.4-style negative test decidable.
    A0 amendment request AM-4 asks A0-4.8 to bless this case.
    **Decided** (product owner, 2026-09-21, PR #2 §6.11)._
  - Mid-paging semantics are the easy case (A0-4.7): the chain is append-only,
    so a cursor is a **stable watermark** — no row is skipped or duplicated, and
    rows appended after the cursor appear on later pages. A client MUST still
    deduplicate by `event_id`, and MUST NOT treat a paged set as a snapshot of
    anything but the prefix it walked.
  - A client MAY detect gaps itself: `seq` is dense, so `items[i].seq + 1 ==
    items[i+1].seq` within a page and across pages. A gap in a served read is a
    platform defect (`internal`) or tampering (A1-6.3) — never normal.

- **A1-8.3** The closed filter set. A read MAY be filtered by **envelope fields
  only**, each an equality or range predicate, combined with AND:

  | Filter | Shape |
  |---|---|
  | `kind` | one or more values, OR'd inside the filter (repeated parameter); closed list A1-3.1, unknown → `validation` (A0-6.3) |
  | `recorded_at` | inclusive `from` / exclusive `to`, A0-5.1 format |
  | `seq` | inclusive `from` / inclusive `to` |
  | `run_id`, `job_id`, `task_id`, `node_id` | equality on one id, validated per A0-1.5 |
  | `actor_type` | equality, closed enum A1-2.1 |
  | `actor_principal_id` | equality on one id |
  | `evidence_ref` | equality on one `evi_`: "events that reference this artifact" (ADR-0009 §2) |

  There is **no** payload predicate, no free-text search, no aggregation, no
  sort parameter and no join — Q1 ships no free-form query, and A6 stays a
  stub. This is why the taxonomy grows by kinds: an occurrence that must be
  independently *filterable* has to be a `kind`, because `kind` is the only
  payload-visible filter (A1-3.5). Everything else is a payload field, visible
  only after the row is read. A filter that names another engagement's id
  yields an empty page or `notfound` per A0-3.9 — never another engagement's
  rows (A1-8.6).

- **A1-8.4** Who may read — the seam A5 turns into scopes. Reads are
  engagement-scoped for every principal (A1-8.6); this table adds the
  principal-level rules:
  - **`worker` (`task_`) and `node` (`slp_node_`)**: **no read scope at all**
    (Q6 worker addendum, A1-2.6). Any read attempt → `forbidden` (403,
    A0-3.1). A worker reports; it never reads back the log, its own rows
    included. _A worker that can read the log can read every other agent's
    findings, credentials references and target data — a lateral-information
    channel the report-only rule exists to close._
  - **`orchestrator` (`job_`)**: SHOULD hold no event-log read scope in v1
    (**Decided** — product owner, 2026-09-21, PR #2 §6.8). Its planning surface
    is the graph and the stage views (ADR-0016 §4, Q1, A3), not the log; if A5
    grants a read at all it MUST be limited to its own `run_id` and MUST NOT
    include another run's events.
  - **`user`**: admin — every engagement; operator — engagements assigned to
    them (SPEC §3, ADR-0012 §1); viewer — read-only on assigned engagements.
    Assignment is A5's; A1 requires that the read be engagement-scoped and that
    an unassigned engagement resolve to `notfound` (A0-3.9).
  - **Platform-internal readers** (graph ingest, report builder, stage views,
    notification service) read through the same seam with the same engagement
    binding; there is no privileged unscoped read path (A1-7.10).

- **A1-8.5** SSE stream mapping (Q12, ADR-0011 §5, SPEC §9). Framing, endpoints,
  authentication and reconnect headers are A4's; A1 fixes what the stream
  **means** and what a consumer may rely on:
  - One stream per engagement (A1-8.6); an optional `run_id` filter narrows it
    (A1-8.3). There is no cross-engagement stream and no global stream.
  - The SSE `id:` field MUST be the event's decimal **`seq`** (**Decided** —
    product owner, 2026-09-21, PR #2 §6.10), so `Last-Event-ID` is a chain
    position and a reconnect resumes at `seq + 1` — gap-free and in order. The
    `data:` payload MUST be the full 17-key canonical envelope (A1-1.8): the
    stream carries events, not summaries, and a stream consumer needs no second
    read to verify or display one. Event-type naming (`event:`) and any
    control/resync message are A4's.
  - **Emit after commit, in `seq` order.** A stream MUST NOT carry an event
    whose transaction has not committed, and MUST NOT emit `seq` 412 before
    411: an out-of-order or rolled-back emit looks exactly like a chain gap to
    a consumer. Emission is per engagement, from committed rows in `seq` order.
  - Delivery is **at-least-once**: a reconnect may repeat events. A consumer
    MUST deduplicate by `event_id` and MUST treat `seq` as the authority on
    order and completeness. The stream is a live view, never the record: the
    chain is (A1-1.8, A1-5.7).
  - Resume is bounded: if the platform cannot resume from the requested `seq`
    (a reconnect after long absence, a retention boundary — A1-8.8), it MUST
    tell the client to restart from the first page rather than silently skip
    ahead. The mechanism is A4's; the guarantee (never a silent gap) is A1's.
  - Notifications (ADR-0012 §2–§6) are a separate channel with their own
    payload schema (A4/ADR-0012 §3); an A1 `notification_sent` event records
    their delivery, and `notification_kind` is ADR-0012's vocabulary, not an
    event kind (A1-3.3).
  - A stream MUST NOT be reachable by a worker or node principal (A1-8.4):
    SSE is a read.

- **A1-8.6** Engagement-scoped reads only (SPEC C8, adversarial A12). Every
  read method on the store seam MUST require an engagement id; no unfiltered,
  multi-engagement or "all events" read method MAY exist — a method that can be
  called without an engagement id will eventually be (A2-11.1's rule, applied
  here). No endpoint, cursor, filter or SSE stream accepts a list of engagement
  ids. A well-formed `evt_` that belongs to another engagement resolves to
  `notfound` (404), **never** `forbidden` and never a distinguishable message
  (A0-3.9). There is no cross-engagement aggregation, statistics or pattern
  mining in v1 (ADR-0016 §3, C8). Negative contract tests (shared suite,
  "cross-engagement negatives"):
  - `TestEventEngagementBReadNeverReturnsA` — with populated chains in A and B,
    every B-scoped read (list, filtered list, single event, SSE stream, export
    input) returns only B rows; no A `evt_` id appears in any response byte.
  - `TestEventIDFromAIsNotFoundInB` — a well-formed A `evt_` requested in B
    yields `notfound` (404) whose `message` does not distinguish "absent" from
    "elsewhere" (A0-3.9).
  - `TestCursorFromEngagementARejectedInB` — replaying A's cursor on a B-scoped
    list yields `validation` (A1-8.2), never A data and never a B page from an
    unrelated position.
  - `TestFilterWithForeignIDReturnsNothing` — a B-scoped read filtered by an A
    `run_id`/`job_id`/`task_id`/`node_id`/`evi_` returns an empty page.
  - `TestSSEStreamNeverCarriesAnotherEngagement` — events appended to A never
    appear on a B stream, before or after a reconnect.
  - `TestNoBulkEventReadSpansEngagements` — reflection/endpoint audit over the
    A4 route table and the store seam: no event read accepts more than one
    engagement id or omits it (A2-11.4's analogue).

- **A1-8.7** Single-event read. Reading one event by `event_id` returns the
  full 17-key canonical envelope (A1-1.8) — there is no projection, no field
  selection and no compact form in v1 (additive later, A0-6.5). Outcomes:
  malformed id → `validation` (A0-1.5); well-formed id not in this engagement's
  chain → `notfound` (404), which is also the cross-engagement case (A0-3.9);
  success → the envelope plus, when the chain is not `verified`, the integrity
  state (A1-6.4). A read MUST NOT trigger verification (A1-6.8) and MUST NOT
  fail because a referenced `evi_` artifact is unavailable (A1-8.8).

- **A1-8.8** Retention and cleanup seam (ADR-0009). ADR-0009 §4's cleanup is
  about the **client environment** (revert records → plan → execute → verify);
  it is not log deletion, and nothing in A1 authorizes deleting log rows
  (A1-7.2). Retention, archival and compaction of the log and of evidence
  artifacts are **not fixed here** (A1-7.9, backlog session 6). Any mechanism
  that is later introduced MUST preserve these invariants:
  - it MUST NOT break chain verification: no row below the head may be removed
    without a checkpoint/re-genesis design that A1-6.9 defers to a new ADR;
  - it MUST NOT renumber `seq`, reuse an `event_id`, or alter stored preimage
    bytes (A1-5.4/5.7);
  - it MUST keep events at least as long as the artifacts and reports that
    reference them, so a delivered report's `evidence_refs` and
    `*_event_id` links stay resolvable for the engagement's retention period;
  - an archived-off artifact MUST degrade to a **dangling reference, not an
    error**: a read MUST succeed and the report builder MUST mark the reference
    unavailable (A1-8.7). A dangling `evi_` is not an integrity break — the
    event's hash covers the reference, not the bytes;
  - evidence artifacts are write-once while they live (ADR-0009 §2, adversarial
    A11) and their deletion, when a policy eventually allows it, MUST itself be
    recorded as an `evidence`-component event so the removal is auditable.

- **A1-8.9** What a consumer may rely on — the guarantee table for A3, A4, the
  report builder and A2's ingest. Anything not listed here is not promised.

  | Guarantee | Clause |
  |---|---|
  | one envelope shape, 17 keys, all always present, served in canonical form — byte-reproducible across platforms and releases | A1-1.1, A1-1.2, A1-1.8 |
  | `seq` order = commit order = chain order; dense, immutable, unique per engagement | A1-5.4, A1-8.1 |
  | a committed event never changes; annotation is a new event | A1-1.6, A1-7.2 |
  | `event_id` is unique per engagement and stable across retries | A1-7.6, A1-7.7 |
  | `recorded_at` is platform time and non-decreasing along `seq`; `occurred_claimed_at` is untrusted and orders nothing | A1-1.4, A1-5.4, A0-5.6/5.7 |
  | `actor` is platform-stamped; a client cannot forge it | A1-2.3 |
  | `payload` is flat, closed per kind, capped, secret-free, and every string is untrusted-for-rendering | A1-4 |
  | a filtered read is a subset in the same order; the filter set is closed and envelope-only | A1-8.3 |
  | cursor paging is a stable watermark: no skips, no duplicates, `next_cursor` absent when exhausted | A0-4.2/4.6/4.7, A1-8.2 |
  | SSE `id:` is `seq`; at-least-once, in order, after commit | A1-8.5 |
  | reads never span engagements; a foreign id is `notfound` | A1-8.6, A0-3.9 |
  | the chain head `(head_seq, head_hash)` and the engagement's `integrity_state` are available to any authorized reader and to every export | A1-5.6, A1-6.4, A1-6.6 |
  | no total count anywhere (A0-4.1); `chain_verified.verified_count` and `head_seq` are the only counts the platform states | A0-4.1, A1-6.2 |

## 4. Types

Illustrative sketches — **not compiled** (`contracts/README.md`). They are the
source of truth for field names and JSON shapes until `internal/events` merges
(DESIGN §1: domain types, no I/O; the store seam is `internal/store`, the only
pgx importer is `store/postgres`, ADR-0010). Foundation imports only:
`internal/ids`, `internal/cjson`, `internal/errs` (A0 §4).

### 4.1 Go sketch

```go
// internal/events — the A1 contract types. Domain layer: types + invariants,
// no I/O (DESIGN §1). Foundation imports only.
package events

// Kind is the closed event taxonomy (A1-3.1, 42 kinds). A0-8.5 spelling,
// byte-exact comparison, no synonyms.
type Kind string

const (
	// Integrity (A1-3.3).
	KindChainGenesis       Kind = "chain_genesis"
	KindChainVerified      Kind = "chain_verified"
	KindChainBreakDetected Kind = "chain_break_detected"
	KindIntegrityOverride  Kind = "integrity_override"
	KindArtifactReleased   Kind = "artifact_released" // A1-6.6, override attribution (A1-6.5, ADR-0021)

	// Engagement and run lifecycle.
	KindEngagementCreated        Kind = "engagement_created" // seq 1 (A1-5.3 keeps genesis at seq 0)
	KindEngagementClosed         Kind = "engagement_closed"
	KindScopeChanged             Kind = "scope_changed"
	KindEngagementPolicyChanged  Kind = "engagement_policy_changed"
	KindRunStarted               Kind = "run_started"
	KindRunEnded                 Kind = "run_ended"
	KindModelConfigSnapshotted   Kind = "model_config_snapshotted"
	KindJobSpawned               Kind = "job_spawned"
	KindTaskSpawned              Kind = "task_spawned"
	KindContainerStarted         Kind = "container_started"
	KindContainerKilled          Kind = "container_killed"
	KindHardStopFired            Kind = "hard_stop_fired"

	// Spawn request (Q14, ADR-0017).
	KindSpawnRequested Kind = "spawn_requested"

	// Command execution, evidence, results — the client-appendable C kinds.
	KindCommandExecuted Kind = "command_executed" // C (A1-3.4)
	KindEvidenceStored  Kind = "evidence_stored"
	KindTaskResult      Kind = "task_result"      // C
	KindRevertRecorded  Kind = "revert_recorded"  // C

	// Approvals (Q10, ADR-0018).
	KindApprovalRequested Kind = "approval_requested"
	KindApprovalGranted   Kind = "approval_granted"
	KindApprovalDenied    Kind = "approval_denied"
	KindApprovalExpired   Kind = "approval_expired"
	KindApprovalExecuted  Kind = "approval_executed"

	// LLM traffic (ADR-0020).
	KindLLMCall Kind = "llm_call"

	// Graph mutation (ADR-0016, A2 seam — A1-3.6).
	KindGraphNodeWritten      Kind = "graph_node_written"
	KindGraphEdgeWritten      Kind = "graph_edge_written"
	KindGraphNodeQuarantined  Kind = "graph_node_quarantined"
	KindQuarantineRecomputed  Kind = "quarantine_recomputed" // added for A2-8.5
	KindGraphEdgeRetracted    Kind = "graph_edge_retracted"   // added for A2-3.9
	KindReportInclusionChanged Kind = "report_inclusion_changed"

	// Cleanup (ADR-0009 §4).
	KindCleanupPlanned  Kind = "cleanup_planned"
	KindCleanupExecuted Kind = "cleanup_executed"
	KindCleanupVerified Kind = "cleanup_verified"

	// Enforcement denials and agent errors.
	KindScopeDenied    Kind = "scope_denied"
	KindBlacklistDenied Kind = "blacklist_denied"
	KindActionBlocked  Kind = "action_blocked"
	KindAgentError     Kind = "agent_error"

	// Notification delivery (ADR-0012).
	KindNotificationSent Kind = "notification_sent"
)

// HeadLogIntervalSeq is the A1-5.8 out-of-band emission interval: the head
// hash is written to chain_head_trail at every verification, at every append
// crossing this boundary, and at every run_ended / hard_stop_fired.
const HeadLogIntervalSeq int64 = 100 // A1-5.8

// ClientAppendable reports whether k is one of the three C kinds reachable
// through events:append (A1-3.4, A1-7.4).
func ClientAppendable(k Kind) bool

type ActorType string // A1-2.1, closed.

const (
	ActorUser         ActorType = "user"
	ActorOrchestrator ActorType = "orchestrator"
	ActorWorker       ActorType = "worker"
	ActorPlatform     ActorType = "platform"
	ActorNode         ActorType = "node"
)

type Component string // A1-2.4, closed: the platform subsystem that wrote the row.

const (
	CompEventStore  Component = "event_store"
	CompScope       Component = "scope"
	CompApproval    Component = "approval"
	CompSpawnBroker Component = "spawn_broker"
	CompLLMGateway  Component = "llm_gateway"
	CompGraph       Component = "graph"
	CompEvidence    Component = "evidence"
	CompCleanup     Component = "cleanup"
	CompIntegrity   Component = "integrity"
	CompNotify      Component = "notify"
	CompRuntime     Component = "runtime"
	CompAPI         Component = "api"
)

// Actor is exactly three keys, all always present (A1-2.1, A0-2.14).
type Actor struct {
	Type        ActorType `json:"type"`
	PrincipalID string    `json:"principal_id"` // A0-1.2 id selected by Type; "" for platform
	Component   Component `json:"component"`    // "" unless Type == platform
}

// ChainBlock groups the three chain fields. It is EMBEDDED in Event so the
// fields serialize at the envelope's top level (A1-1.3): A0-2.12 exclusion
// lists take plain top-level names only, so a nested "chain":{…} object would
// be unexcludable and is forbidden.
type ChainBlock struct {
	Seq      int64  `json:"seq"`       // per-engagement position (A1-5.4); excluded from the digest
	PrevHash string `json:"prev_hash"` // hash of seq-1, or ChainZero at genesis (A1-5.5)
	Hash     string `json:"hash"`      // SHA-256 over the canonical preimage (A1-5.2)
}

// Event is the one envelope: 17 top-level keys, all always present with their
// zero value when unset (A1-1.1, A1-1.2). No omitempty anywhere — a missing key
// changes the canonical bytes and therefore the digest. Payload is a
// kind-specific struct (A1-4.1) behind a custom marshaler that keeps the key
// set closed and flat.
type Event struct {
	EventID           string    `json:"event_id"`            // evt_ (A0-1.2)
	EngagementID      string    `json:"engagement_id"`       // eng_, chain scope (A1-5.1)
	RunID             string    `json:"run_id"`              // "" when not run-scoped
	JobID             string    `json:"job_id"`              // orchestrator container
	TaskID            string    `json:"task_id"`             // worker container
	NodeID            string    `json:"node_id"`             // remote agent node slp_node_ (Q9)
	OccurredAt        string    `json:"occurred_at"`         // platform stamp (A1-1.4, A0-5.1)
	OccurredClaimedAt string    `json:"occurred_claimed_at"` // untrusted (A0-5.7)
	RecordedAt        string    `json:"recorded_at"`         // authoritative ingest time (A0-5.6)
	Actor             Actor     `json:"actor"`
	Kind              Kind      `json:"kind"`
	Payload           Payload   `json:"payload"`             // one flat type per kind (A1-4.1)
	EvidenceRefs      []string  `json:"evidence_refs"`       // evi_, sorted + deduped (A1-4.7); [] when none
	Untrusted         bool      `json:"untrusted"`           // platform-computed (A1-4.4)
	ChainBlock                                              // seq, prev_hash, hash — top level (A1-1.3)
}

// Payload is the interface every per-kind payload struct satisfies: a flat
// object (A1-4.8) with a closed key set (A1-4.1) and a Kind it belongs to.
type Payload interface {
	Kind() Kind
	Validate() error // A1-4.2 obligations + A1-4.5 caps; errs kind "validation"/"summary_too_large"
}

// A1-local cap constants (A1-4.5, mechanism R everywhere). Q4 constants come
// from A0-7.1 and are not redeclared; §6 asks A0 to adopt these into its table.
const (
	ProseLongMaxBytes   = 2048 // command, task_description, result_summary, revert_action
	ProseMediumMaxBytes = 512  // reason, detail, action_summary, message, entry
	TargetMaxBytes      = 256  // target, attempted_target, blacklist_entry
	LabelMaxBytes       = 128  // origin, container_ref, network_name, media_type, model/endpoint/target_name, old/new_value
	VersionMaxBytes     = 64   // tool_version
	KindNameMaxBytes    = 32   // node_kind, edge_kind, risk_tier (A2 owns the values)
	DigestMaxBytes      = 256  // image_digest
	EvidenceRefsMax     = 8    // evidence_refs count (A1-4.7)
	EventRefsMax        = 64   // revert_event_ids / non_revertable_event_ids count
	ExitCodeMin         = -1   // -1 = no exit status (A1-4.2)
	ExitCodeMax         = 255

	// EventMaxCanonicalBytes bounds the canonical bytes of any event: 6x the
	// largest decoded prose cap (A0-2.7 worst-case \u00xx expansion) plus
	// envelope overhead. A platform invariant asserted at composition, not a
	// client-facing cap (A1-4.5).
	EventMaxCanonicalBytes = 32768

	// IdempotencyKeyMaxBytes bounds the A1-7.6 client dedup key.
	IdempotencyKeyMaxBytes = 64
)

// Chain constants (A1-5.3, A1-5.5).
const (
	ChainSpecV1 = "sleipnir/chain/v1"                                    // genesis payload value
	ChainZero   = "0000000000000000000000000000000000000000000000000000000000000000" // 64 '0'
)

// ChainExclude is the A0-2.12 exclusion list — exactly three plain top-level
// names, immutable after Freeze (A0-2.13, A1-5.2).
var ChainExclude = []string{"hash", "prev_hash", "seq"}

// Preimage returns the exact canonical bytes hash is computed over (A1-5.2,
// A1-5.7). Those bytes are persisted with the row and are the only input a
// verifier may use.
func Preimage(e Event) ([]byte, error) // cjson.CanonicalValue(e, ChainExclude...)

// HashEvent digests the preimage (A1-5.5, A0-2.15).
func HashEvent(preimage []byte) string // cjson.SHA256Hex

// Served returns the canonical 17-key representation of a stored event:
// decode the stored preimage, add the three chain fields, re-canonicalize
// (A1-1.8, A1-5.7). Never re-marshal the decoded struct.
//
// Served(preimage, c) = cjson.With(preimage, map[string]any{"seq": c.Seq,
// "prev_hash": c.PrevHash, "hash": c.Hash}) — the generic document only, never
// a typed struct; it decodes with UseNumber() and re-emits every number's
// literal text verbatim (A0-2.5), and rejects a preimage that fails A0-2 with
// preimage_mismatch (A1-6.3).
func Served(preimage []byte, c ChainBlock) ([]byte, error)

// UnmarshalEvent decodes a served event document in two passes, both with
// DisallowUnknownFields: the envelope keys give kind, then payload decodes into
// that kind's concrete type (A1-4.1). No map[string]any intermediate (A0-2.5).
func UnmarshalEvent(b []byte) (Event, error)

// AppendRequest is the events:append body (A1-7.3): four keys, closed.
// Everything else in an envelope is derived by the platform; a body carrying
// event_id, actor, engagement_id, seq, hash, ... is an unknown field on a
// write (A1-2.3, A0-6.2). Decoding is two passes, both with
// DisallowUnknownFields: read the envelope keys to learn "kind", select the
// concrete payload type for it (A1-4.1), then decode "payload" into that type
// — no map[string]any intermediate, which would defeat A0-2.5's duplicate-key
// detection and A1-4.1's closed key set.
type AppendRequest struct {
	Kind              Kind    `json:"kind"`
	Payload           Payload `json:"payload"`
	OccurredClaimedAt string  `json:"occurred_claimed_at,omitempty"` // untrusted (A0-5.7)
	IdempotencyKey    string  `json:"idempotency_key"`               // A1-7.6; not an envelope key, not hashed
}

// DedupRecord is the stored half of the A1-7.6 key. Not a wire shape, not an
// event field, not hashed.
type DedupRecord struct {
	EngagementID string
	PrincipalID  string
	Kind         Kind
	Key          string // client idempotency_key
	PayloadHash  string // SHA-256 of the accepted canonical payload
	EventID      string // the event the first write produced
}

// ChainHead is the per-engagement bookkeeping of A1-5.6. Mutable platform
// state, never event content (A1-1.6) and never hashed.
type IntegrityState string

const (
	IntegrityUnverified IntegrityState = "unverified"
	IntegrityVerified   IntegrityState = "verified"
	IntegrityFailed     IntegrityState = "failed"
	IntegrityOverridden IntegrityState = "overridden"
)

type ChainHead struct {
	EngagementID     string
	ChainSpec        string
	HeadSeq          int64
	HeadHash         string
	EventCount       int64
	IntegrityState   IntegrityState
	VerifiedAt       string // "" until first verified
	VerifiedHeadHash string

	// A1-5.4 / A1-6.3 bookkeeping, written inside the append transaction.
	LastRecordedAt   string    // the forward-clamp input of the next append (A1-5.4)
	LastBreakSeq     int64     // break-dedup state (A1-6.3); 0 when no break recorded
	LastBreakKind    BreakKind // "" when no break recorded
	LastBreakEventID string    // the chain_break_detected row of the last break
}

// Verification triggers and break kinds — the closed enums of A1-3.3.
type VerifyTrigger string

const (
	TriggerStartup   VerifyTrigger = "startup"
	TriggerPreExport VerifyTrigger = "pre_export"
	TriggerOnDemand  VerifyTrigger = "on_demand"
)

type BreakKind string

const (
	BreakPreimageMismatch   BreakKind = "preimage_mismatch"
	BreakLinkMismatch       BreakKind = "link_mismatch"
	BreakSeqGap             BreakKind = "seq_gap"
	BreakSeqDisorder        BreakKind = "seq_disorder"
	BreakDuplicateEventID   BreakKind = "duplicate_event_id"
	BreakGenesisInvalid     BreakKind = "genesis_invalid"
	BreakRowCountMismatch   BreakKind = "row_count_mismatch"
	BreakEngagementMismatch BreakKind = "engagement_mismatch"
	BreakHeadRegression     BreakKind = "head_regression" // A1-5.8, A1-6.3
)

// VerifyResult is what one walk returns (A1-6.1). The platform turns it into a
// chain_verified or chain_break_detected event; it is not served as-is.
type VerifyResult struct {
	EngagementID   string
	Trigger        VerifyTrigger
	OK             bool
	HeadSeq        int64
	HeadHash       string
	VerifiedCount  int64
	DurationMS     int64
	Break          BreakKind // zero value when OK
	BreakSeq       int64
	BreakEventID   string
	ExpectedPrev   string
	ActualPrev     string
}

// ExportIntegrity is the A1-6.6 metadata block every customer-facing artifact
// must carry. The document that embeds it is the report session's / A4's.
type ExportIntegrity struct {
	ChainSpec               string `json:"chain_spec"`
	EngagementID            string `json:"engagement_id"`
	HeadSeq                 int64  `json:"head_seq"`
	HeadHash                string `json:"head_hash"`
	IntegrityState          string `json:"integrity_state"` // verified | failed_overridden
	VerifiedAt              string `json:"verified_at"`
	ChainVerifiedEventID    string `json:"chain_verified_event_id"`
	IntegrityOverrideEventID string `json:"integrity_override_event_id"` // "" when verified
}

// NewEvent is the only constructor (DESIGN §4: types carry their invariants).
// It runs the A1-7.10 pass and returns errs kinds validation /
// summary_too_large / forbidden / notfound (A1-7.5). Chain fields are assigned
// by the store seam inside the append transaction (A1-5.4, A1-7.8).
func NewEvent(in Draft, act Actor, bind Binding) (Event, error)

// Draft is the platform-internal composition input: no ids, no actor, no
// chain fields, no timestamps. Binding carries engagement/run/job/task/node
// from the authenticated request (A1-2.5/2.6).
type Draft struct {
	Kind              Kind
	Payload           Payload
	OccurredClaimedAt string
	IdempotencyKey    string
}

type Binding struct {
	EngagementID, RunID, JobID, TaskID, NodeID string
}
```

```go
// Payload structs: one per kind, transcribed 1:1 from the A1-3.3 tables
// (field names exact, every field always present — A1-4.1, A0-2.14). Two of
// the three client-appendable C kinds and the four integrity/graph shapes that
// A1-3.8 and A1-6 depend on are shown; the remaining kinds follow the same
// pattern and carry no additional rule.

// CommandExecutedPayload — kind command_executed (C, ADR-0009 §1).
type CommandExecutedPayload struct {
	Command          string `json:"command"`            // ≤ 2048 B, untrusted (A1-4.4)
	Target           string `json:"target"`             // ≤ 256 B, untrusted
	ToolID           string `json:"tool_id"`            // tool_ (A0-1.2)
	ToolVersion      string `json:"tool_version"`       // ≤ 64 B
	ExitCode         int    `json:"exit_code"`          // [-1,255]; -1 = no exit status
	DurationMS       int64  `json:"duration_ms"`
	OutputBytes      int64  `json:"output_bytes"`
	OutputEvidenceID string `json:"output_evidence_id"` // evi_; non-empty when OutputBytes > 0
	Redacted         bool   `json:"redacted"`           // A1-4.6
}

// TaskResultPayload — kind task_result (C, Q6 worker scope).
type TaskResultPayload struct {
	Status         string   `json:"status"` // enum succeeded|failed|partial
	ResultSummary  string   `json:"result_summary"` // ≤ 2048 B, untrusted
	DurationMS     int64    `json:"duration_ms"`
	CommandCount   int64    `json:"command_count"`
	RevertEventIDs []string `json:"revert_event_ids"`     // ≤ 64, sorted (A1-4.7)
	ErrorKind      string   `json:"error_kind"`           // A0-3.1 kind; non-empty when Status == failed
}

// ChainGenesisPayload — kind chain_genesis (A1-5.3).
type ChainGenesisPayload struct {
	ChainSpec string `json:"chain_spec"` // == ChainSpecV1, byte-exact
}

// ChainBreakDetectedPayload — kind chain_break_detected (A1-6.3).
type ChainBreakDetectedPayload struct {
	BreakKind        BreakKind `json:"break_kind"`
	BreakSeq         int64     `json:"break_seq"`
	BreakEventID     string    `json:"break_event_id"` // "" when the row is missing
	ExpectedPrevHash string    `json:"expected_prev_hash"`
	ActualPrevHash   string    `json:"actual_prev_hash"`
	VerifiedCount    int64     `json:"verified_count"`
}

// IntegrityOverridePayload — kind integrity_override (A1-6.5).
type IntegrityOverridePayload struct {
	Reason       string `json:"reason"`         // ≤ 512 B, untrusted, MUST be non-empty
	BreakEventID string `json:"break_event_id"` // chain_break_detected event answered
	Scope        string `json:"scope"`          // enum export|internal_view
}

// GraphEdgeRetractedPayload — kind graph_edge_retracted (added for A2-3.9).
type GraphEdgeRetractedPayload struct {
	GraphEdgeID      string `json:"graph_edge_id"`       // ge_ (A2)
	EdgeKind         string `json:"edge_kind"`           // ≤ 32 B; A2 owns the value list
	FromGraphNodeID  string `json:"from_graph_node_id"`  // gn_
	ToGraphNodeID    string `json:"to_graph_node_id"`    // gn_
	Reason           string `json:"reason"`              // ≤ 512 B, untrusted
	SourceEventID    string `json:"source_event_id"`     // evt_ that contradicted the edge; "" if operator-initiated
}

// QuarantineRecomputedPayload — kind quarantine_recomputed (added for A2-8.5).
type QuarantineRecomputedPayload struct {
	Trigger          string `json:"trigger"` // enum scope_changed|blacklist_changed|policy_changed
	TriggerEventID   string `json:"trigger_event_id"`
	NodesEvaluated   int64  `json:"nodes_evaluated"`
	NodesQuarantined int64  `json:"nodes_quarantined"`
	NodesReleased    int64  `json:"nodes_released"`
	DurationMS       int64  `json:"duration_ms"`
}
```

### 4.2 JSON examples

The append request a worker sends (A1-7.3): four keys, no envelope field, no
`actor`, no id, no `seq`. `evidence_refs` and `untrusted` are derived by the
platform (A1-7.3, A1-4.4).

```json
{
  "kind": "command_executed",
  "payload": {
    "command": "nmap -sV -p 445 10.20.0.14",
    "target": "10.20.0.14",
    "tool_id": "tool_01m1y2whfhfjdvwqp9pfxqekmf",
    "tool_version": "1.4.2",
    "exit_code": 0,
    "duration_ms": 48210,
    "output_bytes": 18432,
    "output_evidence_id": "evi_01m1y2whfh3ca875z2x8v8h7qt",
    "redacted": false
  },
  "occurred_claimed_at": "2026-09-07T14:03:19.900Z",
  "idempotency_key": "tsk7f3q-cmd-00042"
}
```

A served event (A1-1.8). Shown pretty-printed for readability; the **served
form is canonical** (A0-2: keys in UTF-8 byte order, no whitespace) and all 17
keys are always present (A1-1.2). This is the `graph_edge_retracted` kind added
for A2-3.9.

```json
{
  "event_id": "evt_01m1y2whfhm6p3kq9z2x1vb4rt",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
  "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
  "task_id": "",
  "node_id": "",
  "occurred_at": "2026-09-07T15:22:04.118Z",
  "occurred_claimed_at": "",
  "recorded_at": "2026-09-07T15:22:04.140Z",
  "actor": { "type": "platform", "principal_id": "", "component": "graph" },
  "kind": "graph_edge_retracted",
  "payload": {
    "graph_edge_id": "ge_01m1y2whfh62ej11jf4x5gjzv4",
    "edge_kind": "reachable",
    "from_graph_node_id": "gn_01m1y2whfh9x2b4c7d1e8f0a3b",
    "to_graph_node_id": "gn_01m1y2whfhdc01srv4x8mc5a0g",
    "reason": "Reachability was inferred from a single ICMP reply that a second observation contradicted; the edge is withdrawn, its endpoints are unchanged.",
    "source_event_id": "evt_01m1y2whfhp17g0avdqztd2p3x"
  },
  "evidence_refs": [],
  "untrusted": true,
  "seq": 4313,
  "prev_hash": "4c81f6a0d92b57e13c7a4b08fd26e5910b3c7a48d2e6f1093ba5c8d7e4026f1b",
  "hash": "3710d121becfde25268d96bcf9f47f6618bd028842ea8324596fd66d58397647"
}
```

The integrity kinds (A1-6.3, A1-6.5): a break, and the human decision that
released an export anyway. Both are ordinary chained events — the override is
itself immutable evidence, with the deciding user as `actor`.

```json
{
  "event_id": "evt_01m1y2whfhq9z4n8x2c5vb1rtk",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "run_id": "",
  "job_id": "",
  "task_id": "",
  "node_id": "",
  "occurred_at": "2026-09-07T18:02:11.400Z",
  "occurred_claimed_at": "",
  "recorded_at": "2026-09-07T18:02:11.400Z",
  "actor": { "type": "platform", "principal_id": "", "component": "integrity" },
  "kind": "chain_break_detected",
  "payload": {
    "break_kind": "preimage_mismatch",
    "break_seq": 4127,
    "break_event_id": "evt_01m1y2whfhd8k3m9qz1x7vb4nr",
    "expected_prev_hash": "1c0f8ab73e5d9246b0a17fe4c8d2593ba6e0147d92c85b3f10ae6d47c8b92015",
    "actual_prev_hash": "1c0f8ab73e5d9246b0a17fe4c8d2593ba6e0147d92c85b3f10ae6d47c8b92015",
    "verified_count": 4127
  },
  "evidence_refs": [],
  "untrusted": false,
  "seq": 4501,
  "prev_hash": "a19f4c07be35d8219f06c2b4e7d09a13f5b8c6e2d4a97013bf6e8c25a0d4719e",
  "hash": "db711606e0868d55f0bdc2b0dd725ce4248361430065c0a52870df94cd3f506a"
}
```

```json
{
  "event_id": "evt_01m1y2whfhr2t5v8x1z4b7d0f2",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "run_id": "",
  "job_id": "",
  "task_id": "",
  "node_id": "",
  "occurred_at": "2026-09-07T18:40:55.010Z",
  "occurred_claimed_at": "",
  "recorded_at": "2026-09-07T18:40:55.010Z",
  "actor": {
    "type": "user",
    "principal_id": "usr_01m1y2whfhv3x6z9b2d5f8h1jk",
    "component": ""
  },
  "kind": "integrity_override",
  "payload": {
    "reason": "Break at seq 4127 matches a known bad-disk page replaced from the standby replica; the region was re-read from the 18:00 backup and the customer report is due today. Export released with the integrity stamp.",
    "break_event_id": "evt_01m1y2whfhq9z4n8x2c5vb1rtk",
    "scope": "export"
  },
  "evidence_refs": [],
  "untrusted": true,
  "seq": 4502,
  "prev_hash": "db711606e0868d55f0bdc2b0dd725ce4248361430065c0a52870df94cd3f506a",
  "hash": "3359ec51ec7c3c565c0dc151b15d05591e377f9e7e49584ea827d18106325135"
}
```

_The two events above are linked: the override's `prev_hash` is the break
event's `hash` (A1-5.5). Their own `prev_hash` values point at predecessors not
shown here; each `hash` is the SHA-256 of exactly that event's 14 chainable keys
in canonical form (A1-5.2), so both are reproducible from this document._

A paginated read (A0-4.1, A1-8.1/8.2): `k` is the last row's `seq`, `id` its
`event_id`; `next_cursor` absent means exhausted (A0-4.2).

```json
{
  "items": [
    {
      "event_id": "evt_01m1y2whfhp17g0avdqztd2p3x",
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "seq": 1,
      "kind": "command_executed"
    }
  ],
  "next_cursor": "eyJpZCI6ImV2dF8wMW0xeTJ3aGZocDE3ZzBhdmRxenRkMnAzIiwiayI6MX0"
}
```

_The item above is abbreviated to four keys for legibility; a served item is
always the full 17-key canonical envelope (A1-1.8, A1-8.7). There is no compact
projection in v1._

An SSE frame's `data:` payload (A1-8.5; framing is A4's) — `id:` is the decimal
`seq`, so `Last-Event-ID` resumes at `seq + 1`:

```
id: 4313
data: {"actor":{"component":"graph","principal_id":"","type":"platform"},"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","event_id":"evt_01m1y2whfhm6p3kq9z2x1vb4rt","evidence_refs":[],"hash":"3710d121becfde25268d96bcf9f47f6618bd028842ea8324596fd66d58397647","job_id":"job_01m1y2whfhbt69j0h0fbxepw90","kind":"graph_edge_retracted","node_id":"","occurred_at":"2026-09-07T15:22:04.118Z","occurred_claimed_at":"","payload":{"edge_kind":"reachable","from_graph_node_id":"gn_01m1y2whfh9x2b4c7d1e8f0a3b","graph_edge_id":"ge_01m1y2whfh62ej11jf4x5gjzv4","reason":"Reachability was inferred from a single ICMP reply that a second observation contradicted; the edge is withdrawn, its endpoints are unchanged.","source_event_id":"evt_01m1y2whfhp17g0avdqztd2p3x","to_graph_node_id":"gn_01m1y2whfhdc01srv4x8mc5a0g"},"prev_hash":"4c81f6a0d92b57e13c7a4b08fd26e5910b3c7a48d2e6f1093ba5c8d7e4026f1b","recorded_at":"2026-09-07T15:22:04.140Z","run_id":"run_01m1y2whfhnjx2am9103w0pnqw","seq":4313,"task_id":"","untrusted":true}
```

Rejections (A1-7.5): a capped field under mechanism R is `summary_too_large`, a
reused dedup key with different content is `conflict`. Note the correlation
attrs of A0-3.6 — `node_id` is the remote agent node, never a graph node.

```json
{
  "error": {
    "kind": "summary_too_large",
    "message": "events.NewEvent: append rejected for engagement=eng_01m1y2whfhgbz06ays6dxnvyws kind=command_executed task=task_01m1y2whfh1txm57x8dn41r9hg: field=command cap=2048 actual=3194 bytes",
    "attrs": {
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
      "job_id": "job_01m1y2whfhbt69j0h0fbxepw90"
    }
  }
}
```

```json
{
  "error": {
    "kind": "conflict",
    "message": "events.Append: idempotency key reused with different payload for engagement=eng_01m1y2whfhgbz06ays6dxnvyws task=task_01m1y2whfh1txm57x8dn41r9hg: key=tsk7f3q-cmd-00042 original_event=evt_01m1y2whfhp17g0avdqztd2p3x",
    "attrs": {
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
      "job_id": "job_01m1y2whfhbt69j0h0fbxepw90"
    }
  }
}
```

The integrity metadata block every customer-facing export carries (A1-6.6), and
the same block for an overridden export — the artifact names the human decision
that released it:

```json
{
  "chain_spec": "sleipnir/chain/v1",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "head_seq": 4313,
  "head_hash": "3710d121becfde25268d96bcf9f47f6618bd028842ea8324596fd66d58397647",
  "integrity_state": "verified",
  "verified_at": "2026-09-07T15:30:00.000Z",
  "chain_verified_event_id": "evt_01m1y2whfhs4v7x0z3b6d9f2hk",
  "integrity_override_event_id": ""
}
```

```json
{
  "chain_spec": "sleipnir/chain/v1",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "head_seq": 4502,
  "head_hash": "3359ec51ec7c3c565c0dc151b15d05591e377f9e7e49584ea827d18106325135",
  "integrity_state": "failed_overridden",
  "verified_at": "2026-09-07T18:02:11.400Z",
  "chain_verified_event_id": "",
  "integrity_override_event_id": "evt_01m1y2whfhr2t5v8x1z4b7d0f2"
}
```

_The two blocks are the two states, not one timeline: the first is a report
released after a clean walk, the second the same engagement after the 18:02 walk
found the break at `seq` 4127 and an admin overrode it. With `integrity_state:
failed_overridden` the artifact additionally renders the literal words "integrity
verification failed" plus the override `reason` and the overriding user's id
(A1-6.6, Q11). `usr_` is pending the A0 amendment request of §6.2._

Chain-head state (A1-5.6) is internal bookkeeping, never served and never
hashed; shown only so the field names are unambiguous for the store seam.
`verified_at`/`verified_head_hash` are the **last successful** walk (15:30, head
`seq` 4313); the 18:02 walk failed, so `integrity_state` moved to `failed` and
the 18:40 override to `overridden` (A1-6.4/6.5). `event_count` = `head_seq + 1`
because `seq` starts at 0 and is dense (A1-5.4) — that equality is the
`row_count_mismatch` check of A1-6.3:

```json
{
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "chain_spec": "sleipnir/chain/v1",
  "head_seq": 4502,
  "head_hash": "3359ec51ec7c3c565c0dc151b15d05591e377f9e7e49584ea827d18106325135",
  "event_count": 4503,
  "integrity_state": "overridden",
  "verified_at": "2026-09-07T15:30:00.000Z",
  "verified_head_hash": "3710d121becfde25268d96bcf9f47f6618bd028842ea8324596fd66d58397647"
}
```

### 4.3 Normative chain vector

Normative, like A0-2.17: the shared contract-test suite MUST reproduce these
bytes and digests byte-exactly. Four events of one engagement chain, built
with the A1-5.5 formula — genesis, a worker's `command_executed`, a
platform-composed `graph_node_quarantined`, and a worker's `task_result` that
locks the A0-2.6 number bound and the literal U+2028/U+2029/U+007F/non-BMP
bytes. The **preimage** column is the exact
byte string SHA-256 was computed over (A1-5.2: canonical JSON, the three
excluded names absent); the **served** column is the A1-1.8 representation's
length and digest, i.e. the preimage plus `seq`, `prev_hash` and `hash`
re-canonicalized.

| # | kind | preimage len | preimage bytes (exact) | `hash` | served len | served SHA-256 |
|---|---|---|---|---|---|---|
| 0 | `chain_genesis` | 428 | see below | `d65ade155302e33cf80bad7b9cbd98833373a7eed30c166510410eee0d6cfb4a` | 589 | `137b750dbc31c301ad87a5cc0406850ca24d887614cd589bf2d5781a36c80ad2` |
| 1 | `command_executed` | 816 | see below | `a564115554e9764d74195f835155a19310c065553279c19c082fbbe726b6f60c` | 977 | `752bd8344fdb852e15105081f09d61dc3683336832f20c1324be81fafa40d2f2` |
| 2 | `graph_node_quarantined` | 715 | see below | `05a06eabc89552dd1798ac918c65f10ab2c8c778cbda08d626929b6770785317` | 876 | `9d55ad66afe6e699012ed8a8fbd4f3859dc018358f2b81dfa49cadffd35a4e92` |
| 3 | `task_result` | 702 | see below | `1cae22e3bba2c7cf2b15fc920aadbd1c45f18ce07675680c96a4ce46538215c1` | 863 | `0c79020c4ebce6315d2d29eeef75b3d301e929769de719804a078f730217b614` |

```
seq 0 preimage:
{"actor":{"component":"event_store","principal_id":"","type":"platform"},"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","event_id":"evt_01m1y2whfhb2c4d6f8h0j1k3m5","evidence_refs":[],"job_id":"","kind":"chain_genesis","node_id":"","occurred_at":"2026-09-07T13:00:00.000Z","occurred_claimed_at":"","payload":{"chain_spec":"sleipnir/chain/v1"},"recorded_at":"2026-09-07T13:00:00.012Z","run_id":"","task_id":"","untrusted":false}

seq 1 preimage:
{"actor":{"component":"","principal_id":"task_01m1y2whfh1txm57x8dn41r9hg","type":"worker"},"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","event_id":"evt_01m1y2whfhp17g0avdqztd2p3x","evidence_refs":["evi_01m1y2whfh3ca875z2x8v8h7qt"],"job_id":"job_01m1y2whfhbt69j0h0fbxepw90","kind":"command_executed","node_id":"","occurred_at":"2026-09-07T14:03:22.481Z","occurred_claimed_at":"2026-09-07T14:03:19.900Z","payload":{"command":"nmap -sV -p 445 10.20.0.14","duration_ms":48210,"exit_code":0,"output_bytes":18432,"output_evidence_id":"evi_01m1y2whfh3ca875z2x8v8h7qt","redacted":false,"target":"10.20.0.14","tool_id":"tool_01m1y2whfhfjdvwqp9pfxqekmf","tool_version":"1.4.2"},"recorded_at":"2026-09-07T14:03:22.502Z","run_id":"run_01m1y2whfhnjx2am9103w0pnqw","task_id":"task_01m1y2whfh1txm57x8dn41r9hg","untrusted":true}

seq 2 preimage:
{"actor":{"component":"graph","principal_id":"","type":"platform"},"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","event_id":"evt_01m1y2whfhk4m2nq8x7z1vb3rt","evidence_refs":[],"job_id":"job_01m1y2whfhbt69j0h0fbxepw90","kind":"graph_node_quarantined","node_id":"","occurred_at":"2026-09-07T14:11:52.007Z","occurred_claimed_at":"","payload":{"graph_node_id":"gn_01m1y2whfhq7z3m9x1c4vb8nrt","quarantine_kind":"blacklist_match","reason":"Host resolved by an in-scope DNS query and matched global blacklist range 10.99.0.0/16: recorded, never tested.","source_event_id":"evt_01m1y2whfhp17g0avdqztd2p3x"},"recorded_at":"2026-09-07T14:11:52.031Z","run_id":"run_01m1y2whfhnjx2am9103w0pnqw","task_id":"","untrusted":true}

seq 3 preimage:
{"actor":{"component":"","principal_id":"task_01m1y2whfh1txm57x8dn41r9hg","type":"worker"},"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","event_id":"evt_01m1y2whfhz8k3p5r7t9v1x3z5","evidence_refs":[],"job_id":"job_01m1y2whfhbt69j0h0fbxepw90","kind":"task_result","node_id":"","occurred_at":"2026-09-07T14:20:11.380Z","occurred_claimed_at":"2026-09-07T14:20:09.120Z","payload":{"command_count":3,"duration_ms":9007199254740991,"error_kind":"","result_summary":"Relay confirmed. Second line. DEL: done 😀","revert_event_ids":[],"status":"succeeded"},"recorded_at":"2026-09-07T14:20:11.400Z","run_id":"run_01m1y2whfhnjx2am9103w0pnqw","task_id":"task_01m1y2whfh1txm57x8dn41r9hg","untrusted":true}
```

_Legend for the fenced block above (finding I-02): the `result_summary` of row 3
carries four code points that are invisible or ambiguous in a terminal. They are
written as the **literal characters**, not escapes, because A0-2.7 requires it:
U+2028 LINE SEPARATOR = bytes `E2 80 A8` (3 B) · U+2029 PARAGRAPH SEPARATOR =
bytes `E2 80 A9` (3 B) · U+007F DELETE = byte `7F` (1 B) · U+1F600 GRINNING FACE
= bytes `F0 9F 98 80` (4 B). A raw U+2028/U+2029 makes a file's line count
reader-dependent (`str.splitlines` and several editors break on them; `grep`
and `awk` do not), so **the length and digest columns of the table are the
authority, never the rendered line count**. Where this document prints these
code points inside a table cell it writes the placeholder forms `<2028>`,
`<2029>`, `<7F>` with this byte legend._


The four served envelopes (A1-1.8), pretty-printed here and canonical on the
wire — `prev_hash` of each row is the `hash` of the row before it, and row 0's
is `ChainZero` (A1-5.5):

```json
{
  "event_id": "evt_01m1y2whfhb2c4d6f8h0j1k3m5",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "run_id": "",
  "job_id": "",
  "task_id": "",
  "node_id": "",
  "occurred_at": "2026-09-07T13:00:00.000Z",
  "occurred_claimed_at": "",
  "recorded_at": "2026-09-07T13:00:00.012Z",
  "actor": { "type": "platform", "principal_id": "", "component": "event_store" },
  "kind": "chain_genesis",
  "payload": { "chain_spec": "sleipnir/chain/v1" },
  "evidence_refs": [],
  "untrusted": false,
  "seq": 0,
  "prev_hash": "0000000000000000000000000000000000000000000000000000000000000000",
  "hash": "d65ade155302e33cf80bad7b9cbd98833373a7eed30c166510410eee0d6cfb4a"
}
```

```json
{
  "event_id": "evt_01m1y2whfhp17g0avdqztd2p3x",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
  "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
  "task_id": "task_01m1y2whfh1txm57x8dn41r9hg",
  "node_id": "",
  "occurred_at": "2026-09-07T14:03:22.481Z",
  "occurred_claimed_at": "2026-09-07T14:03:19.900Z",
  "recorded_at": "2026-09-07T14:03:22.502Z",
  "actor": { "type": "worker", "principal_id": "task_01m1y2whfh1txm57x8dn41r9hg", "component": "" },
  "kind": "command_executed",
  "payload": {
    "command": "nmap -sV -p 445 10.20.0.14",
    "target": "10.20.0.14",
    "tool_id": "tool_01m1y2whfhfjdvwqp9pfxqekmf",
    "tool_version": "1.4.2",
    "exit_code": 0,
    "duration_ms": 48210,
    "output_bytes": 18432,
    "output_evidence_id": "evi_01m1y2whfh3ca875z2x8v8h7qt",
    "redacted": false
  },
  "evidence_refs": ["evi_01m1y2whfh3ca875z2x8v8h7qt"],
  "untrusted": true,
  "seq": 1,
  "prev_hash": "d65ade155302e33cf80bad7b9cbd98833373a7eed30c166510410eee0d6cfb4a",
  "hash": "a564115554e9764d74195f835155a19310c065553279c19c082fbbe726b6f60c"
}
```

```json
{
  "event_id": "evt_01m1y2whfhk4m2nq8x7z1vb3rt",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
  "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
  "task_id": "",
  "node_id": "",
  "occurred_at": "2026-09-07T14:11:52.007Z",
  "occurred_claimed_at": "",
  "recorded_at": "2026-09-07T14:11:52.031Z",
  "actor": { "type": "platform", "principal_id": "", "component": "graph" },
  "kind": "graph_node_quarantined",
  "payload": {
    "graph_node_id": "gn_01m1y2whfhq7z3m9x1c4vb8nrt",
    "quarantine_kind": "blacklist_match",
    "reason": "Host resolved by an in-scope DNS query and matched global blacklist range 10.99.0.0/16: recorded, never tested.",
    "source_event_id": "evt_01m1y2whfhp17g0avdqztd2p3x"
  },
  "evidence_refs": [],
  "untrusted": true,
  "seq": 2,
  "prev_hash": "a564115554e9764d74195f835155a19310c065553279c19c082fbbe726b6f60c",
  "hash": "05a06eabc89552dd1798ac918c65f10ab2c8c778cbda08d626929b6770785317"
}
```

```json
{
  "actor": {
    "component": "",
    "principal_id": "task_01m1y2whfh1txm57x8dn41r9hg",
    "type": "worker"
  },
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "event_id": "evt_01m1y2whfhz8k3p5r7t9v1x3z5",
  "evidence_refs": [],
  "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
  "kind": "task_result",
  "node_id": "",
  "occurred_at": "2026-09-07T14:20:11.380Z",
  "occurred_claimed_at": "2026-09-07T14:20:09.120Z",
  "payload": {
    "command_count": 3,
    "duration_ms": 9007199254740991,
    "error_kind": "",
    "result_summary": "Relay confirmed. Second line. DEL: done 😀",
    "revert_event_ids": [],
    "status": "succeeded"
  },
  "recorded_at": "2026-09-07T14:20:11.400Z",
  "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
  "task_id": "task_01m1y2whfh1txm57x8dn41r9hg",
  "untrusted": true,
  "seq": 3,
  "prev_hash": "05a06eabc89552dd1798ac918c65f10ab2c8c778cbda08d626929b6770785317",
  "hash": "1cae22e3bba2c7cf2b15fc920aadbd1c45f18ce07675680c96a4ce46538215c1"
}
```

_Row 3's `result_summary` contains the same four literal code points as its
preimage (see the legend above): U+2028 = `E2 80 A8`, U+2029 = `E2 80 A9`,
U+007F = `7F`, U+1F600 = `F0 9F 98 80`. The served length (863) and digest
(`0c79020c…`) are the authority; the rendered block above is not
line-count-stable. `duration_ms` = 9007199254740991 = 2^53−1, the A0-2.6
maximum (16 digits) — the point of the row: it locks a number's literal token
text through preimage → served → preimage (A0-2.5, F-01). A1-4.2 declares no
upper bound for `duration_ms`; the 16-digit value is deliberately implausible:
it exercises the A0-2.6 bound and F-01's literal-text rule._


_These are the four rows the A1-5.9 tamper matrix mutates (rows 0–2 are the
three-event clean chain T1–T17 name; row 3 extends it). Note what the vector
locks: `event_id` sorts before `evidence_refs` (UTF-8 byte order, A0-2.4), the
`actor` and `payload` objects are canonicalized recursively, `""` and `[]` and
`false` are present rather than absent (A1-1.2), and `seq`/`prev_hash`/`hash`
are absent from the preimage but present in the served form (A1-5.2, A1-1.8)._

### 4.4 Contract tests (A1)

The ids the shared contract-test suite MUST implement for A1, per clause. Each
safety rule carries a positive and a negative id (AGENTS.md); the clause that
owns the rule also names them, this subsection is the index.

| Clause | Test ids |
|---|---|
| A1-1.2 / A1-4.8 | `TestNilCollectionNeverSerializesAsNull`, `TestPayloadsAreFlat` |
| A1-2.1 | `TestActorComponentIsEmptyForNonPlatform`, `TestActorVocabularyMatchesA2` |
| A1-2.3 | `TestClientCannotSupplyEnvelopeFields`, `TestUntrustedFlagCannotBeSupplied` |
| A1-2.6 / A1-2.7 | `TestActorCannotBeForged`, `TestOrchestratorCannotClaimCommandExecuted`, `TestMachinePrincipalCannotReachIntegrityKinds` |
| A1-3.1 / A1-3.5 | `TestKindListIs42AndClosed`, `TestUnknownKindIsRejectedOnWrite` |
| A1-3.4 / A1-7.4 | `TestWorkerCannotAppendNonCKind` |
| A1-4.2 | `TestApprovalFingerprintAndExpiryAreEqual`, `TestApprovalMetadataMismatchIsConflict`, `TestActionSpecArtifactIsChained`, `TestSingleUseApprovalRaceConsumesOnce`, `TestConsumedApprovalCannotSpawnAgain`, `TestContentHashMatchesGraphFingerprint`, `TestDedupCollapseStillEmitsEvent` |
| A1-4.5 | `TestMaximalPayloadFitsCanonicalBound`, `TestCapsRejectWithSummaryTooLarge` |
| A1-4.6 | `TestRedactedIsPlatformSetOnly`, `TestRedactedNotUsedForTruncation` |
| A1-4.7 | `TestArraysSortedDedupedAtComposition`, `TestEvidenceRefsDerivation` |
| A1-4.9 | `TestEventSecretFreeSerialization`, `TestSecretScanNamesFieldNotValue`, `TestNoSecretValueOrDigestInError` |
| A1-4.10 | `TestServedEventIgnoresUnknownFields`, `TestAppendRejectsUnknownField` |
| A1-4.11 / A1-4.12 | `TestNoTimeTimeInCanonicalizedTypes`, `TestUntrustedContextComputedNotSupplied`, `TestNoUntrustedTextInUnmarkedFields`, `TestUntrustedFlagMatchesStarredFields` |
| A1-5.3 | `TestSecondGenesisIsRejected`, `TestAppendRefusedWithoutValidGenesis` |
| A1-5.4 | `TestRecordedAtClampIsMonotone`, `TestRecordedAtClampBeyondBoundIsLogged`, `TestFailedAppendConsumesNoSeq`, `TestNoGlobalSequenceSharedAcrossEngagements`, `TestConcurrentAppendsProduceContiguousSeq`, `TestConcurrentAppendsAcrossEngagements` |
| A1-5.7 | `TestServedBytesReproducible`, `TestServedRoundTripIsIdentity`, `TestEventRoundTrip` |
| A1-5.8 | `TestHeadRegressionDetected`, `TestHeadTrailIsAppendOnly`, `TestHeadAnchorWebhookCarriesTuple`, `TestWebhookAnchorFailureNeverBlocksAppend`, `TestAnchorDeliveryNeverTriggersAnotherAnchor` |
| A1-5.9 | `TestChainVectorDigests` (the §4.3 vector, byte-exact, rows 0–3) |
| A1-6.2 | `TestReadDuringStartupWalkIsFlaggedUnverified`, `TestExportFromUnverifiedChainRefused`, `TestAppendRefusedBeforeStartupVerificationCompletes` |
| A1-6.3 | `TestOneBreakEventPerRun`, `TestBreakDedupStateWrittenInSameTransaction` |
| A1-6.4 | `TestExportBlockedOnFailedChain`, `TestInternalViewOverrideDoesNotAuthorizeExport`, `TestAppendsContinueOnFailedChain` |
| A1-6.5 | `TestOverrideIsAdminOnly`, `TestOperatorCannotOverrideAnyBreak`, `TestEveryReleaseUnderOneOverrideIsChained`, `TestReleaseWithoutLiveOverrideRefused`, `TestOverrideDiesAtNextBreak`, `TestArtifactReleasedNamesItsOverride` |
| A1-6.6 | `TestExportComposesArtifactReleased`, `TestExportIntegrityStateIsNotTheChainStateEnum` |
| A1-6.8 | `TestVerificationIdempotent`, `TestVerificationWritesNoOtherKind` |
| A1-7.6 | `TestIdempotencyKeyReuseWithDifferentPayloadIsConflict`, `TestRetryCannotShiftClaimedTime`, `TestPayloadHashVector`, `TestDedupHitWritesNoEvent`, `TestPlatformDedupKeyIsNonEmptyPerKind`, `TestPlatformDedupKeyEmptyRejected` |
| A1-7.10 | `TestValidationOrderIsDeterministic` |
| A1-7.11 | `TestRetryReusesIdempotencyKey` |
| A1-7.12 | `TestKillPathDoesNotBlockOnEventStore`, `TestHardStopProceedsWhenEventStoreUnavailable`, `TestKillPathDoesNotBlockOnAppendLatency`, `TestKillOutboxIsDrainedAtStartup`, `TestKillOutboxIsAppendOnly` |
| A1-8.2 | `TestCursorFromEngagementARejectedInB`, `TestCursorWithInconsistentKAndIDRejected` |
| A1-8.4 | `TestWorkerAndNodeHaveNoReadScope`, `TestSSENotReachableByMachinePrincipal` |
| A1-8.5 | `TestSSEEmitAfterCommitInSeqOrder`, `TestSSEReplayDedupesByEventID`, `TestSSEResumeFromForeignSeqNeverServesUnrelatedPosition` |
| A1-8.6 (the six cross-engagement negatives) | `TestEventIDFromAIsNotFoundInB`, `TestEventEngagementBReadNeverReturnsA`, `TestFilterWithForeignIDReturnsNothing`, `TestNoBulkEventReadSpansEngagements`, `TestSSEStreamNeverCarriesAnotherEngagement`, `TestCursorFromEngagementARejectedInB` |
| A0-2.15 (via A1-5.5) | `TestGatingComparisonsAreConstantTime` |

_The six ids on the A1-7.4/A1-2.x write path (`TestWorkerCannotAppendNonCKind`,
`TestClientCannotSupplyEnvelopeFields`, `TestActorCannotBeForged`,
`TestOrchestratorCannotClaimCommandExecuted`,
`TestMachinePrincipalCannotReachIntegrityKinds`,
`TestUntrustedFlagCannotBeSupplied`) are the negative half of AGENTS.md's
safety-test rule for the write path; A1 does not reach `Frozen` without them
(E-01)._

## 5. Traceability

| Source | Decision | Clauses |
|---|---|---|
| **Q6 + worker addendum** | machine principals never mutate scope/blacklist, never decide approvals, never touch kill controls; three enforcement layers (scopes → middleware → per-request engagement/run binding); worker is **report-only** (`events:append`, `evidence:upload`, `task:result`, no read scope) | A1-2.3, A1-2.5–2.7, A1-7.1, A1-7.4 (principal × kind matrix), A1-7.5 (`forbidden` rows), A1-8.4 (no read scope for worker/node), A1-6.5 (no machine override), A1-6.8 (no machine-triggered verification) |
| **Q10** | approvals are uniformly single-use; one approval = exactly one execution; consumption is recorded | A1-3.3 (`approval_granted.single_use`, `approval_executed.single_use_consumed`), A1-4.2 (fingerprint equality across request/decision/execution), A1-3.3 (`action_blocked{approval_consumed}`), A1-1.4 (single-use checks use platform time only, A0-5.6) |
| **Q11** | hash chaining from day one on the write side; per-engagement chains with genesis events; verification at startup and before export; on failure exports blocked, internal views flagged, explicit operator override logged as an event, report stamped; no external timestamp anchoring | A1-5.1 (per engagement), A1-5.2 (digest + exclusion list), A1-5.3 (genesis, `chain_spec`), A1-5.4 (`seq`), A1-5.5 (linking formula), A1-5.6 (chain-head state), A1-5.7 (stored preimage), A1-5.8 (declared limits + out-of-band head trail), A1-6.1–6.6 (walk, triggers, break, block, override, export stamp), A1-6.9 (no repair), A1-3.3 (the four integrity kinds), §4.3 (normative vector) |
| **Q12** | full `/api/v1` from day one; HTMX UI + JSON API + SSE | A1-8.5 (SSE mapping: `id:` = `seq`, `data:` = canonical envelope, emit-after-commit, at-least-once, engagement-scoped), A1-8.1–8.3 (order, cursors, closed filters), A1-8.9 (consumer guarantees), A1-7.8 (batch endpoint is additive/A4) |
| **Q13** | additive-only `/api/v1`; clients ignore unknown fields; breaking changes → `/api/v2` | A1-3.1/A1-3.5 (closed taxonomy, kinds are the additive unit), A1-4.10 (read/write asymmetry; a new payload field or an 18th envelope key is **breaking** because the bytes are hashed), A1-5.10 (`chain_spec` versioning), A1-4.12 (unknown enum on read preserved), A1-8.7 (no projection in v1) |
| **Q14** | spawn request carries tool id + version, a capped untrusted task description and a resource ask; the platform derives image digest and risk tier from the registry; risky spawns wait on approval | A1-3.3 (`spawn_requested`, `job_spawned`, `task_spawned`), A1-4.2 (digest/tier are registry-derived, never orchestrator-supplied), A1-3.6 (A7 owns fingerprint and `risk_tier` vocabulary), A1-4.5 (`image_digest`, `tool_version`, `task_description` caps) |
| **ADR-0009 §1** | central append-only command log: command, output reference, actor, target, timestamp | A1-3.3 (`command_executed`), A1-4.2 (exact argv, `output_evidence_id` when `output_bytes > 0`), A1-4.3 (references, never bytes), A1-7.2 (append-only), A1-1.1 (actor/timestamps in the envelope) |
| **ADR-0009 §2** | evidence artifacts stored per engagement and referenced from the log | A1-1.7 (`evi_` references), A1-3.3 (`evidence_stored`, `evidence_refs`), A1-4.2 (`sha256` = artifact-integrity digest), A1-4.9 (a digest of a credential value is forbidden), A1-7.3 (`evidence_refs` derived from the payload), A1-8.3 (`evidence_ref` filter), A1-8.8 (a dangling reference is not an integrity break) |
| **ADR-0009 §3** | revert records captured at action time for every state-changing effect | A1-3.3 (`revert_recorded`, client-appendable), A1-4.2 (`revertable:false` → carried into the plan), A1-1.6 (the revert record **is** an event, identified by its `event_id`) |
| **ADR-0009 §4** | cleanup plan assembled from revert records, human-approved, executed then verified; non-revertable effects documented | A1-3.3 (`cleanup_planned`/`_executed`/`_verified`), A1-4.2 (`approval_id` mandatory on the plan; `non_revertable_event_ids`), A1-8.8 (log retention is **not** client-environment cleanup) |
| **ADR-0016 §1/§2** | graph content is evidence with provenance; scope alignment enforced on the graph; out-of-scope discoveries quarantined, never actionable | A1-2.8 (provenance = `event_id` + `actor`), A1-3.6 (field names A2 must reference), A1-3.3 (`graph_node_written`, `graph_edge_written`, `graph_node_quarantined`, `quarantine_recomputed`), A1-4.2 (`source_event_id` non-empty), A1-3.8 (A2 answers) |
| **ADR-0016 §4** | revision by a new node + `supersedes`, history never overwritten; self-correction via `contradicts` | A1-3.3 (`supersedes_graph_node_id`), A1-3.8 (**corrected**: no `node_superseded` kind — one fact, one encoding), A1-3.3 (`graph_edge_retracted` for A2-3.9), A1-7.2 (annotation is a new event) |
| **ADR-0017 §2–§3** | only the platform talks to the runtime; orchestrators request workers through the spawn broker; the platform owns lifecycle, quotas, image allowlist, hard stop | A1-7.1 (platform-only composition), A1-7.4 (orchestrator appends nothing; its occurrences are broker-composed with the orchestrator as `actor`), A1-3.3 (`spawn_requested`/`job_spawned`/`task_spawned`/`container_killed`), A1-4.2 (registry-derived digest), A1-3.3 (`action_blocked{image_not_allowed, quota_exceeded, node_not_paired}`) |
| **ADR-0018 §1–§4** | approvals bind to an action fingerprint; agent prose is *in addition*; execution-time re-validation; fingerprint and action recorded together; untrusted-context flag | A1-3.3 (`approval_requested.fingerprint_hash`, `untrusted_context`, `approval_executed.revalidated`), A1-4.2 (fingerprint equality; `action_summary` is prose, not the binding), A1-4.5 (mechanism R on `action_summary`/`target`: an approver is never shown a silently shortened action), A1-3.6 (A7 owns fingerprint content), A1-4.4 (untrusted marking feeds the ADR-0018 §4 UI flag), A1-3.3 (`action_blocked{fingerprint_mismatch, approval_expired_at_exec}`) |
| **ADR-0019 §2–§3** | self-contained greppable errors; origin function; correlation ids; log-or-return once | A1-7.5 (message rules, one error per write), A1-3.3 (`agent_error.origin` = `component.Function`), A1-2.4 (`component` closed enum so a reader can attribute without a code search), A1-6.3 (log at error level with attrs), A1-5.8 (head hash emitted to `slog`), A0-3.8 cited |
| **ADR-0019 §5** | redaction is mandatory: no credentials, tokens or secret material in errors or logs | A1-4.9 (no secret values in a payload; ingest scan; negative tests), A1-4.3 (references only), A1-3.3 (`endpoint_name`/`target_name` are names, never URLs), A1-7.5 (a rejection message never echoes the value), A1-4.6 (`redacted` marker) |
| **ADR-0020 §1–§2** | per-engagement LLM data policy enforced by the gateway; the gateway records egress **metadata** only | A1-3.3 (`run_started.llm_data_policy`, `llm_call` metadata fields, `egress_policy`), A1-4.2 (`llm_call` MUST NOT carry prompt/completion text), A1-4.5 (`model_name`/`endpoint_name` caps) |
| **ADR-0020 §3–§4** | per-run pseudonymization mapping is secret; captured credentials/hashes/tokens never reach a cloud endpoint under any policy | A1-4.9 (mapping values are secret material; exclusion at the gateway is defence in depth), A1-3.3 (`llm_call.masked_entity_count`, `excluded_secret_count`), A1-3.3 (`action_blocked{secret_excluded, llm_egress_blocked}`), A2-9 cross-referenced |
| **ADR-0020 §5** | operator-visible egress log; residual masking risk declared | A1-3.3 (`llm_call` is the egress log, one event per call), A1-8.3 (filterable by `kind`), A1-4.6 (`redacted` exists from day one) |
| **ADR-0005 / C5** | allowlist scope, blacklist beats allowlist beats approval, enforcement in the platform core, a denied action is auditable, hard stop is platform-owned and respawn-proof | A1-3.3 (`scope_changed`, `scope_denied`, `blacklist_denied`, `hard_stop_fired`, `container_killed{hard_stop}`, `action_blocked{hard_stop_active}`), A1-4.2 (`entry_hash`; blacklist entry matched), A1-7.12 (kill paths never block on the store), A1-7.1 (enforcement is platform-side) |
| **ADR-0012 §1–§2/§6–§7** | only assigned operators/admins approve; approver identity + timestamp in the log; six notification kinds; retries and delivery are audited; 2 h expiry → orchestrator replans | A1-3.3 (`approval_granted`/`_denied` with the envelope `actor` as approver identity, `approval_expired.expires_at`/`queue_wait_ms`, `notification_sent`), A1-2.1–2.2 (typed actor), A1-1.4 (platform time is authoritative for expiry), A1-4.2 (notification vocabulary is ADR-0012's, not A1's), A1-3.3 (`agent_error` as a notification kind; `chain_head_anchor` as A1's seventh value for the A1-5.8 (4) anchor, product owner decision 2026-09-21, PR #2 item D5) |
| **ADR-0013 / Q9** | remote agent nodes over a mesh/tunnel, buffering when offline; node identity `slp_node_` | A1-1.1 (`node_id`), A1-1.4/A0-5.7 (`occurred_claimed_at` for buffered events), A1-7.6 (dedup key makes replay safe), A1-7.11 (buffer and replay, never drop), A1-7.4 (`node` principals append the three C kinds) |
| **ADR-0011** | one core, two faces (HTMX + JSON/SSE); rate limiting is part of the attack surface | A1-8.5 (SSE mapping), A1-8.7 (no projection), A1-6.8 (`on_demand` verification must be rate-limited, A0-3.1 `rate_limited`), A1-7.5 (one error envelope, A0-3.10 for the UI face) |
| **adversarial A11** | evidence tampering / chain of custody: a compromised component could rewrite history | A1-5.2–5.7 (chainable set incl. `engagement_id`, `actor` and all three timestamps; stored preimage), A1-5.8 (what the chain proves; out-of-band head trail), A1-5.9 (tamper matrix T1–T17), A1-6 (detect, block, override, stamp), A1-6.7 (third-party re-verification bundle), A1-7.2 (no update/delete on the seam), A1-7.9 (DB-level `REVOKE` recommendation), A1-4.6 (`redacted` from day one) |
| **adversarial A12** | cross-engagement leakage through implementation bugs | A1-2.5 (a foreign engagement is `notfound`, never `forbidden`), A1-5.1 (one chain per engagement), A1-5.2 (`engagement_id` is chainable → a stolen event cannot be re-labelled), A1-5.9 T6–T8 (splicing tests, `engagement_mismatch`), A1-7.10 (no multi-engagement write method), A1-8.2 (a cursor must resolve in the requested engagement), A1-8.4–8.6 (scoped reads + five negative tests) |
| **adversarial A1** | prompt injection: untrusted content must not become configuration, and must not be able to erase its own trace | A1-4.4 (untrusted marking + never configuration + no laundering), A1-1.1/A1-5.2 (`untrusted` is inside the digest, so it cannot be flipped), A1-7.5/A1-3.3 (a rejected append is still chained as `action_blocked{append_rejected}`), A1-4.9 (secret scan), A1-8.5 (a stream carries the flag to every consumer) |
| **adversarial A9** | classic web attacks on our own surface: hand-rolled auth, XSS via rendered evidence, SSRF via webhook URLs | A1-4.4 (autoescape every string field, including human-typed prose), A1-4.3 (no artifact bytes in a payload → nothing to render raw), A1-3.3 (`target_name`/`endpoint_name`: a URL is never stored), A1-8.4 (no read scope for machine principals), §6.9 (user/session audit events are a gap) |
| **SPEC §5 steps 1–10** | the normative job lifecycle must be fully covered by the taxonomy | A1-3.7 (step → kinds map, incl. the three added kinds), A1-3.1 (closed and complete on day one), A1-3.5 (a new kind only when independently filterable) |
| **SPEC §6** | tool output and model prose are untrusted input, never configuration; out-of-scope discoveries are quarantined; enforcement is in the platform core | A1-4.4, A1-4.9, A1-7.1, A1-7.10, A1-3.3 (`graph_node_quarantined`, `quarantine_recomputed`, `scope_denied`, `blacklist_denied`) |
| **SPEC §7** | every run snapshots its exact model config into the log (reproducibility); gateway is the single logging point | A1-3.3 (`model_config_snapshotted`, `run_started`, `llm_call`), A1-4.2 (`scope_snapshot_evidence_id` non-empty), A1-4.3 (config snapshot is an artifact reference) |
| **SPEC §8** | the event log is the append-only spine; the graph and the evidence store hang off it | A1-7.2, A1-7.7 (ingest watermarks on `seq`/`event_id`), A1-2.8/A1-3.6 (provenance seam to A2), A1-1.7 (evidence references), A1-8.9 (consumer guarantees) |
| **SPEC C9** | errors and logs descriptive and machine-traceable | A1-2.4, A1-3.3 (`agent_error`), A1-6.3, A1-7.5 |
| **SPEC C11** | all model traffic through the gateway; captured secrets never leave | A1-3.3 (`llm_call`), A1-4.9, A1-4.3 |
| **SPEC C1 / ADR-0001 / ADR-0010** | Go stdlib only, pgx behind the store seam | A1-5.2 (`crypto/sha256` + `encoding/hex` via `internal/cjson`, A0-2.15), A0-1.1 ids from `crypto/rand`, A1-5.4 (`SELECT … FOR UPDATE` / `pg_advisory_xact_lock` — both pgx-reachable, no third-party module), §4 (no third-party type anywhere in A1) |
| **A0-2.12** | a canonicalized type declares its exclusion list: plain top-level names only | A1-1.3 (the three chain fields are top-level), A1-5.2 (exactly `hash`, `prev_hash`, `seq`), §4.1 `ChainExclude`, A1-5.9 T17 |
| **A0-2.13** | an exclusion list is part of the digest's definition; after Freeze it changes only by a new ADR | A1-5.2 (declared immutable), A1-5.10 (algorithm/form/exclusion change = breaking, `/api/v2` + ADR), A1-5.9 T17 (a different exclusion set is a test case) |
| **A0-2.14** | a canonicalized type serializes a fixed key set, zero values instead of absence | A1-1.2 (17 keys always present), A1-4.1 (per-kind payload key set), A1-3.2, §4.1 (no `omitempty` on `Event`), §4.3 (vector shows `""`/`[]`/`false`) |
| **A0-2.16** | persist the exact canonical bytes and verify from them — "A1 must provide the column" | A1-5.7 (preimage column + no second JSON copy + served form derived), A1-1.8, A1-6.1 (recompute, never re-marshal), A1-6.7 (the bundle ships those bytes), A1-5.9 T16 |
| **A0-3.11** | a non-idempotent write may only be retried where the owning contract defines a dedup key — "A1 MUST define one for `events:append`" | A1-7.6 (the key, its enforcement and its semantics), A1-7.3 (`idempotency_key` in the request), A1-7.7 (replay guarantees), A1-7.5 (`conflict` on key reuse, `timeout` retried with the same key), A1-3.8 (the answer A2-4.7 needs) |
| **A0-4.3** | each paginated collection declares a total, stable order from an immutable unique key | A1-8.1 (`seq` ascending; descending only as a second declared order), A1-8.2 (`k` = `seq`, `id` = `event_id`), A1-8.3 (filters never reorder), A1-5.4 (`seq` is immutable, dense, unique) |
| **A0-5.6** | platform-recorded time is authoritative for ordering, verification, expiry and single-use checks | A1-1.4, A1-5.4 (`recorded_at` clamped non-decreasing), A1-6.1 (verification uses stored bytes and `seq`, not any client time), A1-8.1 (`occurred_claimed_at` orders nothing), A1-4.2 (`expires_at` platform-computed) |
| **A0-5.7** | a client-supplied timestamp lives in a `*_claimed_at` field, never drives ordering/expiry/digests, and is stored next to `recorded_at` | A1-1.1 (`occurred_claimed_at`), A1-1.4, A1-4.12 (no payload timestamp is ever a claimed one), A1-5.2 (it **is** hashed, so it cannot be edited afterwards), A1-7.3 (the only client time in an append), A1-5.9 T3 |
| **Adversarial T-01/T-02** | engagement lifecycle and artifact release are chainable facts, not side effects | A1-3.3 (`engagement_created` at `seq` 1, `engagement_closed`, `artifact_released`), A1-3.5 (additive), A1-3.7 (SPEC §5 steps 1 and 9), A1-4.2 (the three obligations rows), A1-6.6 (the export path composes `artifact_released`), §4.1 (three new `Kind` consts, 42 total), A1-4.5 (`TestMaximalPayloadFitsCanonicalBound` over 42 kinds), A1-4.4 (`TestKindListIs42AndClosed`) |
| **Adversarial C-01/S-07 (§6 item 15, ADR-0021)** | override authority is one unambiguous rule (admin-only), and the override's lifetime is a product-owner decision: not single-use, the residual risk accepted by the service owner and compensated by chained attribution | A1-6.5 (authority + lifetime), A1-3.3 (`artifact_released.override_event_id`), A1-4.2 (`override_event_id` non-empty iff `failed_overridden`), A1-6.6 (one `artifact_released` per artifact), §6 item 15, ADR-0021 |
| **Principal S-02/P-37 (§6 item 16)** | the head hash is anchored out-of-band in a store the event role cannot rewrite | A1-5.8 (`chain_head_trail`, `REVOKE UPDATE, DELETE`, 100-`seq` interval), A1-5.8 (4) (the ADR-0012 §3 signed-webhook head anchor, approved 2026-09-21), A1-3.3 (`break_kind:head_regression`), A1-3.3 (`notification_kind:chain_head_anchor`), A1-6.3 (the `head_regression` row), A1-6.2 (startup compares against the trail), A1-6.6 (residual-risk wording), §4.1 (`HeadLogIntervalSeq = 100`), §6 item 16 |
| **Principal P-13/P-32, adversarial T-07** | the `recorded_at` forward clamp is byte-wise, bounded, and its state lives on the chain head | A1-5.4 (byte-wise comparison, 1000 ms bound), A1-5.6 (`LastRecordedAt`, `LastBreakSeq`, `LastBreakKind`, `LastBreakEventID`), A1-6.3 (break-dedup state), §4.1 (`ChainHead`) |
| **Principal P-09** | a refused append during the startup walk is retryable, not a defect | A1-7.5 (`timeout` row + the `internal` reservation), A1-7.11 (retry reuses `idempotency_key`; the one retryable write), A1-6.2 (`unverified`) |
| **Principal P-97/E-03/S-13** | the kill path is durable without being blocking | A1-7.12 (durable outbox, `REVOKE UPDATE, DELETE`, startup drain, UI warning), A1-6.4 (the integrity-warning carrier), A1-6.5 (override tests), A1-4.2 (`container_killed.stop_event_id`) |
| **Principal P-80** | a served event decodes in two passes with no `map[string]any` intermediate | §4.1 (`UnmarshalEvent`), A1-4.1 (per-kind payload type), A1-4.10 (read/write asymmetry), §4.4 (`TestEventRoundTrip`) |
| **Principal P-33/P-08** | the dedup digest is defined once and every platform-composed row has a non-empty key | A1-7.6 (`PayloadHash` definition + worked vector, the per-kind deterministic key list), A1-7.10 (normalization runs first), §4.4 (`TestPayloadHashVector`) |
| **Principal D-03** | number literal text and the literal U+2028/U+2029/U+007F/non-BMP bytes survive preimage → served → preimage | §4.3 row 3 (normative, 702 B / `1cae22e3…`, served 863 B / `0c79020c…`), A1-5.7 (`Served` = `cjson.With`, `UseNumber()`), A0-2.5/A0-2.7 (cited), A1-4.12 (`duration_ms` is `int64`) |
| **Adversarial E-01/E-04** | the write path and the genesis/startup path each carry their negative tests | A1-2.3, A1-2.6, A1-2.7, A1-3.4, A1-7.4 (six write-path negatives), A1-5.3, A1-6.2, A1-7.11 (five genesis/startup negatives), §4.4 (index) |
| **A0-6.6** | a canonicalized type evolves by adding kinds, not by reshaping fields; written canonical bytes are immutable | A1-4.10 (a new payload field or an 18th envelope key is breaking), A1-4.1, A1-3.5, A1-1.6/A1-7.2 (annotation is a new event), A1-3.8 + A1-3.3 (the two kinds added for A2 are additive) |
| **A0-6.1/6.2/6.3** | read: ignore unknown fields, preserve unknown enums; write: reject both | A1-4.10, A1-7.3, A1-7.5, A1-7.10 steps 3/5/7, A1-3.1, A1-4.12 |
| **A0-7.7** | every capped field class gets exactly one mechanism, recorded in the owning contract | A1-4.5 (mechanism table: **R for every A1 class, T for none**, with the derivation of the 32768 B document invariant), A1-3.2 (`(N)` notation), A1-7.5 (`summary_too_large`), §6.3 + AM-2 |
| **A0-7.3/7.6/7.8** | measurement is decoded UTF-8 bytes; mechanism R refuses with `summary_too_large` naming field, cap and actual; enforcement is platform-side | A1-4.5, A1-7.5, A1-7.10 step 9, A1-4.7 (array counts) |
| **A0-3.1/3.9** | closed error kinds; a foreign-engagement object is `notfound`, never `forbidden` | A1-7.5 (the full mapping table), A1-2.5, A1-4.11, A1-6.4 (`integrity_failed` = 409, never 5xx), A1-8.6 |
| **A0-8.x** | key spelling, fixed suffixes, absent-not-null, closed enum spelling, hex digests, adjective booleans, bounded bodies | A1-3.1–3.2 (kind spelling, notation), A1-4.1 (explicit `json` tags), A1-4.8 (`null` rejected), A1-4.12 (enums), A1-1.2/A0-8.3 (zero values, not absence — the canonicalized-type exception), A1-7.10 step 2 (body bound), A1-5.5 (64-char lowercase hex) |
| **A2 cross-contract requests** | envelope names, seven event kinds, the ingest dedup key | A1-3.8 (item-by-item: confirmed / corrected / added), A1-3.3 (`quarantine_recomputed`, `graph_edge_retracted`), A1-7.6/A1-7.7 (dedup key and replay guarantees) |

## 6. Open for product owner

**All items answered** — product owner, 2026-09-21, PR #2 (D1–D9 plus the 46-item confirm checklist). This section is now the decision record; the markers below cite the decision instead of requesting it.

Recommendations that were genuinely product-owner calls: naming, algorithms,
customer-visible wording, and the two places where a safety default could
reasonably go the other way. Each was marked at the clause as needing the
product owner's confirmation; none was decided silently, and each now cites the
decision. A1 is written against **current** A0 throughout — nothing below is
assumed to be already granted, and A1's own numbering of amendment requests
(AM-1…AM-4) is independent of A2's.

1. **A1-5.2 / A1-5.5 — digest and linking.** SHA-256 over the A0-2 canonical
   bytes with an exclusion list of exactly `hash`, `prev_hash`, `seq`;
   `prev_hash(N) = hash(N-1)`; genesis `prev_hash` = 64 ASCII `0`;
   `chain_spec` = the literal `"sleipnir/chain/v1"`. _Stdlib-only
   (`crypto/sha256` + `encoding/hex`), byte-reproducible, and the §4.3 vector
   locks it. A keyed hash (HMAC with a platform secret) would additionally
   resist a tamperer who knows the algorithm — but the secret would live on the
   same host as the store, so it buys nothing against the threat that matters
   (A1-5.8) and costs key management. **Decided** (product owner, 2026-09-21,
   PR #2 §6.1)._
2. **A1-2.2 — user principal id shape: RESOLVED (AM-1 granted, product owner
   decision 2026-09-21, PR #2 item D3).** A0-1.2 registers the human-principal
   prefix `usr_` (`^usr_B{26}$`, 30 B) and A0 §4 adds `KindUser`, so
   `actor.principal_id` for `type:"user"` validates per A0-1.5 and every
   user-composed kind can be stamped correctly (`approval_granted`,
   `approval_denied`, `hard_stop_fired`, `scope_changed`,
   `engagement_policy_changed`, `report_inclusion_changed`,
   `graph_node_quarantined`, `graph_edge_retracted`, `integrity_override`) —
   which is what makes A1-6.5's "name who overrode it" enforceable. This was the
   single **freeze blocker** (same request as A2 §6.2 / A2 AM-1) and it is
   closed. The rule the request carried stays normative: a username or e-mail
   MUST NOT be used instead — both are mutable and both are personal data in a
   customer export.
3. **A1-4.5 — mechanism R for every capped A1 field, T for none.** Consequence:
   an over-cap append is refused with `summary_too_large` (413) and the client
   retries with fewer bytes; nothing is ever silently shortened inside a hashed
   record. _Same posture as A2-7.1, and forced twice over: mechanism T needs a
   sibling `<field>_truncated` boolean that A1-3.3's closed key sets do not
   declare (A1-4.1), and the bulk content an agent actually produces is an
   `evi_` reference, not a payload string (A1-4.3). The operational cost is a
   413 loop for an agent that will not shorten its summary — visible as
   `action_blocked{append_rejected}` rather than silent. **Decided** (product
   owner, 2026-09-21, PR #2 §6.3; AM-2 asks A0-7.7 to name A1 alongside
   A2/A3)._
4. **A1-4.9 — secret found in an append payload: reject, not redact.**
   Recommend hard reject (`validation`, field + rule id named, value never
   echoed) with the rejection chained as `action_blocked{append_rejected}`.
   Alternative: redact the value, set `redacted:true` (A1-4.6), keep the event.
   _Reject is the Q3 posture and matches A2-9.4; redact risks a false positive
   destroying a worker's only report of what it ran, and an agent that cannot
   re-append is blind. Both need the same pattern corpus, which must ship with
   the contract tests. **Decided** (product owner, 2026-09-21, PR #2 item D8;
   aligned with A2 §6.10 — the two contracts MUST NOT answer this
   differently)._

   **Ruled once for both contracts (product owner decision 2026-09-21, PR #2
   item D8): reject, never redact.** A secret-pattern hit (A2-9.4 rule ids) is a
   hard reject — `validation` (400) naming the field and the rule id, the value
   never echoed in whole, in part or as a digest (A0-3.4) — and the rejection is
   chained (`action_blocked`). Redaction was rejected: a false positive would
   silently destroy a worker's only report of what it ran, and `redacted:true`
   (A1-4.6) means platform redaction, never rejection. The false-positive risk
   is controlled by shipping the A2-9.4 rule table with the planted-secret
   corpus. A1 §6.4 and A2 §6.10 are the same question and MUST NOT be answered
   differently.
5. **A1-6.4 — appends continue on a failed chain.** Recommend: exports blocked,
   internal views flagged, **appends keep working**, blast radius one
   engagement. Alternative: fail closed and refuse appends until the break is
   resolved. _Refusing appends stops evidence capture and hands a tamperer a
   one-row denial-of-service over the whole engagement; continuing means new
   events chain onto a head that is already suspect, which the break event and
   the export stamp both disclose. Q11 mandates only the export block.
   **Decided** (product owner, 2026-09-21, PR #2 §6.5)._
6. **A1-6.5 — override is admin-only, reasoned, and dies with the state it
   overrides.** Recommend: `user` principal with the admin role (SPEC §3, as
   A1-3.3 already spells it); non-empty `reason`; `scope:export` vs
   `internal_view` are not interchangeable; the override is valid until the next
   `chain_verified` or `chain_break_detected`, and every released artifact names
   the override event (A1-6.6). Alternatives: allow an assigned operator to
   override their own engagement (faster, weaker separation — adversarial A15
   (insider abuse)), or make an override time-boxed (needs a clock-driven expiry inside a
   safety decision). _Q11 says "explicit operator override … logged as event";
   "operator" there reads as "human", and SPEC §3 puts integrity-class controls
   with the admin, next to the hard stop. **Decided** (product owner,
   2026-09-21, PR #2 §6.6) — the wording the product owner signed is item 15._
7. **A1-5.8 — tail truncation needs an anchor outside the store: DECIDED, the
   webhook anchor ships.** Deleting the last *k* rows and updating the
   chain-head state leaves a self-consistent chain; no hash chain detects that
   from inside. v1 already mandates the internal mitigations (no update/delete
   on the seam, head hash emitted to `slog` at every verification and every
   **100** `seq` (A1-5.8, `HeadLogIntervalSeq`), head hash embedded in every
   delivered export). Recommendation:
   **also** deliver `(engagement_id, head_seq, head_hash)` over the ADR-0012 §3
   signed webhook at those same points — a customer-side or SIEM-side copy that
   a store-only attacker cannot reach, for roughly twenty lines of code and no
   new dependency. _Q11 excluded external **timestamp** anchoring; this is
   head-**hash** anchoring, and it is the only v1 mechanism that closes the
   truncation case. **Approved by the product owner 2026-09-21 (PR #2 item D5)**
   and made normative in A1-5.8 (4); item 16 records the same decision._
8. **A1-7.4 / A1-8.4 — the orchestrator gets neither `events:append` nor a log
   read scope in v1.** Every orchestrator-originated occurrence already has a
   platform-composed kind (`spawn_requested` through the broker,
   `approval_requested` through the approval service, `agent_error`,
   `scope_denied`, `blacklist_denied`, `action_blocked`), and its planning
   surface is the graph plus stage views (ADR-0016 §4, Q1) — not the log.
   _Granting a read "just for the orchestrator's own run" is a lateral channel
   into every other agent's findings and target data; granting an append lets a
   planner write prose into the audit spine that no C kind covers. A1-2.6's run
   binding still applies if A5 ever grants either. **Decided** (product owner,
   2026-09-21, PR #2 §6.8; A5 turns this into the exclusion list)._
9. **A1-3 gap — user and session audit events are not in the taxonomy.** The 42
   kinds cover the engagement/run/agent/integrity spine. They do **not** cover
   authentication and administration: login success/failure, TOTP failure,
   session creation/revocation, password change, role change, engagement
   assignment, tool-registry writes, global-blacklist writes outside an
   engagement, LLM-endpoint credential changes. Some of these have no natural
   `engagement_id`, so they do not fit a per-engagement chain (A1-5.1) at all.
   Recommendation: keep them **out of A1** and give them a platform-scoped
   audit record owned by the authn/session contract (A5) with its own
   append-only store, cross-referencing an engagement where one exists; revisit
   only if a customer export must include them. _SPEC §3 and adversarial A15
   (insider abuse) both assume "audit log + role separation cover v1" — that
   assumption is currently unowned, and A1 is the wrong home for it because the
   chain scope is the engagement. **Decided** (product owner, 2026-09-21, PR #2
   item D9), flagged here so it is not lost (A1 §2 lists it as a gap)._

   Ownership split applied for the Freeze: authentication, session, credential
   and role audit are **not** A1 kinds — they have no natural `engagement_id`
   (A1-5.1) and belong to A5's platform-scoped audit store. Two exceptions are
   engagement-scoped and are chained here: a global-blacklist mutation is
   composed as `scope_changed` into every affected engagement chain (A1-3.3,
   adversarial C-04), and engagement operator assignment is a known gap that A5
   MUST close by adding `engagement_assignment_changed` additively (A1-3.5) —
   see §6 item 17.
10. **A1-8.5 — SSE `id:` is the decimal `seq`.** Resume via `Last-Event-ID` is
    then a chain position: gap-free, ordered, and checkable by the consumer
    against the dense `seq` guarantee (A1-8.1). Alternative: use `event_id`,
    which is opaque and forces a lookup to resume. _`seq` is already public in
    every served envelope (A1-1.1), so exposing it in the frame discloses
    nothing new, and A0-1.6's "do not parse ids" rule is respected because
    `seq` is a declared ordering key, not an id. Framing itself is A4's.
    **Decided** (product owner, 2026-09-21, PR #2 §6.10)._
11. **A1-8.2 — a cursor must resolve inside the requested engagement.** Because
    `seq` is per engagement, engagement A's cursor is a *valid position* in
    engagement B; replaying it would silently serve B's rows from an unrelated
    point. Recommend the platform look the cursor's `id` up in the requested
    chain and answer `validation` (restart from page one) when it is not there.
    _A0-4.4 promises a forged cursor yields "at worst an empty page or
    `validation`"; this makes the cross-engagement case loud instead of
    confusing, at the cost of one indexed lookup per page. **Decided** (product
    owner, 2026-09-21, PR #2 §6.11; AM-4 asks A0-4.8 to bless the case); A2 may
    want the same rule for graph cursors._
12. **A1-7.6 — the dedup key is a client-supplied `idempotency_key`.** Required
    from machine principals, ≤ 64 chars, reused verbatim on retry, never reused
    across different content (that is `conflict`). Alternative considered and
    rejected: a content-derived key (hash of kind + payload) — it needs no
    client discipline but collapses two genuinely identical occurrences (same
    command, same target, same exit code, run twice) into one event, which is
    silent evidence loss. _The client key is the only construction that makes a
    retry safe without inventing occurrences; ADR-0013's offline buffering
    makes retries normal, not exceptional. **Decided** (product owner,
    2026-09-21, PR #2 §6.12)._
13. **A1-6.6 — export integrity metadata and the stamp wording.** The eight
    fields of A1-6.6 are the contract; the customer-visible sentence is the
    literal **"integrity verification failed"** (Q11) rendered with the override
    `reason` and the overriding user's id. _Q11 fixed the words, so this is a
    confirmation of the field set and of where the stamp appears (title block
    plus every page footer is the recommendation, so it survives excerpting).
    Report layout itself belongs to the report session. **Decided** (product
    owner, 2026-09-21, PR #2 §6.13)._
14. **A1-3.3 / A1-3.8 — the additive taxonomy changes this revision made.** Two
    kinds (`quarantine_recomputed` for A2-8.5, `graph_edge_retracted` for
    A2-3.9), three more added at the Freeze (adversarial T-01/T-02:
    `engagement_created`, `engagement_closed`, `artifact_released` — see A1-3.5)
    and three enum values (`chain_break_detected.break_kind:
    engagement_mismatch` for the A12 splicing case, `action_blocked.reason:
    append_rejected` so a refused append stays observable, and
    `notification_sent.notification_kind:chain_head_anchor` for the A1-5.8 (4)
    out-of-band head anchor — product owner decision 2026-09-21, PR #2 item D5),
    taking the closed list from 37 to **42** kinds. `chain_head_anchor` is a
    **notification** kind, not an A1 `Kind`, so the 42 constants are unchanged
    (A1-3.3, A1-4.2). One request from A2 was **corrected** rather than granted:
    no `node_superseded` kind, because `graph_node_written` +
    `graph_edge_written{supersedes}` already record it and a third encoding
    would be a third source of truth. _All additions are additive under A0-6.5
    and cost nothing before Freeze; after Freeze a new kind is still additive,
    so none of this is a one-way door. **Decided** (product owner, 2026-09-21,
    PR #2 §6.14)._

    **Decided (product owner, 2026-09-21, PR #2 item D6) — operator release of
    quarantine is removed.**
    `operator_release` is deleted from A1's `quarantine_kind` enum and A2
    provides no release operation: an `out_of_scope` node is released **only**
    by an operator scope change and the recomputation it causes (A2-8.5,
    ADR-0016 §2 — an out-of-scope node can never be a target of a planned
    action). `operator_quarantine` (tightening) is kept, and a `blacklisted`
    node is never releasable. If the product owner wants a manual release it
    MUST be a new ADR amending ADR-0016 §2 and MUST require the target to be
    inside the widened allowlist at release time.

15. **A1-6.5 — override authority and lifetime: DECIDED (product owner,
    2026-09-11) → ADR-0021.** *Authority:* admin-role only — an operator-scoped
    user, including one assigned to the engagement, MUST NOT override any break
    (SPEC §3 puts integrity-class controls next to the hard stop; Q11's
    "operator" reads as "human", so this narrows it deliberately). The insider
    argument (adversarial A15) was explicitly **not** the reason: the product
    owner does not treat insider abuse as a v1 concern. *Lifetime:* an `export`
    override is **not single-use** — it lives until the next `chain_verified` or
    `chain_break_detected` and may authorize several artifacts, because a
    per-artifact rule does not survive the long-term service vision. The
    residual risk is **accepted and transferred to the service owner** running
    the deployment, who compensates with logging; the platform's obligation is
    that the attribution is complete (one chained `artifact_released` per
    artifact, naming the override, the head, the artifact and the recipient,
    A1-6.6) and stamped into the export. ADR-0021 is **Accepted** (product owner,
    2026-09-21, PR #2 item D1).
16. **A1-5.8 / §6.7 — out-of-band head anchoring: DECIDED, approved (product
    owner, 2026-09-21, PR #2 item D5).** The Freeze ships the in-platform
    anchor — an append-only `chain_head_trail` table (`REVOKE UPDATE, DELETE`)
    plus `break_kind:"head_regression"` — **and** the out-of-band anchor: the
    platform also pushes `(engagement_id, head_seq, head_hash)` over the
    ADR-0012 §3 signed webhook, which is now normative as A1-5.8 (4) (product
    owner decision 2026-09-21, PR #2 item D5; item 7 records the same
    decision). The anchor was **approved, not declined**, so the truncation
    residual is narrowed rather than accepted outright: an attacker with both
    store-write and log-write access on the same host can still forge history
    (default Docker deployment, SPEC §10) and can suppress webhook delivery,
    but suppression is visible to the recipient as a head that stops advancing
    and a forged history contradicts the anchors the recipient already holds.
    A1-6.6 accordingly prints the **narrowed** residual, and detection depends
    on the recipient retaining and comparing its anchors against that block's
    `head_seq`/`head_hash`.
17. **A1 §6.9 — user/session audit ownership — Decided (product owner,
    2026-09-21, PR #2 item D9).** A1 chains engagement-scoped facts only;
    authentication, session and role audit belong to A5 in a platform-scoped
    store. **Known debt accepted at Freeze:** engagement operator assignment
    (who may approve, ADR-0012 §1) is unaudited in v1 until A5 adds
    `engagement_assignment_changed` additively (A1-3.5); an insider admin
    self-assigning and then approving is detectable only in A5's store
    (adversarial C-08).

### A0 amendment requests

A1 is written against current A0; these are requests, not assumptions.

| # | A0 clause | Request | Why A1 needs it |
|---|---|---|---|
| AM-1 | A0-1.2 | Register a prefix for a human user principal (`usr_` recommended), or state that A5 owns it | A1-2.2 cannot validate `actor.principal_id` for `type:"user"` without one (A0-1.5); blocks every user-composed kind, incl. the A1-6.5 override attribution. Identical to A2 AM-1 — one decision serves both |
| AM-2 | A0-7.7 | Extend the mechanism-assignment obligation from "A2 and A3" to **every** contract that defines capped fields, and adopt the A1-local constants of §4.1 (`ProseLongMaxBytes`, `ProseMediumMaxBytes`, `TargetMaxBytes`, `LabelMaxBytes`, `VersionMaxBytes`, `KindNameMaxBytes`, `DigestMaxBytes`, `EvidenceRefsMax`, `EventRefsMax`, `EventMaxCanonicalBytes`) into the single A0-7.1 table/const block | A1-4.5 discharges the obligation but A0-7.7 does not name A1; two const blocks drift (same argument as A2 AM-2) |
| AM-3 | A0-5.4 | Record that A1 clamps `recorded_at` forward to `max(clock_now, prev_recorded_at)` inside the append lock | A1-5.4 makes `recorded_at` non-decreasing along `seq` so "the log shows X before Y" is safe to rely on; that is a refinement of A0-5.4's "one injected clock" rule and should be visible there, not only in A1 |
| AM-4 | A0-4.8 | Add "the cursor's `id` does not resolve in the requested collection" to the list of `validation` cases | A1-8.2 needs it: `seq` is per engagement, so a foreign cursor is otherwise a *valid* position in the wrong chain. A2-11.4's `TestCursorFromEngagementARejectedInB` becomes decidable for both contracts |

### Cross-contract requests (not A0)

- **A2** — answered in full by A1-3.8 (envelope field names confirmed; two kinds
  added; one request corrected; the dedup key A2-4.7 needs is A1-7.6). A2 owes
  A1 nothing further; A1 stores `node_kind`/`edge_kind` as capped strings
  (A1-3.6) so A2 can add kinds without an A1 change. A2 MAY want A1-8.2's
  cursor rule for graph cursors.
- **A3** — the consumer guarantee table is A1-8.9; A3 reads the log only through
  it. A3 owns the stage-view side of the report-inclusion and quarantine flags
  (A2-12.4/12.5); A1 supplies `report_inclusion_changed`,
  `graph_node_quarantined` and `quarantine_recomputed` as the audit trail behind
  them.
- **A4** — endpoint paths/methods/wrappers, the per-endpoint request-body bound
  (A0-8.9, A1-7.10 step 2), SSE framing plus the resync/restart mechanism
  (A1-8.5), the rate limit on `on_demand` verification (A1-6.8), the wire shape
  of the integrity flag on internal views (A1-6.4), the filter parameters of
  A1-8.3 and the read-direction parameter of A1-8.1, the optional batch-append
  endpoint (A1-7.8), and the route-table audit for A1-8.6's
  `TestNoBulkEventReadSpansEngagements`.
- **A5** — scope names and the machine-principal exclusion list derived from
  A1-7.4 and A1-8.4 (worker/node: append the three C kinds, no read;
  orchestrator: neither, per §6.8), the admin role gate for `integrity_override`
  (A1-6.5) and `on_demand` verification (A1-6.8), the operator/viewer
  engagement-assignment rule behind A1-8.4, the `usr_` principal shape (AM-1),
  and ownership of the user/session audit gap (§6.9).
- **A7** — A1 stores `fingerprint_hash`, `risk_tier`, `tool_version` and the
  spawn fields as opaque capped values (A1-3.6); A7 owns their content and
  vocabulary. A7 needs from A1: `spawn_requested.spawn_request_event_id` and
  `approval_requested.request_event_id` as the correlation seam (A1-4.11), and
  the A0-2 canonical form for fingerprints (A0-2.1, A1-5.2's precedent for a
  declared exclusion list).
- **Report session** — A1-6.6 fixes the integrity metadata fields and the stamp
  wording; the report builder owns layout, the verification-bundle document
  format (A1-6.7) and how a dangling evidence reference is rendered (A1-8.8).
- **Backlog session 6 (persistence)** — DDL, indexes, partitioning, retention
  and archival, the locking primitive behind A1-5.4, the dedup table behind
  A1-7.6, the ingest watermark behind A1-7.7, and any checkpoint/re-genesis
  mechanism (A1-6.9, which needs its own ADR). Travelling recommendation:
  enforce append-only in the database too (`REVOKE UPDATE, DELETE`), A1-7.9.
