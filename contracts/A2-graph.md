# A2 — Engagement context graph

## 1. Header

| | |
|---|---|
| **Contract id** | A2 |
| **Status** | `Draft` (`contracts/README.md` lifecycle: Draft → Frozen → Implemented) |
| **Owner** | architect |
| **Gates** | A3 (stage views) · A4 (`/api/v1` graph endpoints) · A5 (token scopes: graph writes are excluded for machine principals) · `internal/graph`, `internal/policy`, the ingest path, the report builder |
| **Implements** | ADR-0016 §1–§5 · ADR-0005 §2/§3/§5 · ADR-0009 §1–§2 · ADR-0019 §2/§5 · ADR-0020 §3–§4 · SPEC §6, §8, C8 · Q1, Q2, Q3, Q4, Q5, Q6 (+ Q6 worker addendum) · adversarial A1, A12 |
| **Depends on** | A0 (frozen for this document: ids A0-1.2, canonical JSON A0-2, error kinds A0-3, paging A0-4, time A0-5, unknown fields A0-6, size caps A0-7, field conventions A0-8) · A1 (event envelope fields consumed by provenance, A2-5.4) |
| **Authority** | ADR > SPEC > DESIGN > contract. A clause here that contradicts an Accepted ADR is a defect in this document. A0 owns every cross-cutting convention; A2 cites A0 clause ids instead of restating them, and requests amendments only in §6. |

## 2. Scope

**Fixed here:** the closed node-kind and edge-kind lists with per-kind
required/optional fields, allowed value types and size caps; the revision
rule (`supersedes`) and how a client resolves the current revision; mandatory
platform-stamped provenance; the bounded flat `attrs` escape hatch; the
quarantine state and report exclusion; the no-secret-values rule; write
validation semantics and error kinds; engagement binding (C8) and its
negative tests; the exact shapes A3 may assume.

**Not fixed here:** stage-view composition, the per-stage view definitions and
the capped 1-hop drill-down (A3; Q1) · the A0-7.10 composition rule for
500 nodes vs 64 KiB (A3, escalated by A0 §6.14) · HTTP endpoints, scopes and
SSE framing (A4) · which scopes machine principals are denied (A5; Q6) ·
event kinds, chain fields and the A1 dedup key (A1) · spawn/fingerprint
schemas (A7) · DDL, indexes and migrations (backlog 6) · scope/blacklist
*content* and matching semantics (`internal/policy`, ADR-0005; A2 consumes its
verdict) · free-form graph query (Q1: none in v1; A6 stub) · UI rendering.

## 3. Normative clauses

### A2-1 · Graph model

- **A2-1.1** The graph is a per-engagement property graph (ADR-0016 §1, §5):
  nodes and edges stored in PostgreSQL behind the store seam (ADR-0010). No
  external graph database, no query language, no traversal API (Q1).
- **A2-1.2** Node ids use prefix `gn_`, edge ids `ge_`, evidence ids `evi_`
  (A0-1.2). Ids are platform-generated (A0-1.4); a client MUST NOT supply,
  construct or reuse an id from another engagement (A0-1.6, A2-11).
- **A2-1.3** Node and edge **content is immutable** (ADR-0016 §4, Q2). The only
  mutable fields in A2 are `quarantined`, `quarantine_reason`,
  `report_excluded` (A2-8) and `retracted` on an edge (A2-3.9). An attempt to
  change any other field of an existing node or edge MUST return `conflict`
  (A0-3.1: immutable field) and MUST NOT partially apply.
- **A2-1.4** Every node and edge carries `seq`: a per-engagement,
  platform-assigned, strictly increasing integer, immutable and unique. `seq`
  is the ordering key for every paginated graph collection (A0-4.3) and the
  `k` of a cursor (A0-4.4). Clients MUST NOT derive order from id text
  (A0-1.6).
- **A2-1.5** Every node and edge carries `engagement_id`, stamped by the
  platform from the authenticated request binding (Q6, third enforcement
  layer). It MUST NOT be accepted from a request body: the write types have no
  such field, so a body that carries one is an unknown field → `validation`
  (A0-6.2).
- **A2-1.6** Field naming inside graph payloads: a bare `node_id` MUST NOT
  appear — A0-3.6 reserves `node_id` for the **remote agent node**
  (`slp_node_`). Graph-node references are `source_id`, `target_id`,
  `supersedes_id`, `superseded_by_id`; the agent node is `agent_node_id`; and
  error/log attributes use `graph_node_id` for a graph node (A0-3.6). §6.3 asks
  A0 to bless the payload spellings.
- **A2-1.7** Value types are closed: UTF-8 strings (A0-2.3, capped per A2-7),
  integers within A0-2.6, booleans (A0-8.8), closed enum strings (A0-8.5),
  ids (A0-1), timestamps (A0-5.1) and bounded lists of those. Floats, `null`
  (A0-8.3) and nested objects MUST NOT appear anywhere in a node or edge
  except the flat `attrs` map (A2-6).
- **A2-1.8** A node is *current* iff `superseded_by_id` is absent (A2-4.4).
  "current" is a defined term, not a stored field — one source of truth.
- **A2-1.9** Declared limitation: A2 has no merge operation. Two nodes that
  turn out to describe the same real-world object (a host seen first by IP,
  then by hostname) stay separate; the remedy is a revision (A2-4) or a
  `contradicts` edge (A2-3.7). Adding a merge kind requires an ADR.

### A2-2 · Node kinds (closed)

- **A2-2.1** The node-kind list is closed (ADR-0016 §1 + Q2). Spelling per
  A0-8.5; an unknown kind on write → `validation` (A0-6.3, Q3). On read a
  client MUST preserve an unknown kind verbatim and skip the item (A0-6.3).

  | # | kind | ADR-0016 §1 name | One line |
  |---|---|---|---|
  | 1 | `host` | hosts | a machine, interface or addressable endpoint |
  | 2 | `network` | networks | an address range / segment |
  | 3 | `service` | services | a listener on a host |
  | 4 | `identity` | accounts/identities | a user, computer or service account |
  | 5 | `group` | groups | a group (AD group, role) |
  | 6 | `credential` | credentials | a **reference** to captured secret material (A2-9) |
  | 7 | `share` | shares | a file/ADMIN share or equivalent resource |
  | 8 | `evidence_ref` | artifacts/evidence refs | an artifact that needs its own edges |
  | 9 | `finding` | findings/hypotheses | confirmed or asserted impact (Q2) |
  | 10 | `hypothesis` | findings/hypotheses | an unconfirmed claim under test (Q2) |

  `finding` and `hypothesis` are first-class kinds, not flags on another kind
  (Q2): they carry different required fields and different status enums.
- **A2-2.2** Fields common to every kind. `label` is required on all kinds;
  `summary` is optional on all kinds except `finding` (A2-2.4).

  | Field | Type | Req | Cap / rule |
  |---|---|---|---|
  | `label` | string | yes | ≤ 128 B (A2-7); the shortest unambiguous name a human recognises in a report |
  | `summary` | string | `finding`: yes | ≤ 512 B, `finding` ≤ 2048 B (A0-7.1) |
  | `evidence_ids` | []string (`evi_`) | no | ≤ 8 entries, each a valid `evi_` (A0-1.5) |
  | `attrs` | flat map | no | A2-6 |
  | `quarantined`, `quarantine_reason`, `report_excluded` | bool / enum / bool | platform-set | A2-8; `quarantined` and `report_excluded` always present (`false` is a value, A0-8.3/8.8) |
  | `supersedes_id`, `superseded_by_id` | string (`gn_`) | no | A2-4 |
  | `content_hash` | 64-char lowercase hex | platform-set | A2-4.6, A0-2.15/8.7 |
  | `provenance` | object | platform-set, mandatory | A2-5 |
  | `id`, `engagement_id`, `seq`, `kind` | per A2-1 | platform-set | A2-1.2/1.4/1.5 |

- **A2-2.3** Kind-specific fields. A field that does not apply to the declared
  kind MUST NOT be present; a write that sets one is rejected → `validation`
  (Q3 hard reject, A2-10.4). "Type" columns are the closed value types of
  A2-1.7.

  | kind | required | optional |
  |---|---|---|
  | `host` | `label` | `addresses` []string ≤ 16, each ≤ 64 B |
  | `network` | `label`, `cidr` | `addresses` (as `host`) |
  | `service` | `label`, `port` int 1–65535, `transport` enum `tcp`\|`udp` | `protocol` string ≤ 32 |
  | `identity` | `label` | `sid` ≤ 64, `domain` ≤ 253 |
  | `group` | `label` | `sid` ≤ 64, `domain` ≤ 253 |
  | `credential` | `label`, `credential_kind`, `evidence_id` (`evi_`) | `domain` ≤ 253 |
  | `share` | `label` | `domain` ≤ 253 |
  | `evidence_ref` | `label`, `evidence_id` (`evi_`), `media_kind` | `size_bytes` int ≥ 0 (A0-8.2) |
  | `finding` | `label`, `summary`, `severity`, `status` | `evidence_ids`, `attrs` |
  | `hypothesis` | `label`, `claim`, `basis`, `status` | `evidence_ids`, `attrs` |

  `cidr` MUST parse with `net.ParseCIDR` (stdlib) — an unparseable value is
  `validation`, never a stored string (A2-10.4). Prefer the common
  `evidence_ids` field; create an `evidence_ref` node only when the artifact
  itself needs edges or independent provenance (DESIGN §2 simplicity).
- **A2-2.4** `credential_kind` (closed, A0-8.5): `password`, `hash`, `ticket`,
  `key`, `token`, `certificate`, `other`. It names the *class* of the captured
  material, never the material (A2-9). `media_kind` (closed): `file`,
  `screenshot`, `capture`, `dump`, `log`, `other`.
