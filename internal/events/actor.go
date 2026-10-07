package events

// ActorType is the closed actor vocabulary of A1-2.1: user | orchestrator |
// worker | platform | node. A2-5.3 adopts this list verbatim as its
// principal_kind vocabulary ("operator" is renamed "user" platform-wide).
type ActorType string

const (
	ActorUser         ActorType = "user"
	ActorOrchestrator ActorType = "orchestrator"
	ActorWorker       ActorType = "worker"
	ActorPlatform     ActorType = "platform"
	ActorNode         ActorType = "node"
)

// Component is the closed enum of A1-2.4 naming the platform subsystem that
// composed the event, so an audit reader can attribute a platform-composed
// event without a code search (ADR-0019 §3, C9). It is "" unless
// Actor.Type == ActorPlatform (A1-2.1, A1-3.3). An unknown value is a
// platform defect → validation at composition, surfaced as internal to the
// caller (A0-3.1) and logged (A0-3.8) — enforced by the composition path
// (A1-7.10), not by this type.
type Component string

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

// Actor is exactly three keys, all always present (A1-2.1, A0-2.14). It is
// stamped by the platform from the authenticated principal and the request
// context (A1-2.3); a client MUST NOT supply it. PrincipalID is the A0-1.2
// identifier selected by Type (A1-2.2: usr_ for user, job_ for orchestrator,
// task_ for worker, slp_node_ for node, "" for platform) and is validated
// per A0-1.5 at composition. Component is "" unless Type is platform.
type Actor struct {
	Type        ActorType `json:"type"`
	PrincipalID string    `json:"principal_id"`
	Component   Component `json:"component"`
}
