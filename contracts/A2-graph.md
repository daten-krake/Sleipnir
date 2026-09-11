# A2 — Engagement context graph

## 1. Header

| | |
|---|---|
| **Contract id** | A2 |
| **Status** | `Draft` (`contracts/README.md` lifecycle: Draft → Frozen → Implemented) |
| **Owner** | architect |
| **Gates** | A3 (stage views) · A4 (`/api/v1` graph endpoints) · A5 (token scopes: graph writes are excluded for machine principals) · `internal/graph`, `internal/policy`, the ingest path, the report builder |
| **Implements** | ADR-0016 §1–§5 · ADR-0005 §2/§3/§5 · ADR-0009 §1–§2 · ADR-0019 §2/§5 · ADR-0020 §3–§4 · SPEC §6, §8, C8 · Q1, Q2, Q3, Q4, Q5, Q6 (+ Q6 worker addendum) · adversarial A1, A12 |
| **Depends on** | A0 (**Draft** at the time of writing; every work package below is gated on the A0/A1/A2 Freeze PR) — ids A0-1.2, canonical JSON A0-2, error kinds A0-3, paging A0-4, time A0-5, unknown fields A0-6, size caps A0-7, field conventions A0-8 · A1 (event envelope fields consumed by provenance, A2-5.4) |
| **Authority** | ADR > SPEC > DESIGN > contract. A clause here that contradicts an Accepted ADR is a defect in this document. A0 owns every cross-cutting convention; A2 cites A0 clause ids instead of restating them, and requests amendments only in §6. |

## 2. Scope

**Fixed here:** the closed node-kind and edge-kind lists with per-kind
required/optional fields, allowed value types and size caps; the revision
rule (`supersedes`) and how a client resolves the current revision; mandatory
platform-stamped provenance; the bounded flat `attrs` escape hatch; the
quarantine state and report exclusion; the no-secret-values rule; write
validation semantics and error kinds; engagement binding (C8) and its
negative tests; the exact shapes A3 may assume; the normative content
fingerprint vector (§4.2) and the A2 contract-test ids (§4.3).

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
  construct or reuse an id from another engagement (A0-1.6, A2-11). The
  served field names are `graph_node_id` (nodes) and `graph_edge_id` (edges),
  never a bare `id` (A2-1.6).
- **A2-1.3** Node and edge **content is immutable** (ADR-0016 §4, Q2). The only
  mutable fields in A2 are `quarantined`, `quarantine_reason`,
  `report_excluded` (A2-8) and `retracted` on an edge (A2-3.9). An attempt to
  change any other field of an existing node or edge MUST return `conflict`
  (A0-3.1: immutable field) and MUST NOT partially apply.
- **A2-1.4** Every node and edge carries `graph_seq`: a per-engagement,
  platform-assigned, strictly increasing integer, immutable and unique.
  `graph_seq` is the ordering key for every paginated graph collection
  (A0-4.3) and the `k` of a cursor (A0-4.4). Clients MUST NOT derive order
  from id text (A0-1.6).
- **A2-1.4a** `seq` is assigned by the platform inside the transaction that
  inserts the row, from **one per-engagement graph sequence shared by nodes
  and edges**, strictly increasing by 1, dense, and independent of the A1
  event `seq` (A1-5.4). To keep the two apart in code, logs and views the
  graph field is named **`graph_seq`** in Go fields, store columns **and
  JSON**; A2-8.9's ordering tuple is `(graph_seq, graph_node_id)`/
  `(graph_seq, graph_edge_id)` and A0-4.4's cursor `k` is that value.
  `content_hash` is unaffected: A2-4.6's 20 keys contain no `seq`.
- **A2-1.5** Every node and edge carries `engagement_id`, stamped by the
  platform from the authenticated request binding (Q6, third enforcement
  layer). It MUST NOT be accepted from a request body: the write types have no
  such field, so a body that carries one is an unknown field → `validation`
  (A0-6.2).
- **A2-1.6** Field naming inside graph payloads: a bare `node_id` MUST NOT
  appear — A0-3.6 reserves `node_id` for the **remote agent node**
  (`slp_node_`). Graph-node references are `source_id`, `target_id`,
  `supersedes_id`, `superseded_by_id`; the agent node is `agent_node_id`; and
  error/log attributes use `graph_node_id` for a graph node (A0-3.6). AM-3
  asked A0 to bless the payload spellings: the reservation is granted, the
  blessing is not, and the table below is the published mapping instead.

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

  This table is the A1↔A2 field mapping (PAIR-M1): A1-3.6 cites it and A1-2.1
  carries the identity half (BLOCK-A2-10). `content_hash` is unaffected — no
  `id` key appears in A2-4.6.
- **A2-1.7** Value types are closed: UTF-8 strings (A0-2.3, capped per A2-7),
  integers within A0-2.6, booleans (A0-8.8), closed enum strings (A0-8.5),
  ids (A0-1), timestamps (A0-5.1) and bounded lists of those. Floats, `null`
  (A0-8.3) and nested objects MUST NOT appear in any node or edge field.
  `attrs` (A2-6) is the only nested structure and is itself restricted to flat
  scalar values: no float, no `null`, no array, no deeper object.
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
  | `provenance` | bounded list, ≤ 8 entries | platform-set, mandatory | A2-5, A2-4.7 |
  | `graph_node_id`/`graph_edge_id`, `engagement_id`, `graph_seq`, `kind` | per A2-1 | platform-set | A2-1.2/1.4/1.4a/1.5/1.6 |

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
  | `evidence_ref` | `label`, `evidence_id` (`evi_`), `media_kind` | `size_bytes` int ≥ 1 (A0-8.2) |
  | `finding` | `label`, `summary`, `severity`, `status` | `evidence_ids`, `attrs` |
  | `hypothesis` | `label`, `claim`, `basis`, `status` | `evidence_ids`, `attrs` |

  `cidr` MUST parse with `net.ParseCIDR` (stdlib) — an unparseable value is
  `validation`, never a stored string (A2-10.4). `size_bytes` MUST be ≥ 1 — a
  zero-byte artifact is not an artifact — so absence and zero cannot be
  confused in the served representation (A0-8.3/8.4); `contentDoc` still
  carries `0` for kinds where the field does not apply (A2-4.6). Prefer the
  common `evidence_ids` field; create an `evidence_ref` node only when the
  artifact itself needs edges or independent provenance (DESIGN §2 simplicity).
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
  confidence of A2-5.6 and appears exactly once per provenance entry on every
  node and edge. Q2's "Finding carries confidence" is discharged by the
  mandatory provenance block — a second confidence field would let a worker
  assert confidence in its own claim without an evidence grade. This is a
  **change to a locked PO decision (Q2) and requires the product owner's
  signature, not confirmation** (§6.1, §6 item 13). With A2-4.7's provenance
  set, `verified` is reachable: it requires a second, independent observation
  (A2-5.6).
- **A2-2.8** Who may set what:

  | Field class | Set by | Changed by |
  |---|---|---|
  | `label`, `summary`, kind fields, `attrs`, `evidence_ids` | platform ingest, from an A1 event (Q6) | never — revision only (A2-4) |
  | `severity`, `status`, `claim`, `basis` | platform ingest; the *value* originates in a worker task result or an orchestrator report, the *write* is platform-side (Q6) | never — revision only; an operator correction is a new node with `principal_kind: user` |
  | `quarantined`, `quarantine_reason` | platform policy engine (A2-8.2) | **tightening only** by an admin or the assigned operator (A2-8.1); a release happens solely as the recomputation an operator scope change causes (A2-8.5) — there is no release operation |
  | `report_excluded` | operator only, and only on a node with `quarantined:true` (A2-8.7) | operator only, event-logged |
  | `provenance`, `graph_seq`, `content_hash`, ids | platform only | never (a dedup collapse appends a provenance entry, A2-4.7) |

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
  existing `ge_` id, store nothing new, and report success. The collapse still
  emits the A1 `graph_edge_written` event with `dedup_hit:true` (A1-3.3), so
  the chain distinguishes "new evidence recorded" from "duplicate absorbed".
  _A re-observation
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
  MUST NOT create the inverse. For `contradicts` the platform normalizes the
  stored direction to `source_id < target_id` byte-wise (A0-1.9) at
  composition, which makes A2-3.5's uniqueness constraint on
  `(engagement_id, kind, source_id, target_id)` sufficient and makes a replayed
  write converge; endpoint roles carry no meaning for this kind. The
  normalized direction is immutable afterwards.
  `Tests: TestContradictsDirectionNormalized`.
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
  returns the whole chain ordered by `graph_seq` (A2-1.4) and paginates per
  A0-4. A
  single chain walk is bounded by `MaxSupersedeChain` (64) revisions; past
  that the read paginates rather than growing (adversarial A14: an agent could
  otherwise build a chain whose walk is unbounded). A superseded node MUST NOT
  be silently dropped from a read that asked for history — it is evidence.
  A `supersedes` write whose target chain already holds `MaxSupersedeChain`
  (64) revisions is `conflict` (A0-3.1) naming the bound and the chain's first
  `gn_`; the bound is enforced at **write** time, and A2-4.5's read bound is
  the consequence. `Tests: TestSupersedeChainBoundAtWrite`.