- **A2-2.5** `severity` (closed, `finding` only): `info`, `low`, `medium`,
  `high`, `critical`. No CVSS field in v1; a numeric score MAY be carried in
  `attrs` as a fixed-point integer `cvss_v3_x10` (scale ×10 fixed **here**, as
  A0-2.6 requires of the owning contract). **PO confirm** (§6.8).
- **A2-2.6** `status` is per kind and closed:
  - `finding`: `open`, `confirmed`, `refuted`, `remediated`.
  - `hypothesis`: `open`, `supported`, `refuted`.
  There is deliberately **no** `superseded` status: supersession is expressed
  by the `supersedes` edge and `superseded_by_id` (A2-4), and a second encoding
  would be a second source of truth. **PO confirm** (§6.7).
- **A2-2.7** `confidence` is **not** a finding field: it is the provenance
  confidence of A2-5.6 and appears exactly once per node and edge. Q2's
  "Finding carries confidence" is discharged by the mandatory provenance block
  — a second confidence field would let a worker assert confidence in its own
  claim without an evidence grade. **PO confirm** (§6.1).
- **A2-2.8** Who may set what:

  | Field class | Set by | Changed by |
  |---|---|---|
  | `label`, `summary`, kind fields, `attrs`, `evidence_ids` | platform ingest, from an A1 event (Q6) | never — revision only (A2-4) |
  | `severity`, `status`, `claim`, `basis` | platform ingest; the *value* originates in a worker task result or an orchestrator report, the *write* is platform-side (Q6) | never — revision only; an operator correction is a new node with `principal_kind: operator` |
  | `quarantined`, `quarantine_reason` | platform policy engine (A2-8.2) | only as a consequence of an operator scope/blacklist change (Q6), event-logged |
  | `report_excluded` | operator only (A2-8.7) | operator only, event-logged |
  | `provenance`, `seq`, `content_hash`, ids | platform only | never |

  A machine principal MUST NOT hold a graph-write scope at all (Q6 worker
  addendum, A5); "set by platform ingest" is the only agent-derived path.

### A2-3 · Edge kinds (closed)

- **A2-3.1** The edge-kind list is closed (ADR-0016 §1). An unknown edge kind
  on write → `validation` (A0-6.3). Edges are directed; `source_id` and
  `target_id` are `gn_` ids in the same engagement (A2-11.3).
- **A2-3.2** Endpoint matrix. A write whose endpoint node kind is not in the
  allowed set → `validation` naming the kind and the allowed set (A2-10.5); an
  endpoint id that does not resolve in this engagement → `notfound` (A0-3.9).

  | kind | source kinds | target kinds | Purpose |
  |---|---|---|---|
  | `reachable` | `host`, `network` | `host`, `network`, `service` | network reachability observed or inferred — the attack-path backbone |
  | `authenticates_to` | `credential`, `identity` | `host`, `service`, `share` | this credential/identity authenticates to that resource |
  | `member_of` | `identity`, `group` | `group` | group membership (nested groups included, one edge per hop) |
  | `grants_access` | `group`, `identity` | `share`, `host`, `service` | membership/ownership confers access to a resource |
  | `exploited_by` | `host`, `network`, `service`, `identity`, `group`, `credential`, `share` | `finding`, `hypothesis` | the finding/hypothesis that describes how this node was or could be exploited (§6.6) |
  | `contradicts` | `finding`, `hypothesis` | `finding`, `hypothesis` | self-correction signal (ADR-0016 §4): two claims cannot both hold |
  | `supersedes` | any kind *K* | same kind *K* | revision; content is never overwritten (A2-4) |

- **A2-3.3** Cardinality: for every kind except `supersedes`, at most one edge
  per `(engagement_id, kind, source_id, target_id)`. `supersedes` is 1:1 at
  both ends (A2-4.3).
- **A2-3.4** Duplicate policy — **collapse, do not error**: a write whose
  `(engagement_id, kind, source_id, target_id)` already exists MUST return the
  existing `ge_` id, store nothing new, and report success. _A re-observation
  carries no new graph information; the observation itself is already in the
  append-only event log (A1, ADR-0009 §1). This also makes a retried edge
  write idempotent, which A0-3.11 requires before a retry is safe._
- **A2-3.5** The edge dedup key is `(engagement_id, kind, source_id,
  target_id)`; the node dedup key is `(engagement_id, kind, content_hash)`
  (A2-4.7). Both MUST be enforced by a uniqueness constraint at the store seam
  so a race cannot create a duplicate (A0-1.4: a uniqueness violation surfaces
  as `internal`, never a silent retry loop).
- **A2-3.6** `contradicts` is symmetric in meaning but stored once: if the
  inverse pair already exists, the write MUST collapse to the existing edge and
  MUST NOT create the inverse. Which of the two directions is stored is
  decided by the platform at first write and is immutable afterwards.
- **A2-3.7** An edge MUST NOT have `source_id == target_id` → `validation`.
  `contradicts` between a node and its own revision is meaningless: use
  `supersedes`.
- **A2-3.8** `source_kind` and `target_kind` are denormalized onto the edge by
  the platform at ingest, immutable, and MUST equal the referenced nodes'
  kinds. _A3 and the report builder need them without a join; the contract test
  asserts equality, so the duplication cannot drift silently._
- **A2-3.9** `retracted` (bool, default `false`) is the only mutable edge
  field: a platform-side correction, operator- or ingest-initiated, always
  accompanied by an A1 event carrying the reason. A retracted edge, and any
  edge with a quarantined endpoint (A2-8.4), MUST NOT be returned by
  planning-facing reads (A2-12.4). Content is not overwritten — `kind` and
  endpoints stay. **PO confirm** (§6.5: alternative is to leave edges
  uncorrectable and revise endpoint nodes instead).

### A2-4 · Revision, supersession and the content fingerprint

- **A2-4.1** Content is never overwritten (ADR-0016 §4, Q2). A correction is a
  **new node** plus a `supersedes` edge from the new node to the old one; the
  old node, its edges and its provenance are preserved byte-for-byte.
- **A2-4.2** Direction: `new --supersedes--> old`. The new node stores
  `supersedes_id = <old gn_>`; the platform sets `superseded_by_id = <new gn_>`
  on the old node. That single field is the only mutation supersession causes,
  and it is a derived pointer, not content — A2-1.3's immutability list does
  not include it because it is platform-maintained from the edge set.
- **A2-4.3** Constraints, each hard-rejected: same `kind` on both ends
  (`validation`) · same engagement (A2-11.3) · the target MUST be current
  (`conflict` if it already has `superseded_by_id`) · no cycles: the target
  MUST NOT be an ancestor of the source (`conflict`) · a node MUST NOT
  supersede itself (A2-3.7). _Linear chains only: a fork would make "the
  current revision" ambiguous, and ambiguity in evidence is worse than a
  rejected write (Q3)._
- **A2-4.4** Finding the current revision, without a query endpoint (Q1):
  every node read returns `supersedes_id` and `superseded_by_id`. A client
  follows `superseded_by_id` until it is absent — that node is current
  (A2-1.8) — and follows `supersedes_id` for history. Chain length is bounded
  by A2-4.5.
- **A2-4.5** List reads MUST return current nodes only unless the caller
  explicitly asks for history (A4 defines the parameter); a history read
  returns the whole chain ordered by `seq` (A2-1.4) and paginates per A0-4. A
  single chain walk is bounded by `MaxSupersedeChain` (64) revisions; past
  that the read paginates rather than growing (adversarial A14: an agent could
  otherwise build a chain whose walk is unbounded). A superseded node MUST NOT
  be silently dropped from a read that asked for history — it is evidence.
- **A2-4.6** `content_hash` is SHA-256 (A0-2.15) over the **canonical JSON**
  (A0-2) of a purpose-built content document with a **fixed key set**
  (A0-2.14): `addresses`, `attrs`, `basis`, `cidr`, `claim`,
  `credential_kind`, `domain`, `evidence_id`, `evidence_ids`, `kind`, `label`,
  `media_kind`, `port`, `protocol`, `severity`, `sid`, `size_bytes`, `status`,
  `summary`, `transport` — every field on every instance, zero-valued when not
  applicable to the kind (A0-2.14), keys in UTF-8 byte order (A0-2.4),
  integers only (A0-2.6). Its A0-2.12 **exclusion list is empty**: ids, `seq`,
  provenance, quarantine flags, `report_excluded`, `superseded_by_id` and
  `content_hash` are not part of the document at all, so none of them can
  influence the digest. _Consequence of A0-2.14: these content fields are the
  one place in A2 where "unset" serializes as a zero value rather than as
  absence (A0-8.3); the API representation of a node keeps A0-8.3 absence
  semantics and the digest is computed from the dedicated document type._
- **A2-4.7** Node dedup: a write whose `(engagement_id, kind, content_hash)`
  already exists MUST return the existing `gn_` id and store nothing (A2-3.4
  rationale). Replayed ingest of the same A1 event is therefore idempotent
  (A0-3.11, ADR-0013 offline buffering).
- **A2-4.8** A0-2.16 applies: the exact canonical bytes `content_hash` was
  computed over MUST be persisted with the node, and any verification MUST
  recompute from those stored bytes — never from a re-serialization of decoded
  fields. _A later added field or a different encoder setting would otherwise
  silently invalidate every stored fingerprint._
- **A2-4.9** `content_hash` is an integrity and dedup value, not a secret and
  not a substitute for the A1 event hash chain (Q11): it is not chained, and a
  mismatch is a platform defect (`internal`), never a customer-facing
  `integrity_failed` (A0-3.1 reserves that for the chain).

### A2-5 · Provenance (mandatory)

- **A2-5.1** Every node and every edge MUST carry a `provenance` object
  (ADR-0016 §1: "graph content is evidence, not opinion"). A record without
  provenance MUST NOT be stored.
