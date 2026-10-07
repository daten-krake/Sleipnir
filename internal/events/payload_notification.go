package events

// NotificationSentPayload is the payload of kind notification_sent
// (A1-3.3, ADR-0012 §2/§6): retries and delivery of notifications are
// themselves audited events. NotificationKind is the ADR-0012 §2 vocabulary
// (approval_required | scan_started | scan_finished | agent_error |
// hard_stop_fired | cleanup_proposed) plus A1's seventh value
// chain_head_anchor for the A1-5.8 (4) out-of-band head anchor (product
// owner decision 2026-09-21, PR #2 item D5) — these are NOTIFICATION kinds,
// distinct from A1 event kinds. Channel is the closed enum webhook | sse;
// DeliveryStatus the closed enum delivered | failed. TargetName is the
// configured channel name, never a webhook URL (A0-3.7). Attempt is ≥ 1
// (A1-4.2). For chain_head_anchor, RelatedEventID names the event whose
// append crossed the emission point, or the chain_verified event.
type NotificationSentPayload struct {
	NotificationKind string `json:"notification_kind"` // enum approval_required|scan_started|scan_finished|agent_error|hard_stop_fired|cleanup_proposed|chain_head_anchor
	Channel          string `json:"channel"`           // enum webhook|sse
	TargetName       string `json:"target_name"`       // ≤ 128 B configured name, never a URL (A0-3.7)
	DeliveryStatus   string `json:"delivery_status"`   // enum delivered|failed
	Attempt          int64  `json:"attempt"`           // ≥ 1 (A1-4.2)
	HTTPStatus       int64  `json:"http_status"`
	DurationMS       int64  `json:"duration_ms"`
	RelatedEventID   string `json:"related_event_id"`
}

// Kind returns KindNotificationSent.
func (NotificationSentPayload) Kind() Kind { return KindNotificationSent }
