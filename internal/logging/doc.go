// Package logging provides the structured-logging primitives mandated by
// ADR-0019 §3: a JSON log/slog setup, subsystem loggers, the correlation
// attribute set, and the single shape of an error-level record.
//
// It sits at the foundation layer (DESIGN §1). Every package above it may use
// it; it imports nothing from internal/ — not even internal/errs — which keeps
// the dependency arrows pointing downward only. That boundary is mechanical, not
// a review convention: TestLoggingHasNoInternalImports fails the build if an
// internal import appears. The consequence is by design: because this package
// cannot read an error's kind or origin operation, the caller passes the origin
// in as a plain string, produced by errs.OpOf(err) at a call site that may
// import both packages.
//
// # Attribute spelling
//
// The correlation attributes follow the frozen contract A0-3.6, not the wording
// of ADR-0019 §3: engagement_id, run_id, job_id and node_id, where node_id is
// always the remote agent node (slp_node_, Q9). A graph node is graph_node_id
// (gn_) and is never accepted here — Correlation has no graph-node parameter, and
// a graph-facing handler adds graph_node_id itself, so the conflation A0-3.6
// forbids cannot happen by construction. ADR-0019 §3 lists these attributes as
// engagement, run, job and node; that spelling predates ADR-0016's graph
// vocabulary, and A0-3.6 rules for error envelopes and slog attributes alike so
// the two stay one vocabulary.
//
// # Log-or-return
//
// ADR-0019 §3 allows an error to be logged or returned, never both. ErrorRecord
// is the logging half: it is called by the layer that handles or decides
// (A0-3.8), it does not return the error and it does not wrap it. Inner layers
// wrap with errs and return without logging.
package logging
