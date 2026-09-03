# ADR-0015: Agent runtime — own loop vs. embedding an existing harness

- **Status:** Proposed (architect recommendation; product owner to confirm)
- **Date:** 2026-09-03
- **Deciders:** architect proposing, product owner to confirm

## Context

Agents run in Docker containers (ADR-0007). Question raised by the product
owner: should those containers embed an existing agent harness (e.g. pi,
opencode) instead of us building the agent loop ourselves?

## Options considered

1. **Own minimal agent loop in stdlib Go** (prompt → model → tool call →
   sandboxed execution → summarized result).
   - Pros: scope/blacklist/approval enforcement sits directly on the
     execution path (our safety model *is* the loop); no Node/other runtimes
     in images; output lands natively in our event log (ADR-0009); fully
     auditable, no external moving parts; the loop is the product's core
     value anyway and is bounded in complexity (JSON + HTTP streaming).
   - Cons: we build context management, retries, tool-call parsing ourselves.
2. **Embed existing harness** (run a coding-agent CLI as subprocess in the
   container).
   - Pros: mature loop immediately; large tool/MCP ecosystem.
   - Cons: harnesses are built for software development, not offensive
     security; their tool surface (broad shell/file access) fights our
     approval gate, which would have to intercept inside the harness; logs
     arrive in foreign formats and must be translated; images carry foreign
     runtimes; supply-chain and audit surface grow; vendor assumptions about
     models/usage; much harder to prove *what the agent did* — the exact
     property our product promises.
3. **Hybrid:** own loop for production; harnesses used only as development
   tooling *on* Sleipnir.

## Decision (recommended)

Option 1 for the platform runtime, with Option 3's spirit: harnesses like
pi/opencode remain excellent tools for *building* Sleipnir, but the
production agent runtime is our own thin, enforceable loop. The tool
ecosystem angle returns later via the tool registry (ADR-0008), and Sleipnir
tools could be exposed outward (e.g. MCP) in the purple-team phase.

## Consequences

- **Positive:** safety model and evidence model are first-class; images stay
  lean; no foreign runtime supply chain.
- **Negative:** upfront investment in the loop; must match harness-quality
  context/retry handling ourselves.
- **Follow-ups:** agent loop design session (ties into the handover-contract
  and program-layout sessions on the backlog).
