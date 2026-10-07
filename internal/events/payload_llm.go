package events

// LLMCallPayload is the payload of kind llm_call (A1-3.3, ADR-0020 §2–§5):
// the operator-visible egress log, metadata only — the platform must not
// compose this kind with prompt or completion text in any field (A1-4.2).
// ModelRole is the closed enum orchestrator | enumeration | research |
// summarizer (ADR-0014); EgressPolicy is the closed enum local_only |
// cloud_masked | cloud_raw; Status is the closed enum ok | error, and an
// error carries a non-empty ErrorKind (A0-3.1). EndpointName and ModelName
// are configured names — a URL must not be stored, it may embed a credential
// (A0-3.7). PromptTokens/CompletionTokens are upstream-reported: integers of
// untrusted origin, not prose, so A1-4.4's marking does not apply, and a
// consumer must not treat them as billing truth (A1-4.12).
type LLMCallPayload struct {
	ModelRole           string `json:"model_role"`    // enum orchestrator|enumeration|research|summarizer
	ModelName           string `json:"model_name"`    // ≤ 128 B configured name, never a URL (A0-3.7)
	EndpointName        string `json:"endpoint_name"` // ≤ 128 B configured name, never a URL (A0-3.7)
	EgressPolicy        string `json:"egress_policy"` // enum local_only|cloud_masked|cloud_raw
	Status              string `json:"status"`        // enum ok|error
	ErrorKind           string `json:"error_kind"`    // A0-3.1 kind; non-empty when Status == "error"
	RequestBytes        int64  `json:"request_bytes"`
	ResponseBytes       int64  `json:"response_bytes"`
	PromptTokens        int64  `json:"prompt_tokens"`
	CompletionTokens    int64  `json:"completion_tokens"`
	MaskedEntityCount   int64  `json:"masked_entity_count"`   // ADR-0020 §3
	ExcludedSecretCount int64  `json:"excluded_secret_count"` // ADR-0020 §4
	DurationMS          int64  `json:"duration_ms"`
	RequestEvidenceID   string `json:"request_evidence_id"`
	ResponseEvidenceID  string `json:"response_evidence_id"`
}

// Kind returns KindLLMCall.
func (LLMCallPayload) Kind() Kind { return KindLLMCall }
