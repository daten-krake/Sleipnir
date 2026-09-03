# ADR-0009: Evidence capture, command log, and environment cleanup/revert

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Pentest deliverables are not just findings: clients expect full
reproducibility and a clean environment afterwards. The product owner
requires: detailed command log, screenshots, and the ability to **revert
changes made in the client environment**.

## Decision

1. **Central event/command log:** every executed command, its output
   reference, actor (agent/tool), target, and timestamp lands in the
   central log (append-only). This doubles as audit trail and evidence store.
2. **Evidence artifacts:** screenshots, captured files, dumps, etc. are
   stored per engagement with references from the log.
3. **Revert plans:** any state-changing action (created accounts, changed
   ACLs, dropped files, scheduled tasks, registry edits...) must record a
   **revert record** describing how to undo it, captured at action time.
4. **Cleanup run:** at engagement end (or on demand) Sleipnir produces and,
   after human approval, executes a cleanup plan from the revert records.
5. **Reports:** HTML report (stdlib templates), PDF export option; reports
   embed/link the detailed log and evidence.

## Consequences

- **Positive:** strong differentiator (auditable + reversible pentesting);
   the event log is also the backbone for persistence and replay.
- **Negative:** revert is inherently imperfect (password changes, data
  touched, detection artifacts can't be "un-seen"); every tool integration
  must produce revert records — extra implementation cost per tool.
- **Follow-ups:**
  - Define which action classes are reversible and the revert-record schema.
  - Screenshot sources: headless browser (later web phase), RDP session
    captures for AD? — needs discussion.
  - ADR: report format/PDF generation approach (external renderer in a
    container is the likely answer).
