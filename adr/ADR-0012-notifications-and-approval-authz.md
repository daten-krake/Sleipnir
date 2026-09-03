# ADR-0012: Notifications and approval authorization

- **Status:** Proposed (product owner asked for a design; direction to confirm)
- **Date:** 2026-09-03
- **Deciders:** architect proposing, product owner to confirm

## Context

Dangerous actions wait for human approval (ADR-0005). Operators must be
notified when attention is needed and when scans finish or fail. Only the
operator(s) assigned to an engagement may approve its actions.

## Decision (proposed)

1. **Approval authorization:** each engagement has assigned operators; only
   they (plus admins) can approve actions of that engagement. Approvals are
   recorded with approver identity + timestamp in the event log.
2. **Notification service** in the platform core, driven by events:
   `approval_required`, `scan_started`, `scan_finished`, `agent_error`,
   `hard_stop_fired`, `cleanup_proposed`.
3. **Delivery channels (v1): generic signed webhooks** — platform POSTs a
   JSON payload (HMAC-signed) to any configured URL. This single mechanism
   covers Slack/Discord/Matrix/ntfy.sh/Telegram bot relays without
   per-vendor code. Optional: SMTP email via stdlib `net/smtp`.
4. Payload includes a deep link to the approval/action page.
5. In-UI notifications via SSE regardless of external channels.
6. Retries with backoff + delivery log (notifications are themselves
   audited events).

## Consequences

- **Positive:** all-stdlib; vendor-neutral; works for consultants with any
  chat stack; no inbound exposure (all pushes are outbound HTTP).
- **Negative:** webhook UX requires per-project configuration; no built-in
  rich vendor integrations.
- **Follow-ups:** confirm channel priority (webhook-only v1? email too?);
  approval **timeout behavior still open** (pause forever vs. auto-expire).