- **A2-4.6** `content_hash` is SHA-256 (A0-2.15) over the **canonical JSON**
  (A0-2) of a purpose-built content document with a **fixed key set**
  (A0-2.14): `addresses`, `attrs`, `basis`, `cidr`, `claim`,
  `credential_kind`, `domain`, `evidence_id`, `evidence_ids`, `kind`, `label`,
  `media_kind`, `port`, `protocol`, `severity`, `sid`, `size_bytes`, `status`,
  `summary`, `transport` — every field on every instance, zero-valued when not
  applicable to the kind (A0-2.14), keys in UTF-8 byte order (A0-2.4),
  integers only (A0-2.6). Its A0-2.12 **exclusion list is empty**: ids,
  `graph_seq`,
  provenance, quarantine flags, `report_excluded`, `superseded_by_id` and
  `content_hash` are not part of the document at all, so none of them can
  influence the digest. `addresses` and `evidence_ids` MUST be sorted ascending
  by unsigned byte value and deduplicated **before** the content document is
  canonicalized (A1-4.7's rule), so two observations of the same content in a
  different input order produce the same `content_hash`. `attrs` keys are
  ordered by A0-2.4. `Tests: TestContentHashStableAcrossArrayOrder` (§4.2 row
  F1-R is the vector that proves it).

  The fingerprint MUST be computed from `contentDoc` **only**; `Node` MUST NOT
  be passed to `cjson` for fingerprinting (`Node` carries `omitempty` tags and
  A0-8.3 absence semantics, `contentDoc` carries the fixed 20-key set).
  `contentDoc.Attrs`, `.EvidenceIDs` and `.Addresses` MUST be non-nil before
  marshaling (A0-2.14): a canonical content document containing `null` is a
  platform defect → `internal`. `Tests: TestContentDocFixedKeySet` (reflection:
  the canonical bytes of a zero-valued `contentDoc` contain exactly the 20 keys
  of A2-4.6 in byte order), `TestContentHashVector`.
  _Consequence of A0-2.14: these content fields are the
  one place in A2 where "unset" serializes as a zero value rather than as
  absence (A0-8.3); the API representation of a node keeps A0-8.3 absence
  semantics and the digest is computed from the dedicated document type._
- **A2-4.7** Node dedup: a write whose `(engagement_id, kind, content_hash)`
  already exists MUST return the existing `gn_` id and store no new row
  (A2-3.4 rationale). Replayed ingest of the same A1 event is therefore
  idempotent (A0-3.11, ADR-0013 offline buffering). Content dedup is the
  **only** protection against a rewound ingest watermark (A1-7.7 item 4);
  `content_hash` MUST therefore be stable across platform releases (A2-4.8,
  A0-2.16).

  `provenance` is a bounded list of at most 8 entries, each carrying its own
  `event_id`, `principal_kind`, `run_id`, `job_id`, `task_id`, `agent_node_id`,
  `tool_id`, `tool_version`, `recorded_at`, `observed_claimed_at` and `confidence`.
  The list is ordered by the `seq` of `provenance[].event_id`, never by arrival, so a
  node's bytes are deterministic. A dedup collapse (A2-4.7) appends the new
  observation's entry instead of discarding it — except that an entry whose `event_id`
  is already present is **not** appended, which is what keeps a replayed A1 event
  idempotent (ADR-0013 offline buffering). The node's `confidence` is the highest
  grade in the list and MUST be raised to `verified` only when the new entry's
  `event_id` differs from every existing one **and** its `task_id`/`agent_node_id`
  differ: a second, independent observation (A2-5.6). That is the only path by which
  `verified` is stored. Every collapse emits the A1 `graph_node_written` event with
  `dedup_hit:true`.
  Tests: TestNodeDedupKeepsEveryObservation, TestVerifiedRequiresIndependentObservation,
  TestReplayedEventAddsNoProvenanceEntry, TestProvenanceListIsOrderedByEventSeq.

  A list that already holds 8 entries is closed: a further distinct observation
  is recorded by the A1 `graph_node_written` event and MUST NOT extend the
  list; the collapse still returns the existing `gn_` id. `content_hash` is
  computed over content only (A2-4.6), so the collapse key
  `(engagement_id, kind, content_hash)` is unchanged by any provenance growth.
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

- **A2-5.1** Every node and every edge MUST carry `provenance` — a bounded
  list of at most 8 entries, ordered by the `seq` of `provenance[].event_id`
  (ADR-0016 §1: "graph content is evidence, not opinion"; A2-4.7). A record
  without provenance MUST NOT be stored.
- **A2-5.2** Provenance is **platform-stamped at ingest** and MUST NOT be
  client-supplied (Q6 worker addendum). The write request types contain no
  provenance field, so a body that carries one is rejected as an unknown field
  → `validation` (A0-6.2). _A worker that could stamp its own provenance could
  attribute a fabricated finding to a tool run that never happened._
- **A2-5.3** Fields of **one** provenance entry (A2-4.7). Absent optional
  fields are omitted, never `null`
  (A0-8.3); every id is validated per A0-1.5.

  | Field | Type | Req | Meaning |
  |---|---|---|---|
  | `principal_kind` | enum | yes | `platform`, `orchestrator`, `worker`, `node`, `user` — A1-2.1's list verbatim (A0-8.5) |
  | `run_id` | `run_` | yes | the run whose work produced this record |
  | `job_id` | `job_` | when `principal_kind` ∈ {`orchestrator`, `worker`} | orchestrator container |
  | `task_id` | `task_` | when `principal_kind` = `worker` | worker container |
  | `agent_node_id` | `slp_node_` | when observed via a remote agent (Q9) | the remote agent node — **not** `node_id` (A2-1.6, A0-3.6) |
  | `user_id` | `usr_` (A0-1.2, AM-1) | when `principal_kind` = `user` | human actor (§6.2) |
  | `tool_id` | `tool_` | when a registry tool produced the observation | ADR-0008, Q14, A0-1.3 |
  | `tool_version` | string ≤ `ToolVersionMaxBytes` (64, A0-7.1) | with `tool_id` | the registry version string; MUST NOT be folded into `tool_id` (A0-1.3) |
  | `event_id` | `evt_` | yes | the originating A1 event (A2-5.4) |
  | `recorded_at` | timestamp | yes | platform ingest time, A0-5.1/5.4 — the entry's creation timestamp |
  | `observed_claimed_at` | timestamp | no | untrusted client-supplied observation time carried by the A1 event; suffix per A0-5.7/A0-8.2 |
  | `confidence` | enum | yes | A2-5.6 |

  **`node` is required** — a Q9/ADR-0013 remote-agent observation must be
  attributable; without it such an observation has no principal.

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

  `principal_kind` is the principal whose **work produced the content**;
  `provenance.event_id`'s event `actor` is the platform subsystem that wrote
  the row (A1-3.3, always `(platform, "graph")`). The two are expected to
  differ, and copying one into the other is a defect.
  `Tests: TestPrincipalKindIsNotCopiedFromActor, TestFieldNameMappingIsTotal`.

  **AM-1 — resolved by default for the Freeze (PO confirm):** A0-1.2 registers the human-principal prefix `usr_` (`^usr_B{26}$`, 30 B) and A0 §4 adds `KindUser`. A0 owns id *shapes*; delegating the spelling to A5 would split A0-1.5 validation across two contracts. Every user-composed A1 kind (`actor.principal_id`, A1-2.2) and every A2 operator write (`user_id`, A2-5.3) validates against it. The product owner MUST confirm the prefix spelling before Frozen; it is additive-only afterwards (A0-1.10).
- **A2-5.4** Tie into A1: `event_id` MUST reference an event that exists **in
  this engagement** (A2-11). A2 consumes the A1 envelope fields `event_id`,
  `kind`, `recorded_at`, `seq` (the A1 **event** seq, not A2-1.4a's
  `graph_seq`), `engagement_id`, `run_id`, `job_id` (envelope shape per
  **A1-1.1**; A0-2.17 V5 is a canonicalization vector with a synthetic key set,
  not a valid event — see A0-2.17's annotation; A0-3.6 for the field
  spellings). Every graph record is thereby anchored to a hash-chained,
  append-only audit row (Q11, ADR-0009 §1): deleting or editing graph content
  cannot remove the evidence of its creation.
- **A2-5.5** `observed_claimed_at` MUST NOT drive ordering, supersession,
  quarantine, expiry or any digest (A0-5.7); ordering uses `graph_seq`
  (A2-1.4) and `recorded_at`. It is stored next to `recorded_at` so divergence
  is visible in the audit trail.
- **A2-5.6** `confidence` (closed, A0-8.5) grades the **evidence**, not the
  author's optimism: `observed` (directly present in captured tool output
  referenced by `event_id`/`evidence_ids`) · `inferred` (derived by reasoning
  from other graph content) · `verified` (reproduced by a second, independent
  observation). Every provenance entry carries its own grade (A2-5.3); the
  node's or edge's `confidence` is the **highest** grade in the list. An
  entry's grade is set by the platform at ingest from the ingest rule and MUST
  NOT be raised by a later write; content change is a revision (A2-4).
  `verified` is stored **only** through A2-4.7's independence rule — a second
  entry whose `event_id` differs from every existing one **and** whose
  `task_id`/`agent_node_id` differ — so no single observation can assert it.
  `Tests: TestVerifiedRequiresIndependentObservation,
  TestSelfObservedNodeStaysInferred`. **PO signature required** (§6.1, §6 item
  13: the grade scale deviates from Q2's literal wording; alternative scale
  `low`/`medium`/`high`).
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
  `validation` (A2-1.7, A0-2.6). The wire and canonical form of `attrs` is
  `{"<key>": <JSON string | integer | boolean>}` — depth exactly one, no
  wrapper object. `AttrValue` MUST carry `json:"-"` on its Go fields and a
  hand-written `MarshalJSON`/`UnmarshalJSON` emitting the bare scalar of the
  live field; the `Type` discriminator exists only in Go and is never
  serialized. `UnmarshalJSON` accepts a JSON string, integer or boolean only
  and rejects `null`, floats, arrays and objects with `validation` (A2-6.1).
  Round-trip MUST preserve `Type`.
  `Tests: TestAttrValueRoundTrip, TestAttrsRejectNestedFloatNull`.
