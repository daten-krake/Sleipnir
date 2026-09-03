# ADR-0016: Engagement-scoped context graph (agent working memory)

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

The agent loop (ADR-0015) is iterative: agents must self-correct and learn
during a running engagement. Raw chat-style context does not scale and
bloats model windows (ADR-0007 context hygiene). The product owner proposed
a **context graph, strictly scoped to the current test/customer**, aligned
with per-engagement policy (scopes as allowlist, blacklist).

## Decision

1. Every engagement owns a **context graph**: its working memory and the
   substrate for learning during that test.
   - **Nodes:** hosts, networks, services, accounts/identities, groups,
     credentials, shares, findings/hypotheses, artifacts/evidence refs.
   - **Edges:** relationships (reachable, authenticates-to, member-of,
     grants-access, exploited-by, contradicts, supersedes).
   - Every node/edge carries **provenance** (agent, tool, event-id,
     timestamp, confidence) — graph content is evidence, not opinion.
2. **Scope alignment is enforced on the graph itself:** nodes outside the
   engagement's allowlist scope may be *recorded as quarantined
   discoveries* but can never be targets of planned actions; blacklist
   entries cannot be represented as actionable targets at all. Policy
   checks (ADR-0005) remain in the platform core and apply on graph reads
   used for planning.
3. **Learning boundary:** learning is **per-engagement only**. No graph
   content, tactics, or findings flow between engagements/customers.
   Cross-engagement learning (e.g. aggregated attack-pattern statistics) is
   explicitly deferred and would require its own ADR with consent +
   anonymization rules.
4. **Consumers:**
   - Orchestrator plans by **querying the graph** instead of scrolling
     history → this is how context stays small.
   - Workers write findings into the graph + event log; handovers between
     stages are *views over the graph* (ties into the handover-contract
     session).
   - Self-correction: failed attempts and contradictions are edges; agents
     can revise hypotheses (a new node `supersedes` the old one) — history
     is preserved, never overwritten.
5. Storage: property-graph-shaped tables in PostgreSQL (ADR-0010); no
   external graph database (keeps the dependency rule).

## Consequences

- **Positive:** enables iterative/self-correcting agents; deduplicates work;
  gives the report its attack-path backbone; makes "what do we know"
  queryable and auditable; per-customer data isolation by construction.
- **Negative:** schema design is non-trivial; graph quality depends on
  agents writing disciplined provenance; quarantined-discovery handling
  needs care (recording out-of-scope data is itself sensitive).
- **Follow-ups:**
  - Graph schema design (with persistence-schema session).
  - Graph query surface exposed to agents (bounded query API, not raw SQL).
  - Attack-path visualization for the UI (BloodHound-style view) → UI
    design session.
  - Handover contracts = graph views → handover-contract session.
