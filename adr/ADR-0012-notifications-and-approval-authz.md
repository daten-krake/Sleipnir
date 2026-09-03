# ADR-0012: Notifications and approval authorization

- **Status:** Accepted (confirmed by product owner 2026-09-03)
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

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
3. **Delivery channel (v1): generic signed webhooks** — platform POSTs a
   JSON payload (HMAC-signed) to any configured URL. This single mechanism
   covers Slack/Discord/Matrix/ntfy.sh/Telegram bot relays without
   per-vendor code. Optional SMTP email deferred (not v1).
4. Payload includes a deep link to the approval/action page.
5. In-UI notifications via SSE regardless of external channels.
6. Retries with backoff + delivery log (notifications are themselves
   audited events).
7. **Approval timeout:** pending approvals **expire after 2 hours** by
   default; the timeout is **configurable per engagement in the web GUI**.
   Expired actions return to the orchestrator as `approval_expired` so it
   can replan instead of waiting forever.

## Consequences

- **Positive:** all-stdlib; vendor-neutral; works for consultants with any
  chat stack; no inbound exposure (all pushes are outbound HTTP); no stalled
  scans — expiry keeps the orchestrator moving.
- **Negative:** webhook UX requires per-project configuration; no built-in
  rich vendor integrations; expiry may abort long-unattended scans (by
  design).
- **Follow-ups:** webhook payload schema (with API design); GUI settings
  screen for timeout + channel config.