- **A2-5.2** Provenance is **platform-stamped at ingest** and MUST NOT be
  client-supplied (Q6 worker addendum). The write request types contain no
  provenance field, so a body that carries one is rejected as an unknown field
  → `validation` (A0-6.2). _A worker that could stamp its own provenance could
  attribute a fabricated finding to a tool run that never happened._
- **A2-5.3** Fields. Absent optional fields are omitted, never `null`
  (A0-8.3); every id is validated per A0-1.5.

  | Field | Type | Req | Meaning |
  |---|---|---|---|
  | `principal_kind` | enum | yes | `platform`, `orchestrator`, `worker`, `operator` (A0-8.5) |
  | `run_id` | `run_` | yes | the run whose work produced this record |
  | `job_id` | `job_` | when `principal_kind` ∈ {`orchestrator`, `worker`} | orchestrator container |
  | `task_id` | `task_` | when `principal_kind` = `worker` | worker container |
  | `agent_node_id` | `slp_node_` | when observed via a remote agent (Q9) | the remote agent node — **not** `node_id` (A2-1.6, A0-3.6) |
  | `operator_id` | id of a prefix registered in A0-1.2 | when `principal_kind` = `operator` | human actor; **blocked** until A0/A5 register the prefix (§6.2) |
  | `tool_id` | `tool_` | when a registry tool produced the observation | ADR-0008, Q14, A0-1.3 |
  | `tool_version` | string ≤ 32 | with `tool_id` | the registry version string; MUST NOT be folded into `tool_id` (A0-1.3) |
  | `event_id` | `evt_` | yes | the originating A1 event (A2-5.4) |
  | `recorded_at` | timestamp | yes | platform ingest time, A0-5.1/5.4 — the node's only creation timestamp |
  | `observed_claimed_at` | timestamp | no | untrusted client-supplied observation time carried by the A1 event; suffix per A0-5.7/A0-8.2 |
  | `confidence` | enum | yes | A2-5.6 |

- **A2-5.4** Tie into A1: `event_id` MUST reference an event that exists **in
  this engagement** (A2-11). A2 consumes the A1 envelope fields `event_id`,
  `kind`, `recorded_at`, `seq`, `engagement_id`, `run_id`, `job_id` (envelope
  shape per A0-2.17 vector V5 and A0-3.6; A1 was not on disk when A2 was
  drafted — §6.9). Every graph record is thereby anchored to a hash-chained,
  append-only audit row (Q11, ADR-0009 §1): deleting or editing graph content
  cannot remove the evidence of its creation.
- **A2-5.5** `observed_claimed_at` MUST NOT drive ordering, supersession,
  quarantine, expiry or any digest (A0-5.7); ordering uses `seq` (A2-1.4) and
  `recorded_at`. It is stored next to `recorded_at` so divergence is visible in
  the audit trail.
- **A2-5.6** `confidence` (closed, A0-8.5) grades the **evidence**, not the
  author's optimism: `observed` (directly present in captured tool output
  referenced by `event_id`/`evidence_ids`) · `inferred` (derived by reasoning
  from other graph content) · `verified` (reproduced by a second, independent
  observation). `confidence` is set by the platform at ingest from the ingest
  rule, MUST NOT be raised by a later write, and can only change through a
  revision (A2-4). **PO confirm** (§6.1: alternative scale `low`/`medium`/
  `high`).
- **A2-5.7** A write whose provenance cannot be established MUST NOT be
  stored, MUST NOT be stored with placeholder or synthesized values, and
  returns: `internal` (500) when the platform ingest path failed to produce an
  `event_id` or a `run_id` — that is a platform defect, since graph writes are
  platform-side (Q6) · `notfound` (404) when a supplied `event_id` is
  well-formed but does not resolve in this engagement (A0-3.9: also the
  cross-engagement case, no existence disclosure) · `validation` (400) when a
  supplied id is malformed (A0-1.5) or a required provenance field is missing
  from a caller-authorized path (operator correction). There is no
  "best-effort provenance" mode.
- **A2-5.8** Provenance MUST NOT contain secret material (A2-9, A0-3.7) and
  MUST NOT be echoed into an error `message` beyond ids and field names
  (A0-3.4/3.5).

### A2-6 · `attrs` escape hatch (bounded)

- **A2-6.1** `attrs` exists so a kind schema does not have to change for every
  new observation (Q3). It is **flat**: `map[string]AttrValue` with a depth of
  exactly one. Nested objects, arrays, `null` and floats MUST be rejected →
  `validation` (A2-1.7, A0-2.6).
- **A2-6.2** Keys MUST match A0-8.1 (`^[a-z][a-z0-9_]{0,39}$`), which is also
  the key length cap: ≤ 40 chars, no new A2 constant. Values are `string`
  ≤ 512 B, `int64` within A0-2.6, or `bool`.
- **A2-6.3** A key MUST NOT equal or shadow a schema field name of the node's
  kind or of A2-2.2 (`label`, `summary`, `kind`, `port`, `severity`,
  `evidence_ids`, …) → `validation`. _Two places to look for one fact is how a
  report ends up contradicting the graph._
- **A2-6.4** Caps (mechanism R, A2-7): ≤ 16 keys (`AttrsMaxKeys`), ≤ 512 B per
  string value (`AttrValueMaxBytes`), ≤ 4096 B for the serialized `attrs`
  object (`AttrsTotalMaxBytes`). Exceeding any of them → `summary_too_large`
  (A0-7.6, 413) naming the key, the cap and the actual byte count.
- **A2-6.5** `attrs` is part of `content_hash` (A2-4.6), so an attrs-only
  change is a revision (A2-4), not an update.
- **A2-6.6** Promotion review: an `attrs` key that recurs is a candidate for
  the kind schema. The platform MUST expose, per engagement and per node kind,
  the distinct `attrs` keys with their occurrence counts, ordered by key
  byte-wise (A0-1.9). A key observed on **≥ 3** nodes of one kind in one
  engagement SHOULD be proposed as a schema field; promotion is an additive
  contract amendment (A0-6.5) and a new ADR only if it changes a cap or an
  enum. _The escape hatch must not become the schema (Q3)._
- **A2-6.7** `attrs` is untrusted content (SPEC §6, adversarial A1): it MUST be
  rendered through `html/template` auto-escaping in the UI, MUST be flagged as
  untrusted in approval views (ADR-0018 §4), and MUST NEVER be interpreted as
  configuration, a path, a command, a scope entry or an id.

### A2-7 · Size caps and enforcement mechanisms

