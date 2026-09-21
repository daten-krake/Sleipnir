// Package errs is the platform's only source of errors (ADR-0019 §1) and the
// owner of the A0-3 client-facing error contract: the closed kind vocabulary,
// its HTTP mapping, the /api/v1 error envelope, and the redaction helpers that
// keep secret material out of every error string.
//
// Layer: foundation (DESIGN §1). It imports nothing from internal/ and is
// imported by every other foundation package (A0 §4) and by all layers above.
//
// Every returned error in the platform is created or wrapped here, never with
// errors.New or fmt.Errorf (AGENTS.md). Each error carries the function it
// originated in, captured automatically via runtime.Caller, and renders as
//
//	component.Function: what was attempted: key identifiers: cause
//
// per ADR-0019 §2, so a log line alone is enough to locate a failure.
//
// Kinds are attached at creation. ADR-0019 §1 sketches New(msg) and separately
// requires that "error kinds [are] attached at creation so handling is
// programmatic"; New(kind, msg) is how both hold. Wrap and Wrapf keep the ADR's
// signature and inherit the cause's kind, so wrapping can never silently
// reclassify an error (A0-3.12: a kind must not be repurposed, because agent
// behaviour is keyed to it).
package errs