- **A2-6.2** Keys MUST match A0-8.1 (`^[a-z][a-z0-9_]{0,39}$`), which is also
  the key length cap: ≤ 40 chars (`AttrKeyMaxBytes`, A0-7.1 registry — the same
  bound A0-8.1's regex expresses). Values are `string`
  ≤ 512 B, `int64` within A0-2.6, or `bool`.
- **A2-6.3** A key MUST NOT equal a reserved field name → `validation`. The
  reserved set is a normative, closed, additive-only (A0-6.5) constant:

  ```go
  // ReservedAttrKeys (A2-6.3): closed list. Membership is byte-exact — no
  // prefix, suffix or substring matching. `seq` stays reserved even though the
  // graph field is `graph_seq` (A2-1.4a); `operator_id` is gone with the
  // A2-5.3 rename to `user_id`.
  var ReservedAttrKeys = map[string]bool{ /* 50 keys */ }
  ```

  `ReservedAttrKeys = {id, engagement_id, seq, graph_seq, kind, label, summary,
  attrs, evidence_id, evidence_ids, addresses, cidr, port, transport,
  protocol, sid, domain, credential_kind, media_kind, size_bytes, severity,
  claim, basis, status, content_hash, quarantined, quarantine_reason,
  report_excluded, supersedes_id, superseded_by_id, provenance, source_id,
  target_id, source_kind, target_kind, retracted, principal_kind, run_id,
  job_id, task_id, agent_node_id, user_id, tool_id, tool_version, event_id,
  recorded_at, observed_claimed_at, confidence, graph_node_id,
  graph_edge_id}` (50 keys). Membership MUST be tested **byte-exactly** — a
  prefix or substring match MUST NOT be used. _Two places to look for one fact
  is how a report ends up contradicting the graph._
  `Tests: TestReservedAttrKeysRejected`.
- **A2-6.4** Caps (mechanism R, A2-7): ≤ 16 keys (`AttrsMaxKeys`), ≤ 512 B per
  string value (`AttrValueMaxBytes`), ≤ 4096 B for the serialized `attrs`
  object (`AttrsTotalMaxBytes`). Exceeding any of them → `summary_too_large`
  (A0-7.6, 413) naming the key, the cap and the actual byte count.
  Document-class caps in A2 are measured on the **A0-2 canonical form** of the
  field's own object — the same bytes that enter `content_hash` (A2-4.6) —
  computed once at ingest and stored alongside it; the API representation is
  never a measurement input, so the cap check and the fingerprint can never
  disagree.
- **A2-6.5** `attrs` is part of `content_hash` (A2-4.6), so an attrs-only
  change is a revision (A2-4), not an update.
- **A2-6.6** Promotion review: an `attrs` key that recurs is a candidate for
  the kind schema. The distinct `attrs` keys and their per-kind occurrence
  counts MUST be **derivable from stored rows** (`attrs` stored per key, never
  as an opaque blob), ordered by key byte-wise (A0-1.9). The operator-facing
  surface is the UI's (ADR-0011), not a `/api/v1` query endpoint (Q1); A4
  decides whether it exists in v1. A key observed on **≥ 3** nodes of one kind
  in one
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

  | Field class | Cap | Constant (A0-7.1 registry unless marked) | Source | Mech |
  |---|---|---|---|---|
  | `summary` (all kinds except `finding`) | 512 B | `NodeSummaryMaxBytes` | Q4 / A0-7.1 | **R** |
  | `summary` (`finding`) | 2048 B | `FindingSummaryMaxBytes` | Q4 / A0-7.1 | **R** |
  | `label` | 128 B | `NodeLabelMaxBytes` (A2-local) | A2 | **R** |
  | `claim` (`hypothesis`) | 512 B | `HypothesisClaimMaxBytes` (A2-local) | A2 | **R** |
  | `basis` (`hypothesis`) | 1024 B | `HypothesisBasisMaxBytes` (A2-local) | A2 | **R** |
  | `attrs` string value | 512 B | `AttrValueMaxBytes` | A2 / A0-7.1 | **R** |
  | `attrs` key length | 40 chars | `AttrKeyMaxBytes` | A0-8.1 / A0-7.1 | **R** |
  | `attrs` serialized object | 4096 B | `AttrsTotalMaxBytes` | A2 / A0-7.1 | **R** |
  | `attrs` key count | 16 | `AttrsMaxKeys` | A2 / A0-7.1 | **R** |
  | `evidence_ids` count | 8 | `EvidenceRefsMax` (was `EvidenceIDsMax`) | A1-4.7 / A0-7.1 | **R** |
  | `addresses` count / entry | 16 / 64 B | `AddressesMax`, `AddressMaxBytes` (the latter A2-local) | A2 / A0-7.1 | **R** |
  | `tool_version` (provenance) | 64 B | `ToolVersionMaxBytes` — **A2's former 32 was a defect; the registry value 64 governs** | A0-7.1 | **R** |
  | `provenance` entries | 8 | `ProvenanceMaxEntries` (A2-local, A2-4.7) | A2 | **R** |
  | supersede chain length | 64 | `MaxSupersedeChain` | A2 / A0-7.1 | **R** |
  | `protocol`, `sid`, `domain`, `credential_kind`-adjacent strings | 32 / 64 / 253 B | A2-local per A2-2.3 | A2 | **R** |
  | stage view document, node count, stage summary | 64 KiB / 500 / 2 KiB | `StageView*`, `StageSummaryMaxBytes` | Q4 / A0-7.1 | not assigned here — A3 per A0-7.7 |

  Every constant named without "(A2-local)" is a row of the A0-7.1 registry
  (AM-2 granted): **A2 declares no value of its own for it and cites the A0
  name** (PAIR-N1); the numbers in the Cap column are restated for readability
  only, and A0-7.1 is the source of truth. `EvidenceIDsMax` is gone — both
  documents cite `EvidenceRefsMax`. A2-local constants are new field classes
  under A0-7.7's delegation, not changes to the Q4 constants (A0-7.2).
  Measurement per A0-7.3 (decoded UTF-8 bytes of the value; serialized bytes
  for a document), truncation never applied by A2 (A0-7.4 is A3's concern),
  and document-class caps measured on the A0-2 canonical form (A2-6.4).
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

  Quarantine vocabulary: A2 owns the **state** vocabulary (`quarantine_reason`:
  `out_of_scope`, `blacklisted`); A1 owns the **occurrence** vocabulary
  (`quarantine_kind`: `out_of_scope_discovery`, `blacklist_match`,
  `operator_quarantine`). Mapping — `out_of_scope_discovery → out_of_scope` ·
  `blacklist_match → blacklisted` · `operator_quarantine → (the reason already in
  force)`. There is no `operator_release` value: a release is
  `quarantine_recomputed{scope_changed}` plus the per-node recomputation, stored as
  `quarantined:false` with `quarantine_reason` absent. The stored reason MUST be
  derived by the platform from this mapping, never copied from an event string.
  (**PO confirm**, §6 item 14 — byte-identical to A1 §6's ruling.)

  An admin or the assigned operator MAY **tighten** quarantine (`SetQuarantine`
  with `quarantined:true`, preserving the reason in force) and MUST be
  accompanied by `graph_node_quarantined{operator_quarantine}`. There is **no
  release operation**: an `out_of_scope` node is released only by an operator
  scope change and the recomputation it causes (A2-8.5, ADR-0016 §2);
  releasing a `blacklisted` node is `conflict` and MUST NOT be offered
  (A2-8.5). A release is stored as `quarantined:false` with
  `quarantine_reason` absent; the A1 event is the only record of the previous
  state.
- **A2-8.2** Quarantine is **derived, never asserted**: the platform core
  policy engine (ADR-0005 §5, SPEC §6) evaluates the node's target identity
  against the engagement
  allowlist and the global/per-engagement blacklist at ingest. A caller MUST
  NOT set `quarantined` or `quarantine_reason` — the write types have no such
  fields (A0-6.2 → `validation`). Blacklist beats allowlist beats approval
  (ADR-0005 §3, SPEC §6).

  The **per-kind evaluated field set** is closed — the policy engine MUST NOT
  match any other field: `host` → `label` + `addresses` · `network` → `label` +
  `cidr` + `addresses` · `identity`/`group` → `label` + `domain` + `sid` ·
  `credential` → `label` + `domain` · `share` → `label` + `domain` · `service`
  → `label` + `protocol` · `evidence_ref`/`finding`/`hypothesis` → derived
  (below). `attrs`, `summary`, `claim` and `basis` MUST NOT be matched — they
  are prose (A2-6.7), and matching prose lets a worker steer quarantine with a
  sentence.

  **Derivation.** A node whose kind has no identity field (`finding`,
  `hypothesis`, `evidence_ref`) is quarantined by derivation from its edges —
  if any non-retracted edge connects it to a quarantined node it is quarantined
  with the same reason; and quarantine propagates **one hop** along
  `reachable`, `authenticates_to`, `grants_access` from a quarantined
  `host`/`network` to the attached `service`/`share`. Neither rule is
  transitive beyond what is stated.

  **Recomputation triggers** (closed): on a scope/blacklist change (A2-8.5),
  on a revision that changes any identity field, and on a new node or a new
  edge touching a quarantined node. A1's `quarantine_recomputed.trigger` is
  the closed list `scope_changed · blacklist_changed · node_written ·
  edge_written` (A1-3.3) — the last two are the recomputations a new node or a
  new edge touching a quarantined node causes.

  `internal/graph` declares and consumes exactly one interface:
  `type QuarantineDecider interface { Classify(ctx context.Context,
  engagementID string, n NodeDraft) (QuarantineState, error) }`.
  `internal/policy` provides the implementation; `internal/graph` never imports
  `internal/policy` (DESIGN §1 layering, DESIGN §4: interfaces are defined at
  the consumer).
  `Tests: TestPerKindMatchedFields, TestOneHopPropagation,
  TestIdentitylessNodeQuarantinedByDerivation,
  TestQuarantineRecomputedOnEveryTrigger`.
- **A2-8.3** A quarantined node is **recorded, never actionable** (ADR-0016 §2,
  SPEC §6): it MUST NOT be returned as a candidate target by any
  planning-facing read (A2-12.4), MUST NOT be accepted as the target of a
  spawn or action even with a valid approval (A0-3.1 `approval_required` is not
  reachable for it — the policy check runs first), and MUST NOT be used to
  derive scope. Recording a discovery is itself sensitive (ADR-0016
  Consequences), so A2-9's minimization rules apply to quarantined nodes too.

  The target of a spawn or action is **never** taken from a graph field. The
  scope engine resolves the target itself (A7 action spec); the graph node a
  request cites MUST be named by `gn_` id so the quarantine check is on the id,
  not on a worker-supplied string. A request citing a quarantined node id is
  refused **before** approval routing with
  `action_blocked{reason:"target_quarantined"}` (A1-3.3).
  `Tests: TestQuarantinedNodeNotTargetableWithValidApproval,
  TestTargetResolvedByIDNotByString`.
- **A2-8.4** Quarantine propagates to edges by derivation, not by flag: an edge
  with a quarantined endpoint MUST NOT be returned by planning-facing reads
  (A2-3.9).
- **A2-8.5** `quarantine_reason: blacklisted` is **not releasable**: it MUST
  survive any scope change and any approval, because the blacklist cannot be
  overridden by scopes or approvals (ADR-0005 §3). `out_of_scope` changes only
  when an operator changes the engagement scope (Q6: machine principals never
  mutate scope/blacklist); the platform MUST recompute quarantine on such a
  change and MUST record the recomputation as an A1 event.

  A blacklisted discovery is **recorded, not refused**: it is stored with
  `quarantine_reason:"blacklisted"`, permanently non-releasable, absent from
  planning views (A2-12.5), reported as "not tested" (A2-8.6), and the write
  MUST emit `graph_node_quarantined{blacklist_match}` (A2-8.10, A2-8.1's
  mapping). **PO confirm** (§6 item 12: the alternative reading of ADR-0016 §2
  is to refuse the write and store nothing).
  `Tests: TestBlacklistedNodeSurvivesScopeWidening,
  TestBlacklistedDiscoveryIsRecordedNotRefused`.
- **A2-8.6** Reporting handover (Q5): quarantined nodes MUST be included in the
  reporting handover and marked *not tested*. The machine-readable marker is
  `quarantined: true` plus `quarantine_reason`; the words "not tested" are
  report prose (A0-3.4: prose is never parsed). A report MUST NOT render a
  quarantined node as tested, attempted or actionable, and MUST NOT render it
  inside an attack path.
- **A2-8.7** An operator MAY exclude a quarantined node from the report by
  setting `report_excluded: true`, which MUST be accompanied by an A1 event
  naming the operator, the `gn_` id and the reason text. `report_excluded`
  MUST be settable to `true` **only** on a node with `quarantined:true` (Q5
  authorizes removal of quarantined discoveries from the report only). An
  attempt on a non-quarantined node is `conflict` (A0-3.1, immutable state) and
  MUST be recorded as `action_blocked{reason:"graph_write_rejected",
  action_kind:"graph_write"}` (A1-3.3). Removing a confirmed finding from a
  report is expressed by a revision (A2-4) to `status:"refuted"` or
  `severity:"info"`, never by a flag. Exclusion-by-flag over
  deletion: the graph stays intact, the exclusion is auditable and reversible,
  and no `supersedes` chain or evidence reference breaks (ADR-0016 §4). A
  delete endpoint for graph content MUST NOT exist in v1.
  `Tests: TestReportExcludedRequiresQuarantine,
  TestMachinePrincipalCannotSetReportExcluded`. **PO confirm** (§6.4).
- **A2-8.8** `report_excluded` MUST NOT be settable by a machine principal
  (Q6) and MUST NOT alter planning behaviour — it is a reporting filter only.
  `Tests: TestQuarantineFlagCannotBeSuppliedOnWrite,
  TestOperatorReportExclusionIsEventLogged`.
- **A2-8.9** Why mutability does not break ordering (A0-4.3): `quarantined`,
  `quarantine_reason`, `report_excluded` and `retracted` MUST NOT participate
  in any collection order, cursor or `graph_seq` (A2-1.4, A2-1.4a). Every A2
  collection orders by `(graph_seq, graph_node_id)` for nodes and
  `(graph_seq, graph_edge_id)` for edges — all four immutable — so an operator
  flipping a flag mid-paging
  can change *which* rows match a filter but can never skip or duplicate a row
  in an in-flight page walk. Clients MUST still deduplicate by id and MUST NOT
  assume a paged set is a snapshot (A0-4.7: mutable rows, no isolation knob).
  `Tests: TestQuarantineFlagFlipDoesNotSkipRows`.
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

  The pattern set is A2-9.4's rule table and nothing else: rule ids live in exactly
  one document, and an error MUST name the field and the rule id, never the value, a
  prefix of it, or a digest of it (A0-3.4, A2-9.5). The scan is a **filter, not a
  guarantee** — a hostile worker can encode, split or re-format a secret past any
  pattern set. Egress exclusion (ADR-0020 §4) is the enforcement point and MUST be
  applied independently at the gateway to every string that leaves the platform; for an
  engagement whose policy is not `local_only` the gateway MUST exclude by **kind** (no
  `credential` node, no node with `credential_kind` set, no `attrs` of such a node) and
  `llm_call` MUST record the exclusion in `excluded_secret_count`.
- **A2-9.4** Ingest scanning: the platform MUST run the `internal/secretscan`
  stdlib pattern scan (`regexp`) over every string field and every `attrs`
  value of a node or edge, and MUST reject a match with `validation` naming the
  **field** and the **rule id** — never echoing the value (A0-3.4 permits
  echoing untrusted material; A2-9.5 overrides that here). The rule set below
  is **normative, closed and additive-only** (A0-6.5): an implementation MUST
  run every rule, and a rule id not in this table MUST NOT be reported. Reject,
  never redact (§6 item 10, ruled once for A1 §6.4 and A2 §6.10).

  | rule id | Go regexp | fields scanned | note |
  |---|---|---|---|
  | `SEC-PEM` | ``-----BEGIN [A-Z ]*PRIVATE KEY-----`` | all string fields, every `attrs` value | PEM private-key block header; the block body is never accepted either |
  | `SEC-NTLM` | ``(?i)\b[0-9a-f]{32}\b`` | all string fields, every `attrs` value | 32-hex NTLM/LM hash shape. Known false-positive class: a bare 32-hex token in prose (reject anyway, §6 item 10) |
  | `SEC-KRB` | ``(?i)krbtgt[/@][A-Za-z0-9._-]{1,128}`` | all string fields, every `attrs` value | `krbtgt` ticket material / TGT principal form |
  | `SEC-AWSKEY` | ``(AKIA\|ASIA)[0-9A-Z]{16}`` | all string fields, every `attrs` value | AWS access-key id (long-term and temporary) |
  | `SEC-GCPKEY` | ``AIza[0-9A-Za-z\-_]{35}`` | all string fields, every `attrs` value | Google API / service-account key shape; a GCP service-account *private key* is `SEC-PEM` |
  | `SEC-AZUREKEY` | ``(?i)(AccountKey\|SharedAccessKey\|sig)=[A-Za-z0-9+/=]{20,}`` | all string fields, every `attrs` value | Azure storage account key, SAS signature and connection-string shapes |
  | `SEC-JWT` | ``eyJ[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+`` | all string fields, every `attrs` value | compact JWS/JWT — the payload is attacker-readable and often holds claims |
  | `SEC-BEARER` | ``(?i)(bearer\|token\|api[_-]?key\|password\|passwd\|secret)\s*[:=]\s*\S{8,}`` | all string fields, every `attrs` value | `key: value` credential assignment in prose |
  | `SEC-URLCRED` | ``[a-z][a-z0-9+.-]*://[^/\s:@]{1,64}:[^/\s:@]{1,64}@`` | all string fields, every `attrs` value | userinfo credentials embedded in a URL |
  | `SEC-ENTROPY` | not a regexp — Shannon entropy **≥ 4.5 bits/char** over a window of **≥ 32 characters** drawn from a base64/hex alphabet (`[A-Za-z0-9+/=_-]`); the formula `-Σ p(c)·log2 p(c)` over the window and the window size are part of the rule | every string field and every `attrs` value of ≥ 32 characters | high-entropy bearer material with no keyword anchor; MUST be deterministic (`TestEntropyRuleIsDeterministic`) |

  "all string fields" means every stored string of the node or edge being
  written — `label`, `summary`, `claim`, `basis`, `protocol`, `domain`, `sid`,
  `cidr`, `transport`, every `addresses` entry, every edge field — and every
  `attrs` value. Provenance is scanned too (A2-5.8).

  The corpus planted by `TestEventSecretFreeSerialization` and
  `TestGraphSecretFreeSerialization` is exactly one value per rule id and lives
  in the shared suite; a rule added later adds a corpus entry.
  `Tests: TestEveryRuleIDMatchesItsCorpusValue, TestNoFalsePositiveOnBenignCorpus,
  TestErrorMessageNamesFieldAndRuleIDOnly, TestEntropyRuleIsDeterministic,
  TestGraphSecretFreeSerialization`.
- **A2-9.5** Secrets never appear in errors or logs (ADR-0019 §5, A0-3.7):
  an error `message`, an error `attrs` entry and any `slog` record MUST NOT
  contain a rejected secret value **in whole, in part or as a digest** — not
  the value, not a prefix of it, not its hash (A0-3.4's exception, PAIR-X1).
  The message names the field, the rule id (A2-9.4) and the byte length only.
- **A2-9.6** What a worker MUST do with a captured secret instead (Q6 worker
  addendum — the worker is report-only and has **no** graph access): upload it
  to the evidence store (`evidence:upload`), reference the returned `evi_` id in
  its task result / A1 event, and let the platform create the `credential` node
  with `evidence_id` set and a pseudonymous `label`. The worker MUST NOT put
  the secret in an event payload, a task result, a log line, a model prompt
  (ADR-0020 §4) or any graph-bound content.
- **A2-9.7** Negative tests (shared suite, `contracts/README.md` "secret-free
  serialization"). The planted corpus is exactly one value per A2-9.4 rule id
  and lives in the shared suite:
  - `TestGraphSecretFreeSerialization` — for a corpus of writes with planted
    secrets (one per A2-9.4 rule id: an NTLM-shaped hash, a PEM private key, a
    cloud access-key id, a bearer token, a URL with embedded credentials, a
    JWT, a high-entropy string) in `label`, `summary`, `attrs` values and edge
    fields: the
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
  (A2-3) · (12) provenance establishment (A2-5) · (13) fingerprint, policy and
  dedup, in this fixed order: **(13a)** compute `content_hash` over the A2-4.6
  document; **(13b)** evaluate policy and quarantine (A2-8.2,
  `QuarantineDecider`); **(13c)** **then** attempt the dedup collapse of
  A2-3.4/A2-4.7. A write MUST report exactly one error:
  the first violated rule. A collapse still emits the A1 `graph_node_written`
  (with `dedup_hit:true`) and still emits `graph_node_quarantined` when the
  recomputed quarantine state differs from the stored one — so an agent that
  re-observes a blacklisted host 500 times produces 500 chained signals, not
  one. `Tests: TestValidationOrderIsDeterministic,
  TestCollapseStillEmitsQuarantineEvent`.
- **A2-10.3** Error kinds, per A0-3.1 and never invented here: schema, enum,
  id, type, not-applicable-field, endpoint-kind, attrs-shape, secret-scan and
  cycle violations → `validation` (400) · cap violations under mechanism R →
  `summary_too_large` (413, A0-7.6) · immutable-field update, supersede of a
  non-current node, supersede cycle → `conflict` (409) · endpoint or
  `event_id` that does not resolve in this engagement → `notfound` (404,
  A0-3.9) · store uniqueness race or missing platform provenance → `internal`
  (500, A0-1.4). `Tests: TestHardRejectMatrix` — one subtest per rejection
  class of §4.3, so this table is exhaustively asserted and a new rejection
  class needs a new subtest.
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
  constructors (`graph.NewNode`, `graph.NewEdge`), and it accepts a
  `PendingNode`/`PendingEdge` and nothing else (§4). `Node` and `Edge` fields
  stay **exported** with the sketched `json` tags — `encoding/json` cannot
  marshal unexported fields, and no `go vet` exported-field audit exists. The
  no-bypass guarantee is instead: (a) `NewNode`/`NewEdge` are the only
  documented construction path; (b) the store seam re-validates every value it
  is given (A2-10.7); (c) `TestNodeHasNoExportedContentSetter` — reflection
  over `graph.Node`/`graph.Edge` asserts no exported method mutates a content
  field (the four mutation methods below and the flags of A2-1.3/A2-3.9 are the
  declared exceptions); (d) review per AGENTS.md. If the product owner prefers
  unexported fields, A2 MUST specify `MarshalJSON`/`UnmarshalJSON` for `Node`
  and `Edge`.

  The graph store seam exposes exactly four mutation methods, each taking an
  engagement id and each returning `notfound`/`conflict` per A2-10.3 —
  `SetSupersededBy`, `SetQuarantine`, `SetReportExcluded`,
  `SetEdgeRetracted`. No other update or delete method MAY exist (A1-7.2's rule
  applied to the graph seam).
  `Tests: TestNodeHasNoExportedContentSetter,
  TestGraphSeamHasExactlyFourMutationMethods, TestNodeDraftRejectsPlatformFields`.
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
  no unfiltered or multi-engagement read method MAY exist. The seam exposes
  exactly the four mutation methods of A2-10.6, each taking an engagement id.
  _A method that can
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
    yields `notfound` (404) with an envelope whose `attrs.graph_node_id` **is
    the requested id** (ids are not secrets, A0-1.7) and whose `message` does
    not distinguish "absent" from "elsewhere" (A0-3.9).
  - `TestNoCrossEngagementEdge` — an edge whose endpoints lie in A and B is
    rejected `notfound`, and no edge row is created in either engagement.
  - `TestCursorFromEngagementARejectedInB` — replaying A's cursor (A0-4.4) on a
    B-scoped list yields **`validation` (400)**, never A data and never a
    silently empty page (A1-8.2 is the same oracle, PAIR-K1); authorization is
    re-derived per page, never from the cursor.
  - `TestNoBulkReadSpansEngagements` — reflection/endpoint audit over the A4
    route table and the store seam: no graph read accepts more than one
    engagement id or omits it.
  - `TestNodeDedupDoesNotSpanEngagements` — A2-4.7's collapse key is
    `(engagement_id, kind, content_hash)`: a node written in A whose content
    is byte-identical to one in B creates a **new** `gn_` in B, and neither
    engagement's `content_hash` is ever looked up in the other's rows.
  - `TestEdgeDedupDoesNotSpanEngagements` — the same for A2-3.5's edge key
    `(engagement_id, kind, source_id, target_id)`: no collapse across
    engagements, and no `ge_` id from A is ever returned by a B-scoped write.
  - `TestNoBareNodeIDInGraphDocuments` — no graph request, response, error
    body or log attribute contains a bare `node_id` key for a graph node
    (A2-1.6, A0-3.6); the only `node_id` in the platform is the remote agent
    node (`slp_node_`, Q9), spelled `agent_node_id` in A2 (A2-5.3).
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
  | node reference shape `{graph_node_id, kind, label, quarantined}` — every node has all four; `label` ≤ 128 B, `kind` ≤ 32 B | A2-1.6, A2-2.2, A2-7.1 |
  | full `summary` ≤ 512 B (≤ 2048 B for `finding`), available on a single-node drill-down read | A0-7.1, A2-7.1 |
  | current-ness is A2's: a view that filters to current nodes uses `superseded_by_id` absence; A3 MUST NOT walk `supersedes` chains itself | A2-1.8, A2-4.4 |
  | deterministic ordering inputs: immutable `graph_seq`, then `graph_node_id`/`graph_edge_id` byte-wise (A0-1.9); no mutable field orders anything | A2-1.4, A2-1.4a, A2-8.9 |
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
  `Tests: TestQuarantinedNodeAbsentFromPlanningView,
  TestRetractedEdgeAbsentFromPlanningView`.

## 4. Types

Illustrative sketches — **not compiled** (`contracts/README.md`). They are the
source of truth for field names and JSON shapes until `internal/graph` merges
(DESIGN §1). **Domain layer** (DESIGN §1); imports foundation only:
`internal/ids`, `internal/cjson`, `internal/errs`. One flat node type per
A2-2.3 rather than a per-kind payload
interface: no hand-rolled JSON type dispatch, one fixed key set for A0-2.14,
one validator `switch` (DESIGN §2, ADR-0001/0010 stdlib-only).

```go
// internal/graph — the A2 contract types. Domain layer: imports internal/ids,
// internal/cjson, internal/errs only (DESIGN §1). No pgx here; the store seam is
// internal/store (ADR-0010). internal/graph never imports internal/policy (A2-8.2).
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

type PrincipalKind string // A2-5.3: A1-2.1's list verbatim (PAIR-A1).

const (
	PrincipalPlatform     PrincipalKind = "platform"
	PrincipalOrchestrator PrincipalKind = "orchestrator"
	PrincipalWorker       PrincipalKind = "worker"
	PrincipalNode         PrincipalKind = "node" // remote agent (Q9): required, A2-5.3
	PrincipalUser         PrincipalKind = "user" // was "operator"; prose "operator" unchanged
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

// AttrValue serializes as the BARE scalar of its live field (A2-6.1): the wire
// and canonical form of attrs is {"<key>": <string | integer | boolean>}, depth
// exactly one, no wrapper object. Type is a Go-only discriminator and is never
// serialized, hence json:"-" on every field below and hand-written methods.
type AttrValue struct {
	Type AttrType `json:"-"` // which of Str/Num/Bool is live
	Str  string   `json:"-"` // ≤ AttrValueMaxBytes (A2-7.1)
	Num  int64    `json:"-"` // within A0-2.6
	Bool bool     `json:"-"`
}

// MarshalJSON emits the bare scalar of the live field; UnmarshalJSON accepts a
// JSON string, integer or boolean only and rejects null, floats, arrays and
// objects with errs kind "validation" (A2-6.1). Round-trip preserves Type.
func (v AttrValue) MarshalJSON() ([]byte, error)
func (v *AttrValue) UnmarshalJSON(b []byte) error

// Cap constants: A2 declares none of the A0-7.1 registry values (PAIR-N1) —
// ToolVersionMaxBytes (64), EvidenceRefsMax (8), MaxSupersedeChain, AttrsMaxKeys,
// AttrKeyMaxBytes, AttrValueMaxBytes, AttrsTotalMaxBytes, AddressesMax,
// NodeSummaryMaxBytes and FindingSummaryMaxBytes come from internal/caps (A0-7.1).
// A2's former `ToolVersionMaxBytes = 32` and `EvidenceIDsMax` are deleted.
// Only these field classes remain A2-local (A0-7.7 delegation, A2-7.1):
const (
	NodeLabelMaxBytes       = 128
	HypothesisClaimMaxBytes = 512
	HypothesisBasisMaxBytes = 1024
	AddressMaxBytes         = 64
	ProvenanceMaxEntries    = 8 // A2-4.7
)

// Provenance is one entry of the mandatory provenance list (A2-5.1, A2-4.7:
// ≤ 8 entries, ordered by the seq of event_id) and is platform-stamped only
// (A2-5.2). Absent optional fields are omitted, never null (A0-8.3).
type Provenance struct {
	PrincipalKind     PrincipalKind `json:"principal_kind"`
	RunID             string        `json:"run_id"`
	JobID             string        `json:"job_id,omitempty"`
	TaskID            string        `json:"task_id,omitempty"`
	AgentNodeID       string        `json:"agent_node_id,omitempty"` // slp_node_ (Q9), never "node_id" (A2-1.6)
	UserID            string        `json:"user_id,omitempty"`       // usr_ (A0-1.2, AM-1); was operator_id
	ToolID            string        `json:"tool_id,omitempty"`       // tool_ (A0-1.3)
	ToolVersion       string        `json:"tool_version,omitempty"`  // ≤ ToolVersionMaxBytes (64, A0-7.1)
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

// QuarantineDecider is the one interface internal/graph declares and consumes
// (A2-8.2, DESIGN §4: interfaces are defined at the consumer). internal/policy
// implements it; internal/graph never imports internal/policy.
type QuarantineDecider interface {
	Classify(ctx context.Context, engagementID string, n NodeDraft) (QuarantineState, error)
}

// Node is one graph node. Fields are immutable except the quarantine and
// report flags (A2-1.3, A2-8). Fields are EXPORTED with these json tags;
// NewNode is the only documented construction path and the seam re-validates
// (A2-10.6).
type Node struct {
	ID           string   `json:"graph_node_id"` // gn_ (A0-1.2, A2-1.6)
	EngagementID string   `json:"engagement_id"`
	GraphSeq     int64    `json:"graph_seq"`     // immutable order key (A2-1.4, A2-1.4a)
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
	Provenance     []Provenance     `json:"provenance"`                // ≤ 8 entries, by event seq (A2-4.7)
}

// Edge is one directed relationship (A2-3). Content is immutable; only
// Retracted may change (A2-3.9).
type Edge struct {
	ID           string    `json:"graph_edge_id"` // ge_ (A0-1.2, A2-1.6)
	EngagementID string    `json:"engagement_id"`
	GraphSeq     int64     `json:"graph_seq"`
	Kind         EdgeKind  `json:"kind"`
	SourceID     string    `json:"source_id"` // gn_
	SourceKind   NodeKind  `json:"source_kind"`
	TargetID     string    `json:"target_id"` // gn_
	TargetKind   NodeKind  `json:"target_kind"`
	Retracted    bool      `json:"retracted"`
	Provenance   []Provenance `json:"provenance"`
}

// contentDoc is the purpose-built document content_hash is computed over
// (A2-4.6). Fixed key set, zero values for inapplicable fields (A0-2.14),
// empty A0-2.12 exclusion list, canonicalized by internal/cjson (A0-2).
// Node MUST NOT be passed to cjson for fingerprinting (A2-4.6): only this type.
// Attrs, EvidenceIDs and Addresses MUST be non-nil before marshaling, and
// Addresses/EvidenceIDs are sorted ascending and deduplicated first (A2-4.6).
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

// NodeDraft is the ingest path's input: content only, no platform-set field.
type NodeDraft struct {
	Kind                     NodeKind
	Label, Summary           string
	Attrs                    Attrs
	EvidenceIDs, Addresses   []string
	CIDR                     string
	Port                     int
	Transport, Protocol      string
	SID, Domain              string
	CredentialKind           CredentialKind
	EvidenceID               string
	MediaKind                MediaKind
	SizeBytes                int64
	Severity                 Severity
	Claim, Basis             string
	Status                   string
	SupersedesID             string // A2-4.2: set by the revising write, not by the platform
}

// PendingNode is validated content + provenance + quarantine state + content_hash,
// with no id and no graph_seq (assigned at insert, A2-1.2/A2-1.4a). DESIGN §4: a
// constructor never returns a half-built value.
type PendingNode struct { /* unexported; NewNode is the only way to build one */ }

func NewNode(in NodeDraft, prov Provenance, q QuarantineState) (PendingNode, error)

// WriteNode assigns graph_node_id (A0-1.4) and graph_seq (A2-1.4a) inside the insert
// transaction, applies A2-3.4/A2-4.7 dedup, and returns the complete Node. A2-10.6:
// the seam accepts PendingNode and nothing else.
func WriteNode(ctx context.Context, engagementID string, n PendingNode) (Node, error)

// Edges follow the same shape: EdgeDraft carries Kind, SourceID, TargetID and
// nothing else; PendingEdge adds provenance; WriteEdge assigns graph_edge_id and
// graph_seq inside the insert transaction and applies A2-3.4 dedup.
type EdgeDraft struct{ Kind EdgeKind; SourceID, TargetID string }
type PendingEdge struct { /* unexported; NewEdge is the only way to build one */ }

func NewEdge(in EdgeDraft, prov Provenance) (PendingEdge, error)
func WriteEdge(ctx context.Context, engagementID string, e PendingEdge) (Edge, error)

// NewNode/NewEdge validate in A2-10.2 order; errors carry an A0-3 kind:
// validation, summary_too_large, conflict, notfound (A2-10.3). A2-10.6 applies
// to PendingNode and Node alike.
// ID, EngagementID, GraphSeq, ContentHash, Quarantined, QuarantineReason,
// ReportExcluded, SupersededByID and Provenance are platform-set and absent
// from the draft; a request body carrying one is an unknown field on a write
// (A0-6.2, A2-1.5, A2-5.2).

// Current reports whether n is the current revision (A2-1.8).
func (n Node) Current() bool { return n.SupersededByID == "" }

// The seam's only mutation methods (A2-10.6, A2-11.1): four, each taking an
// engagement id, each returning notfound/conflict per A2-10.3.
func SetSupersededBy(ctx context.Context, engagementID, nodeID, byNodeID string) error
func SetQuarantine(ctx context.Context, engagementID, nodeID string, q QuarantineState) error
func SetReportExcluded(ctx context.Context, engagementID, nodeID string, excluded bool) error
func SetEdgeRetracted(ctx context.Context, engagementID, edgeID string, retracted bool) error
```

### 4.1 JSON examples

A `finding` node with a two-entry provenance list (A2-2.3, A2-5.1, A2-5.3).
`quarantined` and `report_excluded` are present and `false`; unset optional
fields are absent (A0-8.3). `provenance` is an **array** ordered by the A1
event `seq` of its entries (A2-4.7), never by arrival. Its `content_hash` is
§4.2 vector **F1**.

```json
{
  "graph_node_id": "gn_01m1y2whfhh039ykj5x8mc5a0g",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "graph_seq": 412,
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
  "content_hash": "ad8f188e63b2563f9adad88df585d94df9c662e4082895e974b0b31e8a65076e",
  "quarantined": false,
  "report_excluded": false,
  "provenance": [
    {
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
      "confidence": "observed"
    },
    {
      "principal_kind": "worker",
      "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
      "job_id": "job_01m1y2whfhvk83rt5x1z7c4m2b",
      "task_id": "task_01m1y2whfh4kd8nq2z7x1vb3rt",
      "agent_node_id": "slp_node_01m1y2whfhq3vb8nrt5x1z7c4m",
      "tool_id": "tool_01m1y2whfhfjdvwqp9pfxqekmf",
      "tool_version": "1.4.2",
      "event_id": "evt_01m1y2whfhr8t2nb5x9qz4vb7m",
      "recorded_at": "2026-09-07T14:07:55.140Z",
      "observed_claimed_at": "2026-09-07T14:07:52.880Z",
      "confidence": "verified"
    }
  ]
}
```

_Two entries, two different `task_id`/`agent_node_id` values: the second is the
independent observation that raises the node's grade to `verified` (A2-4.7,
A2-5.6). A single-entry node can never show `verified` — the remaining
examples below each carry one entry and therefore show `observed` or
`inferred`._

A `supersedes` revision pair (A2-4): the old hypothesis is unchanged and
gains only `superseded_by_id`; the new one carries `supersedes_id`. Both stay
in the graph, both keep their own provenance and `content_hash`.

```json
{
  "items": [
    {
      "graph_node_id": "gn_01m1y2whfh9x2b4c7d1e8f0a3b",
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "graph_seq": 388,
      "kind": "hypothesis",
      "label": "dc01 accepts NTLM relay from the VPN range",
      "claim": "NTLM authentication from 10.20.0.0/24 can be relayed to SMB on dc01 because signing is not enforced.",
      "basis": "Port 445 reachable (evt_01m1y2whfhc2v9nq4x7z1m8b3p) and no SMB signing requirement observed in one negotiation.",
      "status": "supported",
      "attrs": { "negotiations_seen": 1 },
      "content_hash": "89af62666566f98a406f79541497b1a3d7116646edbb0fafe83e239dea1425d4",
      "quarantined": false,
      "report_excluded": false,
      "superseded_by_id": "gn_01m1y2whfjk5t8nq2z7x1vb3rt",
      "provenance": [
        {
          "principal_kind": "orchestrator",
          "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
          "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
          "event_id": "evt_01m1y2whfhd8k3m9qz1x7vb4nr",
          "recorded_at": "2026-09-07T13:41:07.220Z",
          "confidence": "inferred"
        }
      ]
    },
    {
      "graph_node_id": "gn_01m1y2whfjk5t8nq2z7x1vb3rt",
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "graph_seq": 405,
      "kind": "hypothesis",
      "label": "dc01 accepts NTLM relay from the VPN range",
      "claim": "NTLM relay from 10.20.0.0/24 to SMB on dc01 succeeds and reaches SYSVOL; three independent negotiations reproduced it.",
      "basis": "Reproduced by task_01m1y2whfh1txm57x8dn41r9hg against dc01 and by a second run against the file server; signing absent in all three negotiations.",
      "status": "supported",
      "attrs": { "negotiations_seen": 3 },
      "content_hash": "1c3c351c6f194a2c3f5217df33016e92b133fc2f035b4a310c8c64decfaf560d",
      "quarantined": false,
      "report_excluded": false,
      "supersedes_id": "gn_01m1y2whfh9x2b4c7d1e8f0a3b",
      "provenance": [
        {
          "principal_kind": "worker",
          "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
          "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
          "task_id": "task_01m1y2whfh1txm57x8dn41r9hg",
          "tool_id": "tool_01m1y2whfhfjdvwqp9pfxqekmf",
          "tool_version": "1.4.2",
          "event_id": "evt_01m1y2whfhm6p3kq9z2x1vb4rt",
          "recorded_at": "2026-09-07T13:58:41.330Z",
          "observed_claimed_at": "2026-09-07T13:58:39.100Z",
          "confidence": "inferred"
        }
      ]
    }
  ]
}
```

_The pair above is a hypothesis revised by a better-supported hypothesis
(same `kind`, A2-4.3); a confirmed exploitation is written as a `finding`
(first example) and attached to its target by the `exploited_by` edge below._
_The current hypothesis stays `inferred` although its `basis` prose claims
three reproductions: one provenance entry is one observation, and `verified`
is stored only through A2-4.7's independence rule. Both `content_hash` values
are computed from each node's own A2-4.6 `contentDoc` (illustrative — they are
not §4.2 vectors); they differ, which is what keeps the revision from
collapsing into the superseded node (A2-4.7)._

A quarantined out-of-scope host (A2-8), and the edge that attaches a finding
to its target (A2-3.2):

```json
{
  "graph_node_id": "gn_01m1y2whfhq7z3m9x1c4vb8nrt",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "graph_seq": 431,
  "kind": "host",
  "label": "payroll.corp.acme.com",
  "addresses": ["10.99.4.20"],
  "summary": "Seen in DNS response for an in-scope resolver. Outside the engagement allowlist and inside the global blacklist range 10.99.0.0/16: recorded, never tested.",
  "attrs": { "seen_via": "dns_zone_transfer_partial" },
  "content_hash": "e6dedac2032b119d837105f6271e233686264bd2ea7b35f50b3bd8d9a1164148",
  "quarantined": true,
  "quarantine_reason": "blacklisted",
  "report_excluded": false,
  "provenance": [
    {
      "principal_kind": "platform",
      "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
      "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
      "event_id": "evt_01m1y2whfhk4m2nq8x7z1vb3rt",
      "recorded_at": "2026-09-07T14:11:52.007Z",
      "observed_claimed_at": "2026-09-07T14:11:48.310Z",
      "confidence": "observed"
    }
  ]
}
```

_`quarantined` and `quarantine_reason` are not part of `contentDoc` (A2-4.6's
exclusion list): the host's `content_hash` is
`e6dedac2032b119d837105f6271e233686264bd2ea7b35f50b3bd8d9a1164148`
(illustrative — not a §4.2 vector), and flipping the flag later never changes
it. A blacklisted discovery is **recorded, not refused** (A2-8.5, §6 item 12)._

```json
{
  "graph_edge_id": "ge_01m1y2whfh62ej11jf4x5gjzv4",
  "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
  "graph_seq": 413,
  "kind": "exploited_by",
  "source_id": "gn_01m1y2whfhdc01srv4x8mc5a0g",
  "source_kind": "host",
  "target_id": "gn_01m1y2whfhh039ykj5x8mc5a0g",
  "target_kind": "finding",
  "retracted": false,
  "provenance": [
    {
      "principal_kind": "platform",
      "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
      "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
      "event_id": "evt_01m1y2whfhp17g0avdqztd2p3x",
      "recorded_at": "2026-09-07T14:03:22.502Z",
      "confidence": "inferred"
    }
  ]
}
```

_An edge carries no `content_hash` (A2-4.6 fingerprints nodes only; the edge
dedup key is A2-3.5's `(engagement_id, kind, source_id, target_id)`), and its
single provenance entry is `inferred`: the platform composed the edge from the
finding's own evidence (A2-5.6)._

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

### 4.2 Normative content fingerprint vector

Normative, like A0-2.17 and A1 §4.3: the shared contract-test suite MUST
reproduce these bytes and digests byte-exactly. Each row is one node's A2-4.6
`contentDoc` — the fixed 20-key set, zero values for the keys that do not
apply to the kind (A0-2.14), `addresses`/`evidence_ids` sorted ascending by
unsigned byte value and deduplicated **before** canonicalization (A2-4.6),
keys in UTF-8 byte order (A0-2.4) — and its SHA-256 (A0-2.15) is that node's
`content_hash`. A `content_hash` computed over unsorted arrays (F1-R's
forbidden digest below) is a **defect**, not a variant.

| row | node | len | SHA-256 |
|---|---|---|---|
| **F1** | the §4.1 `finding` example | **656** | `ad8f188e63b2563f9adad88df585d94df9c662e4082895e974b0b31e8a65076e` |
| **F1-R** | the same node with `evidence_ids` **given** in the reverse order | 656 | `ad8f188e…076e` — identical bytes and identical digest to F1 |
| **F3** | F1 with `attrs.cvss_v3_x10` = 87 (one `attrs` value differs) | **656** | `1748b813a0d28d9f17f4d89dff2a08536741bf45e8fe0defbb3c4363d3f9b983` |
| **S1** | a `service` node, inapplicable keys at their zero values | **304** | `3955d82160284d3e76c9be1b06981530df50e1635a1bad7ceb7b50c5f73c2b67` |

F1-R's **forbidden** variant: canonicalizing the same content without A2-4.6's
sort/dedup yields
`4a017e71658de7ebb2d3a5429f90818b8a69302ff1909badbc782d9049ce9cae`. That
digest is published as the **defective-implementation marker** — an
implementation that produces it for this input fails
`TestContentHashStableAcrossArrayOrder` and MUST NOT ship.

None of the four canonical byte strings contains an invisible code point (all
are pure ASCII, so no `<2028>`/`<2029>`/`<7F>` placeholder form is needed
here); the **len** and **SHA-256** columns are the authority, and a byte string
below that disagrees with them is a transcription defect.

```
F1 and F1-R canonical bytes (656 B, byte-identical for both rows):
{"addresses":[],"attrs":{"cvss_v3_x10":88,"first_seen_task":"task_01m1y2whfh1txm57x8dn41r9hg","relay_tool":"ntlmrelayx"},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":["evi_01m1y2whfh3ca875z2x8v8h7qt","evi_01m1y2whfh7kq2m4c8x1z9vb3n"],"kind":"finding","label":"SMB relay to SYSVOL on dc01","media_kind":"","port":0,"protocol":"","severity":"high","sid":"","size_bytes":0,"status":"confirmed","summary":"Captured NTLM authentication from 10.20.0.14 was relayed to the SYSVOL share on dc01, yielding read access to group policy preferences. Secret material is referenced, not stored (evi_).","transport":""}

F3 canonical bytes (656 B):
{"addresses":[],"attrs":{"cvss_v3_x10":87,"first_seen_task":"task_01m1y2whfh1txm57x8dn41r9hg","relay_tool":"ntlmrelayx"},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":["evi_01m1y2whfh3ca875z2x8v8h7qt","evi_01m1y2whfh7kq2m4c8x1z9vb3n"],"kind":"finding","label":"SMB relay to SYSVOL on dc01","media_kind":"","port":0,"protocol":"","severity":"high","sid":"","size_bytes":0,"status":"confirmed","summary":"Captured NTLM authentication from 10.20.0.14 was relayed to the SYSVOL share on dc01, yielding read access to group policy preferences. Secret material is referenced, not stored (evi_).","transport":""}

S1 canonical bytes (304 B):
{"addresses":["10.20.0.14"],"attrs":{},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":[],"kind":"service","label":"microsoft-ds","media_kind":"","port":445,"protocol":"smb","severity":"","sid":"","size_bytes":0,"status":"","summary":"","transport":"tcp"}
```

F1 and F3 differ in exactly one `attrs` value and therefore in nothing else —
they lock A2-6.5 (an `attrs`-only change is a revision, not an update) and
A2-4.7's dedup key at the same time. S1 locks the zero-value rule of A2-4.6:
`"attrs":{}` and `"evidence_ids":[]` are present and non-`null` (A0-2.14), and
the keys that do not apply to a `service` are present with `""`/`0`.
`Tests: TestContentHashVector, TestContentDocFixedKeySet` (A2-4.6, §4.2).

### 4.3 Contract tests (A2)

The shared contract-test suite (`contracts/README.md`) MUST implement these ids
for A2; the name is the identifier, one name per test, one oracle per name, and
the clause named is the rule the test guards. §4.2's vectors are data, not a
test id: the suite asserts their canonical bytes, lengths and digests
byte-exactly. Ids the plan assigned inside their own clause are listed in the
index at the end of this subsection (a clause that carries both kinds of id
appears in both tables, with different ids in each).

| Clause | Test ids |
|---|---|
| A2-1.3 | `TestImmutableFieldUpdateIsConflict` |
| A2-1.5 | `TestNodeReadIgnoresUnknownFields` |
| A2-3.2 | `TestEdgeEndpointMatrixRejects` |
| A2-4.3 | `TestSupersedeCycleRejected`, `TestSupersedeNonCurrentRejected` |
| A2-4.8 | `TestContentHashRecomputedFromStoredBytes` |
| A2-7.1 | `TestMaximalNodeFitsCaps` (a maximal node per kind: every field at its cap, `attrs` 16 keys × 512 B, `addresses` 16 × 64 B, `evidence_ids` 8) |
| A2-8.9 | `TestQuarantineFlagFlipDoesNotSkipRows` |
| A2-10.2 | `TestNodeWriteRejectsUnknownField` |
| A2-10.3 | `TestHardRejectMatrix` — one subtest per rejection: unknown node kind · unknown edge kind · wrong endpoint kind · not-applicable field · unparseable `cidr` · over-cap field · over-cap `attrs` · reserved key · nested/float/`null` `attrs` · self-edge · supersede of non-current · supersede cycle · cross-engagement endpoint · missing provenance |
| A2-10.8 | `TestRejectedWriteIsStillChained` |
| A2-11.4 | `TestGraphNodeIDFromAIsNotFoundInB`, `TestCursorFromEngagementARejectedInB`, `TestNodeDedupDoesNotSpanEngagements`, `TestEdgeDedupDoesNotSpanEngagements`, `TestNoCrossEngagementEdge`, `TestNoBareNodeIDInGraphDocuments` |

Single-oracle rulings (PAIR-K1 / AM-4 — one name, one oracle, no "either" in
an assertion):

- `TestGraphNodeIDFromAIsNotFoundInB` — the **only** oracle is: `notfound`
  (404) whose envelope `attrs.graph_node_id` **is the requested id** (ids are
  not secrets, A0-1.7) and whose `message` does not distinguish "absent" from
  "elsewhere" (A0-3.9).
- `TestCursorFromEngagementARejectedInB` — the **only** oracle is
  `validation` (400); "an empty page **or** `validation`" is not an acceptable
  assertion (A2-11.4, A1-8.2 is the same oracle).

Index of the ids carried by their own clause (the suite implements these too;
this index exists so one grep finds every A2 test id):

| Clause | Test ids |
|---|---|
| A2-3.6 | `TestContradictsDirectionNormalized` |
| A2-4.5 | `TestSupersedeChainBoundAtWrite` |
| A2-4.6 / §4.2 | `TestContentHashStableAcrossArrayOrder`, `TestContentDocFixedKeySet`, `TestContentHashVector` |
| A2-4.7 | `TestNodeDedupKeepsEveryObservation`, `TestVerifiedRequiresIndependentObservation`, `TestReplayedEventAddsNoProvenanceEntry`, `TestProvenanceListIsOrderedByEventSeq` |
| A2-5.3 | `TestPrincipalKindIsNotCopiedFromActor`, `TestFieldNameMappingIsTotal` |
| A2-5.6 | `TestVerifiedRequiresIndependentObservation`, `TestSelfObservedNodeStaysInferred` |
| A2-6.1 | `TestAttrValueRoundTrip`, `TestAttrsRejectNestedFloatNull` |
| A2-6.3 | `TestReservedAttrKeysRejected` |
| A2-8.2 | `TestPerKindMatchedFields`, `TestOneHopPropagation`, `TestIdentitylessNodeQuarantinedByDerivation`, `TestQuarantineRecomputedOnEveryTrigger` |
| A2-8.3 | `TestQuarantinedNodeNotTargetableWithValidApproval`, `TestTargetResolvedByIDNotByString` |
| A2-8.5 | `TestBlacklistedNodeSurvivesScopeWidening`, `TestBlacklistedDiscoveryIsRecordedNotRefused` |
| A2-8.7 | `TestReportExcludedRequiresQuarantine`, `TestMachinePrincipalCannotSetReportExcluded` |
| A2-8.8 | `TestQuarantineFlagCannotBeSuppliedOnWrite`, `TestOperatorReportExclusionIsEventLogged` |
| A2-9.4 | `TestEveryRuleIDMatchesItsCorpusValue`, `TestNoFalsePositiveOnBenignCorpus`, `TestErrorMessageNamesFieldAndRuleIDOnly`, `TestEntropyRuleIsDeterministic`, `TestGraphSecretFreeSerialization` |
| A2-9.7 | `TestGraphSecretFreeSerialization`, `TestCredentialNodeHoldsReferenceOnly`, `TestNoMaskingMappingInGraph` |
| A2-10.2 | `TestValidationOrderIsDeterministic`, `TestCollapseStillEmitsQuarantineEvent` |
| A2-10.6 | `TestNodeHasNoExportedContentSetter`, `TestGraphSeamHasExactlyFourMutationMethods`, `TestNodeDraftRejectsPlatformFields` |
| A2-11.4 | `TestGraphEngagementBReadNeverReturnsA`, `TestNoBulkReadSpansEngagements` |
| A2-12.5 | `TestQuarantinedNodeAbsentFromPlanningView`, `TestRetractedEdgeAbsentFromPlanningView` |

Safety-path pairing (`contracts/README.md` merge gate, E-06): every safety
clause above carries one positive and one negative id — A2-8.2
(`TestPerKindMatchedFields` / `TestOneHopPropagation`), A2-8.3
(`TestTargetResolvedByIDNotByString` / `TestQuarantinedNodeNotTargetableWithValidApproval`),
A2-8.5 (`TestBlacklistedDiscoveryIsRecordedNotRefused` /
`TestBlacklistedNodeSurvivesScopeWidening`), A2-8.7
(`TestOperatorReportExclusionIsEventLogged` / `TestReportExcludedRequiresQuarantine`),
A2-10.6 (`TestGraphSeamHasExactlyFourMutationMethods` /
`TestNodeDraftRejectsPlatformFields`), A2-4.7
(`TestNodeDedupKeepsEveryObservation` / `TestReplayedEventAddsNoProvenanceEntry`).

## 5. Traceability

| Source | Decision | Clauses |
|---|---|---|
| **ADR-0016 §1** | per-engagement graph; closed node kinds (hosts, networks, services, accounts/identities, groups, credentials, shares, findings/hypotheses, artifacts/evidence refs); closed edge kinds; provenance on every node/edge; property-graph tables in PostgreSQL | A2-1.1, A2-2.1–2.4, A2-3.1–3.2, A2-5.1–5.6 |
| **ADR-0016 §2** | scope alignment enforced on the graph: out-of-scope discoveries recorded as quarantined, never actionable; blacklist not representable as an actionable target; policy checks in the platform core on planning reads | A2-8.1–8.5, A2-10.1, A2-12.4/12.5 |
| **ADR-0016 §2 + ADR-0005 §3 (per-kind evaluation)** | quarantine is derived from a **closed per-kind identity field set**, propagates one hop, is recomputed on a closed trigger list, and is never derived from prose; a target is cited by `gn_` id, never by a worker-supplied string | A2-8.2 (matched fields, derivation, triggers, `QuarantineDecider`), A2-8.3, A2-10.2 (13a/13b/13c) |
| **ADR-0013 (offline buffering) + A0-3.11** | a replayed observation is idempotent and still attributable: the dedup collapse appends a provenance entry, a repeated `event_id` appends nothing, and the collapse is chained | A2-4.7 (bounded list ≤ 8, ordered by the A1 event `seq`, replay guard), A2-5.1, A2-5.6 (`verified` needs an independent second entry), A2-3.4, A2-10.2 (`dedup_hit`) |
| **A1-2.1 / A1-3.3 (PAIR-A1, PAIR-A2, PAIR-M1)** | one principal vocabulary and one name per value platform-wide | A2-5.3 (five kinds, `operator`→`user`, `operator_id`→`user_id`, `principal_kind` ≠ the event `actor`), A2-2.8, A2-1.6 (the A1↔A2 mapping table) |
| **A1-4.7 / A1 §4.3 / A0-2.15–2.16** | arrays are sorted and deduplicated **before** canonicalization, and a published fingerprint vector is normative data the shared suite reproduces byte-exactly | A2-4.6, §4.2 (F1, F1-R, F3, S1), A2-4.8 (recompute from stored bytes) |
| **A0-1.2 / A0 §4 `KindUser` (AM-1)** | a human principal has a registered id shape, so operator attribution is validatable | A2-5.3 (`user_id` = `usr_`), A2-8.7, §6 item 2 |
| **DESIGN §1 / DESIGN §4** | the domain layer imports foundation only; an interface is declared at its consumer; a constructor never returns a half-built value | §4 (`QuarantineDecider` declared in `internal/graph`, `NodeDraft`→`NewNode`→`PendingNode`→`WriteNode`), A2-8.2, A2-10.6, A2-11.1 |
| **AGENTS.md (high-review, safety-path pairing)** | one positive and one negative id per safety rule; untrusted-input validation is high-review | §4.3 (index + single-oracle rulings), A2-9.4, A2-10.2, A2-10.6, A2-6.7 |
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
| **ADR-0019 §5** | mandatory redaction: no captured credentials/tokens/secret material in errors or logs | A2-9.1, A2-9.4 (the normative closed rule table), A2-9.5 (never in whole, in part or as a digest), A2-9.3 (a filter, not a guarantee), A2-5.8, A2-9.7 |
| **ADR-0020 §3/§4** | per-run masking-mapping values are secret; captured credentials/hashes/tokens never reach a cloud endpoint under any policy; agents use opaque references | A2-9.1, A2-9.2, A2-9.3, A2-9.7 |
| **ADR-0008 / Q14** | tool registry id + version; name/version are registry fields, not part of the id | A2-5.3 (`tool_id`, `tool_version`), A0-1.3 cited |
| **Adversarial A1** | prompt injection / untrusted content: graph content is untrusted, never configuration, and a rejected write still leaves a trace | A2-6.7, A2-10.8, A2-5.2 |
| **Adversarial A12** | cross-engagement leakage | A2-11.1–11.4, A2-5.7 (`notfound`), A2-3.2 (`notfound` endpoints) |
| **A0-1.2** | `gn_` for graph nodes, `ge_` for graph edges, `evi_` for evidence | A2-1.2, A2-2.2, A2-3.1, §4 |
| **A0-3.6** | in errors and log attrs a graph node is `graph_node_id`; `node_id` means the remote agent node — never conflated | A2-1.2, A2-1.6 (+ the A1↔A2 mapping table), A2-5.3 (`agent_node_id`), §4.1 examples, §4.3 `TestNoBareNodeIDInGraphDocuments` |
| **A0-7.1 / A0-7.7** | `NodeSummaryMaxBytes` 512 B and `FindingSummaryMaxBytes` 2 KiB apply; every capped field class gets exactly one mechanism, recorded in a table | A2-7.1 (mechanism table, R for all A2 classes; registry citations, AM-2 granted), A2-7.2, A2-6.4 (canonical-form measurement), A2-2.2/2.3 |
| **A0-6.3** | unknown node kind, edge kind or status enum on write = hard reject `validation` | A2-2.1, A2-3.1, A2-2.6, A2-10.2 (steps 5, 7, 11), A2-10.5 |
| **A0-3.9** | an id from another engagement resolves to `notfound` (404), never `forbidden` — no existence disclosure | A2-11.3, A2-11.4, A2-5.7, A2-3.2 |
| **A0-2.14** | a canonicalized value has a fixed key set | A2-4.6 (`contentDoc`, empty A0-2.12 exclusion list, non-nil collections), A2-4.8 (A0-2.16 stored bytes), §4 `contentDoc`, §4.2 (the vectors that lock both) |
| **A0-5.7** | client-supplied time is `*_claimed_at` and untrusted | A2-5.3 (`observed_claimed_at`), A2-5.5 |
| **A0-8.8** | booleans are adjectives, no `is_` prefix | A2-8.1 (`quarantined`), A2-8.7 (`report_excluded`), A2-3.9 (`retracted`), A2-2.2 |
| **A0-4.3** | pagination order from immutable keys, never from a mutable flag | A2-1.4/A2-1.4a (`graph_seq`, one per-engagement sequence shared by nodes and edges), A2-8.9, A2-4.5 |
| **A0-7.10** | the Q4 caps are not jointly satisfiable for a maximal stage view; the composition rule belongs to A3 | A2-12.2, A2-12.3 (what A2 guarantees A3 can rely on), §6.11 |
| **A0-6.2 / A0-8.3 / A0-8.5 / A0-8.2 / A0-2.6 / A0-8.9** | unknown field on write rejected; absent not null; closed enum spelling; fixed suffixes; integers only with the scale fixed by the owning contract; bounded bodies | A2-1.5, A2-1.7, A2-2.1/2.4–2.6, A2-1.6, A2-2.5 (`cvss_v3_x10` ×10), A2-10.2 step 2 |
| **ADR-0001 / ADR-0010** | stdlib only; pgx only behind the store seam | §4 (no third-party type; `net.ParseCIDR`, `regexp`, `encoding/json` only), A2-1.1, A2-10.6 |

## 6. Open for product owner

Recommendations that are genuinely product-owner calls. Each is marked **PO
confirm** at the clause too; none is decided silently. A2 cites **current** A0
throughout (A0-1.2's `usr_`, A0-7.1's registry rows, A0-3.6's `node_id`
reservation); the four A0 amendment requests below were ruled in the Freeze
triage and are recorded with their rulings, not left as open asks.

1. **A2-5.6 / A2-2.7 — confidence semantics.** Recommend the evidence grade
   `observed` · `inferred` · `verified` instead of `low`/`medium`/`high`.
   _ADR-0016 §1: graph content is evidence, not opinion — a grade tied to a
   referenced event is checkable, an adjective about certainty is not._ Also
   confirms that Q2's "Finding carries confidence" is discharged by the
   mandatory provenance block rather than a second finding field. This is a
   **change to a locked PO decision and needs a signature, not a confirmation**
   — item 13 is the wording the product owner signs.
2. **A2-5.3 — the human-principal id shape (AM-1): resolved by default for the
   Freeze.** A0-1.2 now registers `usr_` (`^usr_B{26}$`, 30 B) and A0 §4 adds
   `KindUser`, so a `principal_kind: user` entry's `user_id` validates per
   A0-1.5; operator corrections and A2-8.7 report exclusion are unblocked and
   nothing here waits on A5 (A0 owns id *shapes* — delegating the spelling
   would split A0-1.5 validation across two contracts). BLOCK-PO8 in A2-5.3 is
   the normative text. What remains is a **PO confirm** of the prefix spelling
   only, and it is additive-only afterwards (A0-1.10). The former reading of
   this item — "`operator_id` has no id shape, so `principal_kind: operator`
   writes MUST be refused" — is withdrawn: the enum value is `user` and the
   field is `user_id` (A2-5.3, A2-2.8).
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
9. **A2-5.4 — the A1 envelope and the A1 kinds A2 consumes.** The envelope
   fields A2 consumes (`event_id`, `kind`, `recorded_at`, `seq`,
   `engagement_id`, `run_id`, `job_id`) are **A1-1.1's**: A1 owns the envelope,
   and A0-2.17 V5 is a canonicalization vector with a synthetic key set, not a
   valid event. Every kind A2 needs exists in A1-3.3's closed list — node
   written (`graph_node_written`, with `dedup_hit`) · node quarantined
   (`graph_node_quarantined`) · quarantine recomputed
   (`quarantine_recomputed`) · node report-excluded
   (`report_inclusion_changed{included:false}`) · node superseded
   (`graph_node_written` plus `graph_edge_written{edge_kind:"supersedes"}`;
   there is no separate supersession kind and none is needed, A2-4.1/4.2) ·
   edge written (`graph_edge_written`) · edge retracted
   (`graph_edge_retracted`) · write rejected
   (`action_blocked{reason:"graph_write_rejected", action_kind:"graph_write"}`)
   · target refused (`action_blocked{reason:"target_quarantined"}`, A2-8.3).
   This item is a record, not an open question: a future A1 kind rename is
   caught by this list.
10. **A2-9.4 — secret scan outcome.** Recommend hard reject (`validation`,
    field + rule named, value never echoed). Alternative: redact the value,
    store a marker, keep the node. _Reject is the Q3 posture and keeps the
    graph clean; redact risks a false positive silently destroying evidence —
    but reject risks stalling an engagement on a false positive, which is why
    the pattern set must ship with the contract test corpus._

    **Ruled once for both contracts (PO confirm): reject, never redact.** A secret-pattern hit (A2-9.4 rule ids) is a hard reject — `validation` (400) naming the field and the rule id, the value never echoed in whole, in part or as a digest (A0-3.4) — and the rejection is chained (`action_blocked`). Redaction was rejected: a false positive would silently destroy a worker's only report of what it ran, and `redacted:true` (A1-4.6) means platform redaction, never rejection. The false-positive risk is controlled by shipping the A2-9.4 rule table with the planted-secret corpus. A1 §6.4 and A2 §6.10 are the same question and MUST NOT be answered differently.
11. **A2-12.3 — A0-7.10 is explicitly not resolved here.** A2 states only what
    A3 may rely on (A2-12.2). The 500-node/64 KiB composition rule stays with
    A3 and the PO escalation already recorded in A0 §6.14.
12. **A2-8.5 — blacklisted discovery: recorded, not refused (PO confirm).** The Freeze stores the node with `quarantine_reason:"blacklisted"`, never releasable, never in a planning view, and chains `graph_node_quarantined{blacklist_match}` — "we saw the forbidden target and did not touch it". The alternative reading of ADR-0016 §2 (refuse the write, store nothing about a forbidden system) is defensible and minimizes stored data; the product owner MUST confirm before Frozen, because refusing the write makes the near-miss unprovable in a customer report.
13. **A2-2.7 / A2-5.6 — deviation from Q2 (PO signature required, not confirmation).** Q2 records "Finding carries confidence". A2 implements that as the mandatory provenance grade `observed · inferred · verified` on every node and edge instead of a finding field, so a grade is always tied to a referenced event (ADR-0016 §1: evidence, not opinion). This **changes a locked decision**; the product owner MUST sign it before A2 flips to Frozen. With §4 A2-04 (provenance set) `verified` is now reachable: it requires a second, independent observation.
14. **PO confirm — operator release of quarantine is removed.** `operator_release` is deleted from A1's `quarantine_kind` enum and A2 provides no release operation: an `out_of_scope` node is released **only** by an operator scope change and the recomputation it causes (A2-8.5, ADR-0016 §2 — an out-of-scope node can never be a target of a planned action). `operator_quarantine` (tightening) is kept, and a `blacklisted` node is never releasable. If the product owner wants a manual release it MUST be a new ADR amending ADR-0016 §2 and MUST require the target to be inside the widened allowlist at release time.

### A0 amendment requests

A2 cites current A0; these were requests, and all four are ruled (plan §3 of
the principal review, applied at the Freeze). They are kept as a record of what
A2 depends on in A0, with the ruling in the last column.

| # | A0 clause | Request | Ruling and why A2 needs it |
|---|---|---|---|
| AM-1 | A0-1.2 | Register a prefix for a human principal (`usr_`) | **Resolved by default for the Freeze** (plan §1 PO-8, BLOCK-PO8 in A2-5.3): A0-1.2 registers `usr_` (`^usr_B{26}$`, 30 B) and A0 §4 adds `KindUser`, so A2-5.3's `user_id` validates per A0-1.5. Not delegated to A5 — A0 owns id *shapes*. **Awaits PO confirmation of the spelling only** (§6 item 2); additive-only afterwards (A0-1.10) |
| AM-2 | A0-7.1 | Adopt the A2 cap constants (`NodeSummaryMaxBytes`, `FindingSummaryMaxBytes`, `MaxSupersedeChain`, `AttrsMaxKeys`, `AttrKeyMaxBytes`, `AttrValueMaxBytes`, `AttrsTotalMaxBytes`, `AddressesMax`, `EvidenceRefsMax`, `ToolVersionMaxBytes`) into the single A0 table/const block | **Accepted** (PAIR-N1): A0-7.1 is the one place caps live, and A2-7.1 now cites the registry rows instead of declaring values. Two constants moved with it — A2's `EvidenceIDsMax` is gone in favour of `EvidenceRefsMax` = 8, and A2's former `ToolVersionMaxBytes` = 32 was a defect (the registry value 64 governs). A2-local constants stay A2's under A0-7.7's delegation |
| AM-3 | A0-3.6 | Extend the `node_id` reservation from "errors and log attrs" to graph payloads, and bless A2's payload spellings (`source_id`, `target_id`, `supersedes_id`, `superseded_by_id`, `agent_node_id`) | **Accepted in part.** The reservation is granted: `node_id` means the remote agent node (`slp_node_`) everywhere, so A2-1.6's rule is enforceable and `TestNoBareNodeIDInGraphDocuments` has an A0 basis. Blessing A2's *field names* is refused — A0 does not own per-contract payload vocabularies. A2 publishes the mapping instead: A2-1.6's A1↔A2 table (BLOCK-A2-14 / PAIR-M1) is the single source, cited by A1-3.6 |
| AM-4 | A0-8.2 | Confirm `*_ref` is not needed: A2 uses `evidence_id` / `evidence_ids` for `evi_` references because A0-8.2 fixes `*_id` for identifiers | **Accepted**: no `*_ref` suffix is added, A2-9.2's spellings stand, and `evidence_refs` (A1-1.1) remains the single approved exception to A0-8.2 — one suffix convention per fact, no drift |

### Cross-contract requests (not A0)

- **A1** — the envelope field names and the event kinds item 9 lists (all
  confirmed present in A1-1.1/A1-3.3: `graph_node_written`,
  `graph_node_quarantined`, `graph_edge_written`, `graph_edge_retracted`,
  `quarantine_recomputed`, `report_inclusion_changed`,
  `action_blocked{graph_write_rejected, target_quarantined}`); the ingest dedup
  key A2 relies on for idempotent replay (A2-4.7, A0-3.11); and the
  byte-identical PAIR blocks A2 also carries (BLOCK-A2-14's mapping table,
  BLOCK-PO6, BLOCK-PO9, PAIR-Q1/Q2/SEC1/A1/A2).
- **A3** — the A0-7.10 composition rule, the planning-vs-reporting view
  mapping (A2-12.4), and its own mechanism-T assignments (A0-7.7).
- **A4** — per-endpoint body size bounds (A0-8.9, A2-10.2 step 2), the history
  read parameter (A2-4.5), route-table audit for A2-11.4's
  `TestNoBulkReadSpansEngagements`.
- **A5** — every graph-write verb on the machine-principal exclusion list
  (A2-10.1, Q6). The human-principal id shape is **not** A5's: A0-1.2 owns
  `usr_` (AM-1 resolved, §6 item 2).