- **A2-7.1** A0-7.7 obligation: every capped field class A2 defines, with its
  cap and its mechanism. **A2 assigns mechanism R (reject, A0-7.6) to every
  class and mechanism T (truncate, A0-7.5) to none**: all A2 caps guard
  client- or agent-originated content that becomes evidence, and silently
  shortened evidence misleads the approver and corrupts the report (ADR-0016
  §1, ADR-0018 §1). Mechanism T belongs to A3's platform-computed view fields.

  | Field class | Cap | Constant | Source | Mech |
  |---|---|---|---|---|
  | `summary` (all kinds except `finding`) | 512 B | `NodeSummaryMaxBytes` | Q4 / A0-7.1 | **R** |
  | `summary` (`finding`) | 2048 B | `FindingSummaryMaxBytes` | Q4 / A0-7.1 | **R** |
  | `label` | 128 B | `NodeLabelMaxBytes` (A2-local) | A2 | **R** |
  | `claim` (`hypothesis`) | 512 B | `HypothesisClaimMaxBytes` (A2-local) | A2 | **R** |
  | `basis` (`hypothesis`) | 1024 B | `HypothesisBasisMaxBytes` (A2-local) | A2 | **R** |
  | `attrs` string value | 512 B | `AttrValueMaxBytes` (A2-local) | A2 | **R** |
  | `attrs` serialized object | 4096 B | `AttrsTotalMaxBytes` (A2-local) | A2 | **R** |
  | `attrs` key count | 16 | `AttrsMaxKeys` (A2-local) | A2 | **R** |
  | `evidence_ids` count | 8 | `EvidenceIDsMax` (A2-local) | A2 | **R** |
  | `addresses` count / entry | 16 / 64 B | `AddressesMax`, `AddressMaxBytes` (A2-local) | A2 | **R** |
  | `protocol`, `sid`, `domain`, `credential_kind`-adjacent strings | 32 / 64 / 253 B | A2-local per A2-2.3 | A2 | **R** |
  | stage view document, node count, stage summary | 64 KiB / 500 / 2 KiB | `StageView*`, `StageSummaryMaxBytes` | Q4 / A0-7.1 | **A3** (T expected) |

  Measurement per A0-7.3 (decoded UTF-8 bytes of the value; serialized bytes
  for a document), truncation never applied by A2 (A0-7.4 is A3's concern).
  A2-local constants are new field classes under A0-7.7's delegation, not
  changes to the Q4 constants (A0-7.2); §6.4 asks A0 to adopt them into the
  A0-7.1 table so all caps live in one const block. **PO confirm**.
- **A2-7.2** Enforcement is platform-side at ingest (A0-7.8, Q3, ADR-0005 §5).
  A client-side pre-check MAY exist and is not enforcement. Caps MUST be
  checked before the record is stored, in one pass, in the order of A2-10.2.
- **A2-7.3** The caps are contract constants: changing an A2-local cap requires
  a contract amendment plus a migration note for stored rows; changing a Q4 cap
  requires a new ADR (A0-7.2).

### A2-8 · Quarantine and report exclusion

- **A2-8.1** The quarantine state of a node is the pair `quarantined` (bool,
  adjective — no `is_` prefix, A0-8.8) and `quarantine_reason` (closed enum:
  `out_of_scope`, `blacklisted`; absent when `quarantined` is `false`,
  A0-8.3). Both are serialized on every node read.
- **A2-8.2** Quarantine is **derived, never asserted**: the platform core
  policy engine (ADR-0005 §5, SPEC §6) evaluates the node's target identity
  (`label`, `addresses`, `cidr`, `domain`, `sid`) against the engagement
  allowlist and the global/per-engagement blacklist at ingest. A caller MUST
  NOT set `quarantined` or `quarantine_reason` — the write types have no such
  fields (A0-6.2 → `validation`). Blacklist beats allowlist beats approval
  (ADR-0005 §3, SPEC §6).
- **A2-8.3** A quarantined node is **recorded, never actionable** (ADR-0016 §2,
  SPEC §6): it MUST NOT be returned as a candidate target by any
  planning-facing read (A2-12.4), MUST NOT be accepted as the target of a
  spawn or action even with a valid approval (A0-3.1 `approval_required` is not
  reachable for it — the policy check runs first), and MUST NOT be used to
  derive scope. Recording a discovery is itself sensitive (ADR-0016
  Consequences), so A2-9's minimization rules apply to quarantined nodes too.
- **A2-8.4** Quarantine propagates to edges by derivation, not by flag: an edge
  with a quarantined endpoint MUST NOT be returned by planning-facing reads
  (A2-3.9).
- **A2-8.5** `quarantine_reason: blacklisted` is **not releasable**: it MUST
  survive any scope change and any approval, because the blacklist cannot be
  overridden by scopes or approvals (ADR-0005 §3). `out_of_scope` changes only
  when an operator changes the engagement scope (Q6: machine principals never
  mutate scope/blacklist); the platform MUST recompute quarantine on such a
  change and MUST record the recomputation as an A1 event. **PO confirm**
  (§6.3: whether a blacklisted discovery is recorded at all, or refused).
- **A2-8.6** Reporting handover (Q5): quarantined nodes MUST be included in the
  reporting handover and marked *not tested*. The machine-readable marker is
  `quarantined: true` plus `quarantine_reason`; the words "not tested" are
  report prose (A0-3.4: prose is never parsed). A report MUST NOT render a
  quarantined node as tested, attempted or actionable, and MUST NOT render it
  inside an attack path.
- **A2-8.7** An operator MAY exclude a quarantined node from the report by
  setting `report_excluded: true`, which MUST be accompanied by an A1 event
  naming the operator, the `gn_` id and the reason text. Exclusion-by-flag over
  deletion: the graph stays intact, the exclusion is auditable and reversible,
  and no `supersedes` chain or evidence reference breaks (ADR-0016 §4). A
  delete endpoint for graph content MUST NOT exist in v1. **PO confirm** (§6.4).
- **A2-8.8** `report_excluded` MUST NOT be settable by a machine principal
  (Q6) and MUST NOT alter planning behaviour — it is a reporting filter only.
- **A2-8.9** Why mutability does not break ordering (A0-4.3): `quarantined`,
  `quarantine_reason`, `report_excluded` and `retracted` MUST NOT participate
  in any collection order, cursor or `seq` (A2-1.4). Every A2 collection orders
  by `(seq, id)` — both immutable — so an operator flipping a flag mid-paging
  can change *which* rows match a filter but can never skip or duplicate a row
  in an in-flight page walk. Clients MUST still deduplicate by id and MUST NOT
  assume a paged set is a snapshot (A0-4.7: mutable rows, no isolation knob).
- **A2-8.10** A quarantine hit is a security-relevant observation: the platform
  MUST emit an A1 event when a node is quarantined at ingest (`blacklisted` in
  particular), so "the agent saw a forbidden target and did not touch it" is
  provable from the audit trail alone.

### A2-9 · No secret values in the graph

- **A2-9.1** No graph field — `label`, `summary`, `claim`, `basis`, `attrs`
  value, `protocol`, `domain`, provenance field, edge field — MAY contain
  secret material: plaintext credentials, captured password/NTLM/Kerberos
  hashes, private keys, tickets, tokens, session cookies, TOTP seeds, or the
  per-run masking-mapping values of ADR-0020 §3 (A0-3.7's list). The graph
  holds **references**, never values.
- **A2-9.2** Captured secret material lives only in the evidence store
  (ADR-0009 §2) and is referenced by an `evi_` id: a `credential` node carries
  `evidence_id` plus an opaque `label` (ADR-0020 §4's `CRED-003` style), and
  any other node uses the common `evidence_ids` field. The `credential_kind`
  enum (A2-2.4) names the class; the bytes are never in the graph.
- **A2-9.3** Cloud egress: captured secrets are excluded from every cloud
  endpoint under every policy (ADR-0020 §4). Because A2-9.1 keeps them out of
  the graph entirely, no graph read, stage view (A3) or report path can carry
  them to the LLM gateway; the gateway's exclusion and scanning remain in force
  as defence in depth (ADR-0020 §2, §5).
- **A2-9.4** Ingest scanning: the platform MUST run a stdlib pattern scan
  (`regexp`) over every string field and every `attrs` value for known secret
  shapes (PEM blocks, NTLM/base64 hash shapes, cloud key prefixes, `krbtgt`
  ticket material, high-entropy bearer strings) and MUST reject a match with
  `validation` naming the **field** and the rule — never echoing the value
  (A0-3.4 permits echoing untrusted material; A2-9.5 overrides that here).
  **PO confirm** (§6.10: reject vs redact-and-record).
- **A2-9.5** Secrets never appear in errors or logs (ADR-0019 §5, A0-3.7):
  an error `message`, an error `attrs` entry and any `slog` record MUST NOT
  contain a rejected secret value, a prefix of it, or its hash. The message
  names the field, the rule id and the byte length only.
- **A2-9.6** What a worker MUST do with a captured secret instead (Q6 worker
  addendum — the worker is report-only and has **no** graph access): upload it
  to the evidence store (`evidence:upload`), reference the returned `evi_` id in
  its task result / A1 event, and let the platform create the `credential` node
  with `evidence_id` set and a pseudonymous `label`. The worker MUST NOT put
  the secret in an event payload, a task result, a log line, a model prompt
  (ADR-0020 §4) or any graph-bound content.
- **A2-9.7** Negative tests (shared suite, `contracts/README.md` "secret-free
  serialization"):
  - `TestGraphSecretFreeSerialization` — for a corpus of writes with planted
    secrets (an NTLM-shaped hash, a PEM private key, a cloud access-key id, a
    bearer token) in `label`, `summary`, `attrs` values and edge fields: the
    write is rejected, **and** no planted value appears in the stored node, the
    serialized node, the response body, or the captured `slog` output.
  - `TestCredentialNodeHoldsReferenceOnly` — a `credential` node round-trips
    with `evidence_id` set and contains no field whose value matches a secret
    pattern; the referenced evidence bytes are never fetched by a graph read.
  - `TestNoMaskingMappingInGraph` — no graph field contains an ADR-0020 §3
    mapping value or its inverse.

### A2-10 · Validation semantics and the write path

- **A2-10.1** All graph writes happen **platform-side** (Q6 worker addendum):
  the ingest path turns A1 events and task results into nodes and edges. No
  machine principal holds a graph-write scope; A5 MUST list every graph-write
  verb on the machine-principal exclusion list. Operators write only the
  mutable flags of A2-2.8. Enforcement lives in the platform core, never in the
  agent, the tool image, the UI layer or the store driver (ADR-0005 §5, SPEC
  §6/C5).
- **A2-10.2** One pass, fail fast, in this order — so error messages are
  deterministic and testable: (1) engagement binding (A2-11) · (2) request body
  size bound (A0-8.9, A4 sets the number) · (3) unknown fields (A0-6.2) ·
  (4) id syntax (A0-1.5) · (5) `kind` enum (A0-6.3) · (6) per-kind required /
  not-applicable fields (A2-2.3) · (7) value types and enums (A2-1.7, A2-2.4–6)
  · (8) size caps (A2-7) · (9) `attrs` shape and reserved keys (A2-6) ·
  (10) secret scan (A2-9.4) · (11) edge endpoints, cardinality, cycles
  (A2-3) · (12) provenance establishment (A2-5) · (13) policy evaluation and
  quarantine stamping (A2-8.2). A write MUST report exactly one error: the
  first violated rule.
- **A2-10.3** Error kinds, per A0-3.1 and never invented here: schema, enum,
  id, type, not-applicable-field, endpoint-kind, attrs-shape, secret-scan and
  cycle violations → `validation` (400) · cap violations under mechanism R →
  `summary_too_large` (413, A0-7.6) · immutable-field update, supersede of a
  non-current node, supersede cycle → `conflict` (409) · endpoint or
  `event_id` that does not resolve in this engagement → `notfound` (404,
  A0-3.9) · store uniqueness race or missing platform provenance → `internal`
  (500, A0-1.4).
- **A2-10.4** Every `validation` message is self-contained and greppable per
  ADR-0019 §2 / A0-3.4:
  `component.Function: what was attempted: key identifiers: cause`, naming the
  field, the rule and — where untrusted — the offending *field name*, never a
  secret value (A2-9.5). Example:
  `graph.WriteNode: node create rejected for engagement=eng_01m1y2whfhgbz06ays6dxnvyws kind=service: field=port value=70000 outside 1..65535`.
  The message MUST NOT contain SQL, a stack trace, a request body or a
  canonical-JSON dump (A0-3.5).
- **A2-10.5** Unknown endpoint kind for an edge → `validation` naming the edge
  kind, the offending endpoint kind and the allowed set from A2-3.2 (A0-6.3:
  the closed-list rule covers endpoint kinds, not just the edge kind itself).
- **A2-10.6** The store seam MUST accept only values produced by the validated
  constructors (`graph.NewNode`, `graph.NewEdge`); their fields are unexported
  so no package can build an unvalidated node literal (DESIGN §1 layering,
  ADR-0010). _Testable: the contract suite constructs nodes only through the
  constructors, and a `go vet`-visible exported-field audit fails the build if
  a mutable field appears._
- **A2-10.7** Validation MUST NOT be relaxed for internal callers: ingest,
  operator corrections and report building all pass through the same validator
  (Q3: the platform never trusts discipline — including its own agents').
- **A2-10.8** A rejected write stores nothing (A0-7.6) and MUST still be
  observable: the rejection is recorded as an A1 event so a persistent
  hallucination loop is visible to the operator instead of silently vanishing
  (adversarial A1: untrusted content must not be able to erase its own trace).

### A2-11 · Engagement binding (C8)

- **A2-11.1** Every read and every write is engagement-scoped (ADR-0016 §1/§3,
  SPEC C8). The store seam MUST require an engagement id on every read method;
  no unfiltered or multi-engagement read method MAY exist. _A method that can
  be called without an engagement id will eventually be (adversarial A12)._
- **A2-11.2** No cross-engagement traversal is **expressible**: no endpoint,
  cursor, filter or store method accepts a list of engagement ids, and edges
  cannot span engagements (A2-11.3). Learning is per-engagement only — no
  aggregation, statistics, pattern mining or template reuse across engagements
  or customers in v1; that would need its own ADR with consent and
  anonymization rules (ADR-0016 §3).
- **A2-11.3** An edge write MUST resolve both endpoints inside the caller's
  engagement. An endpoint id from another engagement → `notfound` (404) per
  A0-3.9 — never `forbidden`, never `validation`: the existence of another
  engagement's objects MUST NOT be disclosed (SPEC C8, adversarial A12).
- **A2-11.4** Negative contract tests (shared suite, "cross-engagement
  negatives"):
  - `TestGraphEngagementBReadNeverReturnsA` — with populated graphs in A and B,
    every B-scoped read (node list, edge list, history read, drill-down input)
    returns only B rows; no A `gn_`/`ge_` id appears in any response byte.
  - `TestGraphNodeIDFromAIsNotFoundInB` — a well-formed A id requested in B
    yields `notfound` (404) with an envelope whose `attrs.graph_node_id` is
    either absent or the requested id, and whose `message` does not distinguish
    "absent" from "elsewhere" (A0-3.9).
  - `TestNoCrossEngagementEdge` — an edge whose endpoints lie in A and B is
    rejected `notfound`, and no edge row is created in either engagement.
  - `TestCursorFromEngagementARejectedInB` — replaying A's cursor (A0-4.4) on a
    B-scoped list yields an empty page or `validation`, never A data;
    authorization is re-derived per page, never from the cursor.
  - `TestNoBulkReadSpansEngagements` — reflection/endpoint audit over the A4
    route table and the store seam: no graph read accepts more than one
    engagement id or omits it.
- **A2-11.5** `engagement_id` on a stored record is immutable (A2-1.3). There
  is no re-parenting, no export-to-another-engagement and no copy operation in
  v1; a future one is an ADR (ADR-0016 §3).

### A2-12 · Seam to A3

- **A2-12.1** A2 defines the **shapes** a view may contain; the fixed
  per-stage views and the capped 1-hop drill-down are A3 (Q1). A2 exposes no
  query, traversal or free-form filter (A6 stays a stub).
- **A2-12.2** A3 MAY assume, without recomputing anything:

  | Guarantee | Clause |
  |---|---|
  | node reference shape `{id, kind, label, quarantined}` — every node has all four; `label` ≤ 128 B, `kind` ≤ 32 B | A2-2.2, A2-7.1 |
  | full `summary` ≤ 512 B (≤ 2048 B for `finding`), available on a single-node drill-down read | A0-7.1, A2-7.1 |
  | current-ness is A2's: a view that filters to current nodes uses `superseded_by_id` absence; A3 MUST NOT walk `supersedes` chains itself | A2-1.8, A2-4.4 |
  | deterministic ordering inputs: immutable `seq`, then `id` byte-wise (A0-1.9); no mutable field orders anything | A2-1.4, A2-8.9 |
  | edge endpoint kinds are denormalized and validated (`source_kind`, `target_kind`), so a 1-hop drill-down needs no node fetch to label an edge | A2-3.8 |
  | quarantine is authoritative and already stamped; A3 MUST NOT re-derive scope | A2-8.1/8.2 |
  | per-record size bounds are enforced at ingest, so a view's byte budget is a function of *how many* records it includes, never of how large one can be | A2-7 |
  | every node is anchored to an A1 event via `provenance.event_id` — a view can cite evidence without joining the log | A2-5.4 |

- **A2-12.3** A0-7.10 is **not** resolved here: 500 nodes × 512 B ≈ 250 KiB ≫
  64 KiB, so the caps are not jointly satisfiable for a maximal stage view.
  What A2 guarantees A3 can rely on while it decides: the four-field reference
  above is the smallest complete node identity A2 can offer (~200 B worst case,
  ~60 B typical), a compact reference is enough to render a view row, and full
  summaries are always available one node at a time. A3 owns the composition
  rule (A0 §6.14) and MUST record its own mechanism-T assignments per A0-7.7.
- **A2-12.4** A3 MUST decide, and A2 does not pre-empt: the fixed per-stage
  view definitions and their filters · which reads are *planning-facing* (A2
  requires such reads to omit quarantined nodes, retracted edges and edges with
  quarantined endpoints — A2-8.3/8.4, A2-3.9 — but the mapping of views to
  planning vs reporting is A3's) · the drill-down cap and whether `next_cursor`
  exposes the nodes a view dropped (A0-7.9) · `nodes_truncated` and the stage
  summary (mechanism T) · whether views are transmitted in canonical form
  (A0-7.3 note).
- **A2-12.5** Reporting handover views MUST include quarantined nodes with
  `quarantined` and `quarantine_reason` intact and MUST omit `report_excluded`
  nodes (A2-8.6/8.7); planning views MUST omit quarantined nodes entirely.
  Both rules are A2's; A3 implements them per view.

## 4. Types

Illustrative sketches — **not compiled** (`contracts/README.md`). They are the
source of truth for field names and JSON shapes until `internal/graph` merges
(DESIGN §1). One flat node type per A2-2.3 rather than a per-kind payload
interface: no hand-rolled JSON type dispatch, one fixed key set for A0-2.14,
one validator `switch` (DESIGN §2, ADR-0001/0010 stdlib-only).

```go
// internal/graph — the A2 contract types. Foundation layer: imports internal/ids,
// internal/cjson, internal/errs only (DESIGN §1). No pgx here; the store seam is
// internal/store (ADR-0010).
package graph

type NodeKind string

// A2-2.1: closed list, A0-8.5 spelling.
const (
	KindHost        NodeKind = "host"
	KindNetwork     NodeKind = "network"
	KindService     NodeKind = "service"
	KindIdentity    NodeKind = "identity"
	KindGroup       NodeKind = "group"
	KindCredential  NodeKind = "credential"
	KindShare       NodeKind = "share"
	KindEvidenceRef NodeKind = "evidence_ref"
	KindFinding     NodeKind = "finding"
	KindHypothesis  NodeKind = "hypothesis"
)

type EdgeKind string

// A2-3.1: closed list.
const (
	EdgeReachable       EdgeKind = "reachable"
	EdgeAuthenticatesTo EdgeKind = "authenticates_to"
	EdgeMemberOf        EdgeKind = "member_of"
	EdgeGrantsAccess    EdgeKind = "grants_access"
	EdgeExploitedBy     EdgeKind = "exploited_by"
	EdgeContradicts     EdgeKind = "contradicts"
	EdgeSupersedes      EdgeKind = "supersedes"
)

type Severity string // A2-2.5

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type FindingStatus string    // A2-2.6
type HypothesisStatus string // A2-2.6

const (
	FindingOpen       FindingStatus = "open"
	FindingConfirmed  FindingStatus = "confirmed"
	FindingRefuted    FindingStatus = "refuted"
	FindingRemediated FindingStatus = "remediated"

	HypothesisOpen      HypothesisStatus = "open"
	HypothesisSupported HypothesisStatus = "supported"
	HypothesisRefuted   HypothesisStatus = "refuted"
)

type Confidence string // A2-5.6: evidence grade, not optimism.

const (
	ConfidenceObserved Confidence = "observed"
	ConfidenceInferred Confidence = "inferred"
	ConfidenceVerified Confidence = "verified"
)

type QuarantineReason string // A2-8.1

const (
	QuarantineOutOfScope QuarantineReason = "out_of_scope"
	QuarantineBlacklist  QuarantineReason = "blacklisted"
)

type PrincipalKind string // A2-5.3

const (
	PrincipalPlatform     PrincipalKind = "platform"
	PrincipalOrchestrator PrincipalKind = "orchestrator"
	PrincipalWorker       PrincipalKind = "worker"
	PrincipalOperator     PrincipalKind = "operator"
)

type CredentialKind string // A2-2.4: class of material, never the material (A2-9).
type MediaKind string      // A2-2.4

// Attrs is the bounded flat escape hatch (A2-6). Depth exactly 1: no arrays,
// no objects, no null, no floats (A2-1.7, A0-2.6). Custom MarshalJSON /
// UnmarshalJSON on stdlib encoding/json only; UnmarshalJSON rejects a nested
// value with errs kind "validation" (A2-6.1) and rejects a reserved key (A2-6.3).
type Attrs map[string]AttrValue

type AttrType string

const (
	AttrString AttrType = "string"
	AttrInt    AttrType = "int"
	AttrBool   AttrType = "bool"
)

type AttrValue struct {
	Type AttrType // which of Str/Num/Bool is live
	Str  string   // ≤ AttrValueMaxBytes (A2-7.1)
	Num  int64    // within A0-2.6
	Bool bool
}

// A2-local cap constants (A2-7.1). Q4 constants come from A0-7.1 and are not
// redeclared here; §6.4 asks A0 to adopt the A2-local ones into its table.
const (
	NodeLabelMaxBytes       = 128
	HypothesisClaimMaxBytes = 512
	HypothesisBasisMaxBytes = 1024
	AttrValueMaxBytes       = 512
	AttrsTotalMaxBytes      = 4096
	AttrsMaxKeys            = 16
	EvidenceIDsMax          = 8
	AddressesMax            = 16
	AddressMaxBytes         = 64
	ToolVersionMaxBytes     = 32
	MaxSupersedeChain       = 64 // A2-4.4/4.5: bound on one history walk
)

// Provenance is mandatory on every node and edge (A2-5.1) and is
// platform-stamped only (A2-5.2). Absent optional fields are omitted, never
// null (A0-8.3).
type Provenance struct {
	PrincipalKind     PrincipalKind `json:"principal_kind"`
	RunID             string        `json:"run_id"`
	JobID             string        `json:"job_id,omitempty"`
	TaskID            string        `json:"task_id,omitempty"`
	AgentNodeID       string        `json:"agent_node_id,omitempty"` // slp_node_ (Q9), never "node_id" (A2-1.6)
	OperatorID        string        `json:"operator_id,omitempty"`   // blocked until A0/A5 register the prefix (§6.2)
	ToolID            string        `json:"tool_id,omitempty"`       // tool_ (A0-1.3)
	ToolVersion       string        `json:"tool_version,omitempty"`  // registry string, not part of tool_id
	EventID           string        `json:"event_id"`                // originating A1 event (A2-5.4)
	RecordedAt        string        `json:"recorded_at"`             // platform time, A0-5.1/5.4
	ObservedClaimedAt string        `json:"observed_claimed_at,omitempty"` // untrusted (A0-5.7)
	Confidence        Confidence    `json:"confidence"`
}

// QuarantineState is the platform-internal value the policy engine returns at
// ingest (A2-8.2). It is serialized flat onto Node, not as a nested object.
type QuarantineState struct {
	Quarantined bool
	Reason      QuarantineReason // "" when not quarantined
}

// Node is one graph node. Fields are immutable except the quarantine and
// report flags (A2-1.3, A2-8). Unexported in the implementing package: only
// NewNode may build one (A2-10.6).
type Node struct {
	ID           string   `json:"id"`           // gn_ (A0-1.2)
	EngagementID string   `json:"engagement_id"`
	Seq          int64    `json:"seq"`          // immutable order key (A2-1.4)
	Kind         NodeKind `json:"kind"`

	Label   string   `json:"label"`
	Summary string   `json:"summary,omitempty"`
	Attrs   Attrs    `json:"attrs,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"` // evi_ (A2-9.2)

	// Kind-specific (A2-2.3); a field not applicable to Kind MUST be absent.
	Addresses      []string       `json:"addresses,omitempty"`       // host, network
	CIDR           string         `json:"cidr,omitempty"`            // network (net.ParseCIDR)
	Port           int            `json:"port,omitempty"`            // service, 1..65535
	Transport      string         `json:"transport,omitempty"`       // service: tcp|udp
	Protocol       string         `json:"protocol,omitempty"`        // service
	SID            string         `json:"sid,omitempty"`             // identity, group
	Domain         string         `json:"domain,omitempty"`          // identity, group, credential, share
	CredentialKind CredentialKind `json:"credential_kind,omitempty"` // credential
	EvidenceID     string         `json:"evidence_id,omitempty"`     // credential, evidence_ref
	MediaKind      MediaKind      `json:"media_kind,omitempty"`      // evidence_ref
	SizeBytes      int64          `json:"size_bytes,omitempty"`      // evidence_ref (A0-8.2)
	Severity       Severity       `json:"severity,omitempty"`        // finding
	Claim          string         `json:"claim,omitempty"`           // hypothesis
	Basis          string         `json:"basis,omitempty"`           // hypothesis

	// Status: FindingStatus or HypothesisStatus per Kind (A2-2.6). One JSON
	// field, two closed enums, validated by the kind switch.
	Status string `json:"status,omitempty"`

	ContentHash    string           `json:"content_hash"`              // A2-4.6, A0-8.7
	Quarantined    bool             `json:"quarantined"`               // A2-8.1, always present
	QuarantineReason QuarantineReason `json:"quarantine_reason,omitempty"`
	ReportExcluded bool             `json:"report_excluded"`           // A2-8.7, always present
	SupersedesID   string           `json:"supersedes_id,omitempty"`   // A2-4.2
	SupersededByID string           `json:"superseded_by_id,omitempty"`
	Provenance     Provenance       `json:"provenance"`
}

// Edge is one directed relationship (A2-3). Content is immutable; only
// Retracted may change (A2-3.9).
type Edge struct {
	ID           string    `json:"id"` // ge_
	EngagementID string    `json:"engagement_id"`
	Seq          int64     `json:"seq"`
	Kind         EdgeKind  `json:"kind"`
	SourceID     string    `json:"source_id"` // gn_
	SourceKind   NodeKind  `json:"source_kind"`
	TargetID     string    `json:"target_id"` // gn_
	TargetKind   NodeKind  `json:"target_kind"`
	Retracted    bool      `json:"retracted"`
	Provenance   Provenance `json:"provenance"`
}

// contentDoc is the purpose-built document content_hash is computed over
// (A2-4.6). Fixed key set, zero values for inapplicable fields (A0-2.14),
// empty A0-2.12 exclusion list, canonicalized by internal/cjson (A0-2).
type contentDoc struct {
	Addresses      []string       `json:"addresses"`
	Attrs          Attrs          `json:"attrs"`
	Basis          string         `json:"basis"`
	CIDR           string         `json:"cidr"`
	Claim          string         `json:"claim"`
	CredentialKind CredentialKind `json:"credential_kind"`
	Domain         string         `json:"domain"`
	EvidenceID     string         `json:"evidence_id"`
	EvidenceIDs    []string       `json:"evidence_ids"`
	Kind           NodeKind       `json:"kind"`
	Label          string         `json:"label"`
	MediaKind      MediaKind      `json:"media_kind"`
	Port           int            `json:"port"`
	Protocol       string         `json:"protocol"`
	Severity       Severity       `json:"severity"`
	SID            string         `json:"sid"`
	SizeBytes      int64          `json:"size_bytes"`
	Status         string         `json:"status"`
	Summary        string         `json:"summary"`
	Transport      string         `json:"transport"`
}

// NewNode validates (A2-10.2 order) and returns a node ready for the store
// seam; NewEdge does the same for edges. Both are the only constructors
// (A2-10.6). Errors carry an A0-3 kind: validation, summary_too_large,
// conflict, notfound (A2-10.3).
func NewNode(in NodeDraft, prov Provenance, q QuarantineState) (Node, error)
func NewEdge(in EdgeDraft, prov Provenance) (Edge, error)

// NodeDraft / EdgeDraft are the ingest-side inputs: no id, no engagement_id,
// no seq, no provenance, no quarantine fields — the platform supplies all of
// them (A2-1.5, A2-5.2, A2-8.2). A draft that carries one is an unknown field
// on the wire (A0-6.2).
type NodeDraft struct{ /* Node minus platform-set fields */ }
type EdgeDraft struct{ /* Kind, SourceID, TargetID */ }

// CurrentNode reports whether n is the current revision (A2-1.8).
func (n Node) Current() bool { return n.SupersededByID == "" }
```

### 4.1 JSON examples

A `finding` node with full provenance (A2-2.3, A2-5.3). `quarantined` and
`report_excluded` are present and `false`; unset optional fields are absent
(A0-8.3).

```json
{
  "id": "gn_01m1y2whfhh039ykj5x8mc5a0g",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "seq": 412,
  "kind": "finding",
  "label": "SMB relay to SYSVOL on dc01",
  "summary": "Captured NTLM authentication from 10.20.0.14 was relayed to the SYSVOL share on dc01, yielding read access to group policy preferences. Secret material is referenced, not stored (evi_).",
  "evidence_ids": [
    "evi_01m1y2whfh3ca875z2x8v8h7qt",
    "evi_01m1y2whfh7kq2m4c8x1z9vb3n"
  ],
  "attrs": {
    "cvss_v3_x10": 88,
    "first_seen_task": "task_01m1y2whfh1txm57x8dn41r9hg",
    "relay_tool": "ntlmrelayx"
  },
  "severity": "high",
  "status": "confirmed",
  "content_hash": "7f3c1a92de48b06f5ac7d1e8b93042fa6c5d7e81b2a39f04c6d81e5b7a290c34",
  "quarantined": false,
  "report_excluded": false,
  "provenance": {
    "principal_kind": "worker",
    "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
    "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
    "task_id": "task_01m1y2whfh1txm57x8dn41r9hg",
    "agent_node_id": "slp_node_01m1y2whfhxydsaem68cmazyc8",
    "tool_id": "tool_01m1y2whfhfjdvwqp9pfxqekmf",
    "tool_version": "1.4.2",
    "event_id": "evt_01m1y2whfhp17g0avdqztd2p3x",
    "recorded_at": "2026-09-07T14:03:22.481Z",
    "observed_claimed_at": "2026-09-07T14:03:19.900Z",
    "confidence": "verified"
  }
}
```

A `supersedes` revision pair (A2-4): the old hypothesis is unchanged and
gains only `superseded_by_id`; the new one carries `supersedes_id`. Both stay
in the graph, both keep their own provenance and `content_hash`.

```json
{
  "items": [
    {
      "id": "gn_01m1y2whfh9x2b4c7d1e8f0a3b",
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "seq": 388,
      "kind": "hypothesis",
      "label": "dc01 accepts NTLM relay from the VPN range",
      "claim": "NTLM authentication from 10.20.0.0/24 can be relayed to SMB on dc01 because signing is not enforced.",
      "basis": "Port 445 reachable (evt_01m1y2whfhc2v9nq4x7z1m8b3p) and no SMB signing requirement observed in one negotiation.",
      "status": "supported",
      "attrs": { "negotiations_seen": 1 },
      "content_hash": "1c0f8ab73e5d9246b0a17fe4c8d2593ba6e0147d92c85b3f10ae6d47c8b92015",
      "quarantined": false,
      "report_excluded": false,
      "superseded_by_id": "gn_01m1y2whfjk5t8nq2z7x1vb3rt",
      "provenance": {
        "principal_kind": "orchestrator",
        "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
        "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
        "event_id": "evt_01m1y2whfhd8k3m9qz1x7vb4nr",
        "recorded_at": "2026-09-07T13:41:07.220Z",
        "confidence": "inferred"
      }
    },
    {
      "id": "gn_01m1y2whfjk5t8nq2z7x1vb3rt",
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "seq": 405,
      "kind": "hypothesis",
      "label": "dc01 accepts NTLM relay from the VPN range",
      "claim": "NTLM relay from 10.20.0.0/24 to SMB on dc01 succeeds and reaches SYSVOL; three independent negotiations reproduced it.",
      "basis": "Reproduced by task_01m1y2whfh1txm57x8dn41r9hg against dc01 and by a second run against the file server; signing absent in all three negotiations.",
      "status": "supported",
      "attrs": { "negotiations_seen": 3 },
      "content_hash": "9ab41f0c7d2e8356b1a04ce7f8d2953ba6e1047d82c95b4f20ae7d57c9b83125",
      "quarantined": false,
      "report_excluded": false,
      "supersedes_id": "gn_01m1y2whfh9x2b4c7d1e8f0a3b",
      "provenance": {
        "principal_kind": "worker",
        "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
        "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
        "task_id": "task_01m1y2whfh1txm57x8dn41r9hg",
        "tool_id": "tool_01m1y2whfhfjdvwqp9pfxqekmf",
        "tool_version": "1.4.2",
        "event_id": "evt_01m1y2whfhm6p3kq9z2x1vb4rt",
        "recorded_at": "2026-09-07T13:58:41.330Z",
        "observed_claimed_at": "2026-09-07T13:58:39.100Z",
        "confidence": "verified"
      }
    }
  ]
}
```

_The pair above is a hypothesis revised by a better-supported hypothesis
(same `kind`, A2-4.3); a confirmed exploitation is written as a `finding`
(first example) and attached to its target by the `exploited_by` edge below._

A quarantined out-of-scope host (A2-8), and the edge that attaches a finding
to its target (A2-3.2):

```json
{
  "id": "gn_01m1y2whfhq7z3m9x1c4vb8nrt",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "seq": 431,
  "kind": "host",
  "label": "payroll.corp.acme.com",
  "addresses": ["10.99.4.20"],
  "summary": "Seen in DNS response for an in-scope resolver. Outside the engagement allowlist and inside the global blacklist range 10.99.0.0/16: recorded, never tested.",
  "attrs": { "seen_via": "dns_zone_transfer_partial" },
  "content_hash": "4e7b0c93da18f2650be17ca4d83f952ba0e6174d89c25b7f31ae0d47c2b85913",
  "quarantined": true,
  "quarantine_reason": "blacklisted",
  "report_excluded": false,
  "provenance": {
    "principal_kind": "platform",
    "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
    "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
    "event_id": "evt_01m1y2whfhk4m2nq8x7z1vb3rt",
    "recorded_at": "2026-09-07T14:11:52.007Z",
    "observed_claimed_at": "2026-09-07T14:11:48.310Z",
    "confidence": "observed"
  }
}
```

```json
{
  "id": "ge_01m1y2whfh62ej11jf4x5gjzv4",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "seq": 413,
  "kind": "exploited_by",
  "source_id": "gn_01m1y2whfhdc01srv4x8mc5a0g",
  "source_kind": "host",
  "target_id": "gn_01m1y2whfhh039ykj5x8mc5a0g",
  "target_kind": "finding",
  "retracted": false,
  "provenance": {
    "principal_kind": "platform",
    "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
    "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
    "event_id": "evt_01m1y2whfhp17g0avdqztd2p3x",
    "recorded_at": "2026-09-07T14:03:22.502Z",
    "confidence": "verified"
  }
}
```

A hard-rejected write (Q3, A2-10.3/10.4): a cap violation is
`summary_too_large` (A0-7.6), everything else `validation`. Note
`graph_node_id`, never `node_id`, in `attrs` (A0-3.6).

```json
{
  "error": {
    "kind": "validation",
    "message": "graph.NewEdge: edge create rejected for engagement=eng_01m1y2whfhgbz06ays6dxnvyws kind=member_of: source_kind=service not in allowed set [identity group]",
    "attrs": {
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
      "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
      "graph_node_id": "gn_01m1y2whfhdc01srv4x8mc5a0g"
    }
  }
}
```

## 5. Traceability

| Source | Decision | Clauses |
|---|---|---|
| **ADR-0016 §1** | per-engagement graph; closed node kinds (hosts, networks, services, accounts/identities, groups, credentials, shares, findings/hypotheses, artifacts/evidence refs); closed edge kinds; provenance on every node/edge; property-graph tables in PostgreSQL | A2-1.1, A2-2.1–2.4, A2-3.1–3.2, A2-5.1–5.6 |
| **ADR-0016 §2** | scope alignment enforced on the graph: out-of-scope discoveries recorded as quarantined, never actionable; blacklist not representable as an actionable target; policy checks in the platform core on planning reads | A2-8.1–8.5, A2-10.1, A2-12.4/12.5 |
| **ADR-0016 §3** | learning per-engagement only; no cross-engagement flow; aggregation deferred to its own ADR | A2-11.1, A2-11.2, A2-11.5 |
| **ADR-0016 §4** | orchestrator plans from views; workers write findings; handovers are views; self-correction via `contradicts`; revision by a new node + `supersedes`, history never overwritten | A2-4.1–4.5, A2-3.6, A2-12.1, A2-12.4 |
| **ADR-0016 §5** | PostgreSQL property-graph tables, no external graph DB | A2-1.1, A2-10.6 |
| **Q1** | fixed stage views + capped 1-hop drill-down; no free-form query in v1 | A2-12.1, A2-12.2, A2-12.4, A2-4.4 (chain walk without a query), §2 (A6 stub) |
| **Q2** | `finding` and `hypothesis` as separate typed kinds; revision via `supersedes`, never overwrite | A2-2.1, A2-2.3, A2-2.5–2.7, A2-4.1–4.5, A2-1.3 |
| **Q3** | hard reject on schema violation + bounded `attrs`; platform enforces caps, never trusts worker discipline | A2-6.1–6.4, A2-7.1–7.3, A2-10.2–10.8, A2-2.1 (unknown kind), A2-3.1 (unknown edge kind), A2-2.6 (unknown status) |
| **Q4** | size budgets as contract constants | A2-7.1 (table incl. `NodeSummaryMaxBytes` 512 B, `FindingSummaryMaxBytes` 2 KiB), A2-7.3, A2-12.3 |
| **Q5** | quarantined discoveries in the reporting handover marked *not tested*, non-actionable, operator can remove them from the report | A2-8.3, A2-8.6, A2-8.7, A2-8.8, A2-12.5 |
| **Q6 + worker addendum** | worker report-only, no graph access; graph writes platform-side with platform-stamped provenance; three enforcement layers incl. per-request engagement binding; machine principals never mutate scope/blacklist | A2-5.2, A2-10.1, A2-1.5, A2-8.2, A2-8.5, A2-8.8, A2-11.1, A2-9.6 |
| **SPEC C8 / ADR-0016 §3** | learning/memory strictly per-engagement, no cross-customer data flow | A2-11.1–11.5 |
| **SPEC §6** | enforcement point is the platform core; blacklist beats allowlist beats approval; out-of-scope discoveries are quarantined nodes; untrusted content never executed as configuration | A2-10.1, A2-8.2, A2-8.3, A2-6.7 |
| **SPEC §8** | context graph = per-engagement nodes/edges with provenance; evidence store referenced from events and graph nodes | A2-1.1, A2-5.1, A2-9.2 |
| **ADR-0005 §2/§3/§5** | allowlist scope, blacklist beats allowlist and cannot be overridden, scope/blacklist checks in the platform core, agents never police themselves | A2-8.2, A2-8.3, A2-8.5, A2-10.1, A2-10.7 |
| **ADR-0009 §1/§2** | central append-only log doubles as evidence trail; evidence artifacts stored per engagement and referenced | A2-5.4, A2-9.2, A2-3.4, A2-10.8 |
| **ADR-0019 §2** | self-contained, greppable error messages | A2-10.4, A2-6.4, A2-5.7 |
| **ADR-0019 §5** | mandatory redaction: no captured credentials/tokens/secret material in errors or logs | A2-9.1, A2-9.5, A2-5.8, A2-9.7 |
| **ADR-0020 §3/§4** | per-run masking-mapping values are secret; captured credentials/hashes/tokens never reach a cloud endpoint under any policy; agents use opaque references | A2-9.1, A2-9.2, A2-9.3, A2-9.7 |
| **ADR-0008 / Q14** | tool registry id + version; name/version are registry fields, not part of the id | A2-5.3 (`tool_id`, `tool_version`), A0-1.3 cited |
| **Adversarial A1** | prompt injection / untrusted content: graph content is untrusted, never configuration, and a rejected write still leaves a trace | A2-6.7, A2-10.8, A2-5.2 |
| **Adversarial A12** | cross-engagement leakage | A2-11.1–11.4, A2-5.7 (`notfound`), A2-3.2 (`notfound` endpoints) |
| **A0-1.2** | `gn_` for graph nodes, `ge_` for graph edges, `evi_` for evidence | A2-1.2, A2-2.2, A2-3.1, §4 |
| **A0-3.6** | in errors and log attrs a graph node is `graph_node_id`; `node_id` means the remote agent node — never conflated | A2-1.6, A2-5.3 (`agent_node_id`), §4.1 error example |
| **A0-7.1 / A0-7.7** | `NodeSummaryMaxBytes` 512 B and `FindingSummaryMaxBytes` 2 KiB apply; every capped field class gets exactly one mechanism, recorded in a table | A2-7.1 (mechanism table, R for all A2 classes), A2-7.2, A2-2.2/2.3 |
| **A0-6.3** | unknown node kind, edge kind or status enum on write = hard reject `validation` | A2-2.1, A2-3.1, A2-2.6, A2-10.2 (steps 5, 7, 11), A2-10.5 |
| **A0-3.9** | an id from another engagement resolves to `notfound` (404), never `forbidden` — no existence disclosure | A2-11.3, A2-11.4, A2-5.7, A2-3.2 |
| **A0-2.14** | a canonicalized value has a fixed key set | A2-4.6 (`contentDoc`, empty A0-2.12 exclusion list), A2-4.8 (A0-2.16 stored bytes), §4 `contentDoc` |
| **A0-5.7** | client-supplied time is `*_claimed_at` and untrusted | A2-5.3 (`observed_claimed_at`), A2-5.5 |
| **A0-8.8** | booleans are adjectives, no `is_` prefix | A2-8.1 (`quarantined`), A2-8.7 (`report_excluded`), A2-3.9 (`retracted`), A2-2.2 |
| **A0-4.3** | pagination order from immutable keys, never from a mutable flag | A2-1.4 (`seq`), A2-8.9, A2-4.5 |
| **A0-7.10** | the Q4 caps are not jointly satisfiable for a maximal stage view; the composition rule belongs to A3 | A2-12.2, A2-12.3 (what A2 guarantees A3 can rely on), §6.11 |
| **A0-6.2 / A0-8.3 / A0-8.5 / A0-8.2 / A0-2.6 / A0-8.9** | unknown field on write rejected; absent not null; closed enum spelling; fixed suffixes; integers only with the scale fixed by the owning contract; bounded bodies | A2-1.5, A2-1.7, A2-2.1/2.4–2.6, A2-1.6, A2-2.5 (`cvss_v3_x10` ×10), A2-10.2 step 2 |
| **ADR-0001 / ADR-0010** | stdlib only; pgx only behind the store seam | §4 (no third-party type; `net.ParseCIDR`, `regexp`, `encoding/json` only), A2-1.1, A2-10.6 |

## 6. Open for product owner

Recommendations that are genuinely product-owner calls. Each is marked **PO
confirm** at the clause too; none is decided silently. A2 is written against
**current** A0 throughout — nothing below is assumed to be already granted.

1. **A2-5.6 / A2-2.7 — confidence semantics.** Recommend the evidence grade
   `observed` · `inferred` · `verified` instead of `low`/`medium`/`high`.
   _ADR-0016 §1: graph content is evidence, not opinion — a grade tied to a
   referenced event is checkable, an adjective about certainty is not._ Also
   confirms that Q2's "Finding carries confidence" is discharged by the
   mandatory provenance block rather than a second finding field.
2. **A2-5.3 — `operator_id` has no id shape.** A0-1.2 registers no prefix for a
   human principal, so operator attribution is currently unvalidatable (A0-1.5).
   Recommend A0 add `usr_` (or delegate the spelling to A5) — **A0 amendment
   request**, see below. Until it lands, `principal_kind: operator` writes
   cannot satisfy A2-5.3 and MUST be refused; that blocks operator corrections
   and A2-8.7 report exclusion, so this is on the critical path for A2 Frozen.
3. **A2-8.5 — is a blacklisted discovery recorded at all?** Recommend: recorded
   with `quarantine_reason: blacklisted`, permanently non-releasable, plus an
   A1 event (A2-8.10). _"We saw the forbidden host and did not touch it" is
   the most valuable line in a customer report; refusing the write would leave
   the near-miss unprovable._ The alternative reading of ADR-0016 §2
   ("cannot be represented at all" = refuse the write) is defensible and
   minimizes stored data about a forbidden system — PO call.
4. **A2-8.7 — report exclusion by flag + event, not deletion.** Recommend
   `report_excluded` plus a mandatory A1 event; no graph delete endpoint in v1.
   _Deletion breaks `supersedes` chains, dangling `evidence_ids`, and the
   auditability of who removed what from a customer report (Q5 says "operator
   can remove them from the report", not "from the graph")._
5. **A2-3.9 — mutable `retracted` on edges.** Recommend keeping it.
   _Self-correction (ADR-0016 §4) needs a way to withdraw a wrong
   `reachable`/`grants_access` edge; the alternative — leave edges
   uncorrectable and revise endpoint nodes — cannot express "that observation
   was wrong" at all._ It is the only mutable edge field, it never overwrites
   content, and A2-8.9 keeps it out of every ordering.
6. **A2-3.2 — `exploited_by` may target a `hypothesis`.** ADR-0016 §1 lists
   seven edge kinds and no way to attach a hypothesis to the host it is about;
   extending `exploited_by`'s target set to `{finding, hypothesis}` closes the
   gap without inventing an eighth kind. Direction stays source = the affected
   node, target = the claim; the inverse is never stored.
7. **A2-2.6 — status enums.** `finding`: `open` · `confirmed` · `refuted` ·
   `remediated`; `hypothesis`: `open` · `supported` · `refuted`. No
   `superseded` value (A2-4 owns that). _Report wording follows these values,
   so they are customer-visible._
8. **A2-2.5 — severity scale.** `info` · `low` · `medium` · `high` ·
   `critical`, no CVSS field in the schema; a score MAY ride in `attrs` as
   `cvss_v3_x10` (integer ×10, the scale A0-2.6 requires the owning contract
   to fix). _A CVSS vector string is a 2 KiB free-text blob that would eat the
   `FindingSummaryMaxBytes` budget and is not needed for v1 routing._
9. **A2-5.4 — A1 was not on disk when A2 was drafted.** The A1 envelope fields
   A2 consumes (`event_id`, `kind`, `recorded_at`, `seq`, `engagement_id`,
   `run_id`, `job_id`) are taken from A0-2.17 vector V5 and A0-3.6. If A1 names
   them differently, A2-5.3/5.4 follow A1 (A1 owns the envelope). A2 also needs
   A1 to provide event kinds for: node created, node quarantined, quarantine
   recomputed, node report-excluded, node superseded, edge retracted, write
   rejected (A2-8.5/8.7/8.10, A2-3.9, A2-4.1, A2-10.8).
10. **A2-9.4 — secret scan outcome.** Recommend hard reject (`validation`,
    field + rule named, value never echoed). Alternative: redact the value,
    store a marker, keep the node. _Reject is the Q3 posture and keeps the
    graph clean; redact risks a false positive silently destroying evidence —
    but reject risks stalling an engagement on a false positive, which is why
    the pattern set must ship with the contract test corpus._
11. **A2-12.3 — A0-7.10 is explicitly not resolved here.** A2 states only what
    A3 may rely on (A2-12.2). The 500-node/64 KiB composition rule stays with
    A3 and the PO escalation already recorded in A0 §6.14.

### A0 amendment requests

A2 is written against current A0; these are requests, not assumptions.

| # | A0 clause | Request | Why A2 needs it |
|---|---|---|---|
| AM-1 | A0-1.2 | Register a prefix for a human operator principal (`usr_` recommended), or state that A5 owns it | A2-5.3 `operator_id` cannot be validated (A0-1.5) without one; blocks operator corrections and A2-8.7 |
| AM-2 | A0-7.1 | Adopt the A2-local cap constants of A2-7.1 (`NodeLabelMaxBytes`, `HypothesisClaimMaxBytes`, `HypothesisBasisMaxBytes`, `AttrValueMaxBytes`, `AttrsTotalMaxBytes`, `AttrsMaxKeys`, `EvidenceIDsMax`, `AddressesMax`, `AddressMaxBytes`) into the single A0 table/const block | A0-7.7 delegates *mechanism* assignment to A2 but A0-7.1 is the one place caps live; two const blocks will drift |
| AM-3 | A0-3.6 | Extend the rule from "errors and log attrs" to graph payloads: `node_id` is reserved for the remote agent node everywhere, and A2's payload spellings (`source_id`, `target_id`, `supersedes_id`, `superseded_by_id`, `agent_node_id`) are the blessed names | A2-1.6 currently states a payload convention that A0-3.6 does not literally cover |
| AM-4 | A0-8.2 | Confirm `*_ref` is not needed: A2 uses `evidence_id` / `evidence_ids` for `evi_` references because A0-8.2 fixes `*_id` for identifiers | Avoids a second suffix convention for the same thing (A2-9.2) |

### Cross-contract requests (not A0)

- **A1** — envelope field names (item 9) and the event kinds listed there; the
  ingest dedup key A2 relies on for idempotent replay (A2-4.7, A0-3.11).
- **A3** — the A0-7.10 composition rule, the planning-vs-reporting view
  mapping (A2-12.4), and its own mechanism-T assignments (A0-7.7).
- **A4** — per-endpoint body size bounds (A0-8.9, A2-10.2 step 2), the history
  read parameter (A2-4.5), route-table audit for A2-11.4's
  `TestNoBulkReadSpansEngagements`.
- **A5** — every graph-write verb on the machine-principal exclusion list
  (A2-10.1, Q6); the operator principal id shape (AM-1).
