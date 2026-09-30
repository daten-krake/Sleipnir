// Package secretscan is the closed A2-9.4 secret-pattern rule table that ingest
// scanning runs over every string field and every attrs value before the value
// is stored (A2-9.4, A1-4.9). It exports exactly one function, Scan, and one
// decision: a value that matches is rejected (DESIGN §2).
//
// Layer: foundation. Like the other foundation packages it imports exactly one
// internal package, internal/errs, for the error kind and message shape of
// ADR-0019. It imports nothing else: the rule set has to be reviewable against
// A2-9.4's table on one screen.
//
// # Clauses implemented here
//
//   - A2-9.4: the ten rules of the normative table, in table order, each
//     transcribed byte-exactly from contracts/A2-graph.md. The table is closed
//     and additive-only (A0-6.5): the set of rule ids this package can report
//     is exactly the ten, the reported string is the table's own string
//     (rulings S1, S2), and the first match in that order is the one reported.
//   - A2-9.4 / A1-4.9 reject-with-validation: the error kind is
//     errs.Validation (A0-3.1).
//   - A2-9.5 with A0-3.4's single exception: the message names the field, the
//     rule id and the byte length only. A value rejected by a secret-pattern
//     rule MUST NOT be echoed in whole, in part or as a digest — not in the
//     message, not in an error attrs entry, not in any slog record.
//   - A2-9.3 (the scanning half): the pattern set is A2-9.4's table and nothing
//     else; rule ids live in exactly one document.
//
// # Clauses deliberately NOT implemented here (who owns them)
//
//   - ADR-0020 §4, the gateway egress exclusion by kind, and
//     llm_call.excluded_secret_count: the llm package. That is the enforcement
//     point; this package is a filter.
//   - A0-7.1 mechanism R field caps: internal/caps, applied by the caller
//     BEFORE calling Scan. Scan is not a size guard and does not bound work.
//   - A2-9.2, the credential-node reference-only rule: internal/graph (WP-15).
//   - A2-9.6, the worker procedure (upload to the evidence store, reference the
//     evi_ id, never inline the bytes): the worker.
//   - A1-7.5, the action_blocked{reason:"append_rejected"} observability of a
//     rejection: internal/events (WP-11), which decides what to log.
//   - TestEventSecretFreeSerialization and TestGraphSecretFreeSerialization:
//     the shared contract suite (WP-20/WP-21).
//
// # This is a filter, not a guarantee
//
// A2-9.3 is explicit. A hostile worker can encode, split or re-format a secret
// past any pattern set, so nothing here may be read as "no secret can reach a
// model". The enforcement point is the gateway egress exclusion of ADR-0020 §4,
// applied independently to every string that leaves the platform (and, for an
// engagement whose policy is not local_only, by node kind). The two defences are
// independent and neither substitutes for the other: a field that clears Scan can
// still be excluded at the gateway, and a field that Scan rejects has already
// been kept out of storage.
//
// # Why ten patterns on hostile input is bounded (the A14 argument)
//
// Go's regexp is RE2: every pattern compiles to a finite automaton that is run
// over the input once, in time linear in the input length, with no
// backtracking and no exponential blowup on an alternating-traversal string.
// There is therefore no catastrophic-backtracking class in this table, and ten
// patterns over an adversarial field cost O(10n) with n the byte length of the
// value. SEC-ENTROPY is likewise a single linear pass. The scan does not
// reserve, cap or shrink the value: field caps belong to internal/caps (A0-7.1).
//
// # SEC-ENTROPY: the interpretation this implementation takes
//
// The clause reads "Shannon entropy >= 4.5 bits/char over a window of >= 32
// characters drawn from a base64/hex alphabet [A-Za-z0-9+/=_-]". A sliding
// window is NOT well defined by that wording: ">= 32" fixes no window size, so
// every length from 32 to the whole input qualifies and the scan becomes
// O(n*w) on hostile input — which is exactly the blowup the determinism
// requirement of A2-9.4's SEC-ENTROPY row ("MUST be deterministic
// (TestEntropyRuleIsDeterministic)"; A2-9.3 itself is cloud egress and states no
// such requirement) and the A14 bound argument forbid. So instead of a window,
// this
// package takes every MAXIMAL run of alphabet characters, measures
// -Σ p(c)·log2 p(c) over each run of length >= 32, and calls a run at or above
// 4.5 bits/char a match. This is an interpretation of a frozen clause, not a
// ruling: a fixed-size sliding-window reading would differ only where a long
// run mixes high-entropy material with low-entropy filler, since it could flag
// a 32-character window that the maximal-run measurement averages away. The
// maximal-run reading never misses a short high-entropy run; it can decline to
// flag a high-entropy substring that is padded inside one longer run.
//
// Two consequences of that reading, both asserted in
// TestEntropyRuleIsDeterministic:
//
//   - A 64-character lowercase hex digest never trips SEC-ENTROPY: its alphabet
//     has 16 symbols, so entropy is at most log2(16) = 4.0 bits/char, under the
//     4.5 bound by construction. Artifact digests and the sha256:<64 hex>
//     image-digest form of A0-8.7 therefore pass.
//   - A random 64-character base64 token does trip it: over 500 seeded samples
//     its empirical entropy lands in [4.82, 5.47] bits/char. The ceiling is
//     log2(64) = 6.0 and is reached only when all 64 symbols are distinct, so
//     the ~5.8 figure in the work-package brief is optimistic; the measured
//     range clears 4.5 comfortably all the same.
//
// # SEC-ENTROPY's known false-positive class (A2-9.4 erratum, RULED 2026-09-29)
//
// Recorded the way A2-9.4 records SEC-NTLM's "Known false-positive class", and
// for the same reason: the class is documented, the rule stays as written.
// This table is closed and additive-only (A2-9.4, A0-6.5), so this package MAY
// NOT narrow it — no threshold change, no window change, no alphabet change,
// no exemption parameter on Scan.
//
// The product owner ruled the remedy on 2026-09-29 (A2 §6 item 15, mirrored by
// A1 §6 item 18): a **platform-minted exemption**, applied by the CALLER, not
// here. A caller skips Scan for a field whose value passes ids.Valid(k, v) for
// the A0-1.2 kind its contract declares for that field, and for a field the
// platform stamps and A1-2.3 forbids a caller to supply. The rule table above
// is byte-identical to A2-9.4 and Scan's signature is unchanged. WP-11
// (internal/events) and WP-15 (internal/graph) MUST implement the caller side;
// wiring this scan without it rejects roughly one Pi-node event in four.
//
// That this package cannot implement the exemption is enforced mechanically,
// not by a test: doing so would need ids.Valid, and this package's import set
// is internal/errs plus stdlib. `go list -deps ./internal/secretscan` is the
// guard — a second internal import fails it.
//
// The class is the platform's own identifiers. SEC-ENTROPY rejects them: of
// 20 000 ids.New(ids.AgentNode) values measured against this table, 5 671
// (28.4%) are rejected — a re-run of the same measurement here gave 25%, so
// the figure is sampling-dependent but the class is not in doubt. Example:
// slp_node_01m3pqy611zczgaf4ajy4qkbxq measures H = 4.5147 bits/char. Cause:
// "slp_node_" plus a 26-character body is 35 characters, the only A0-1.2 form
// reaching the 32-character window, and underscore is inside the rule alphabet
// [A-Za-z0-9+/=_-], so the whole id is ONE run. The class is wider than node
// ids: container and network names built from an id body trip it too
// (sleipnir-worker-<26 chars> H = 4.6241; slp-run-<26 chars>-net H = 4.5568).
//
// Why that is a problem and not a curiosity: A1 envelope key 6 is
// `node_id:string slp_node_` and A1-4.9 mandates this scan over every string
// field of an append payload, so roughly one Pi-node event in four would be
// rejected at ingest.
//
// What does NOT trip it (all measured): 64-character lowercase hex digests
// (H <= 3.96, and the sha256:<64 hex> form splits at the colon anyway),
// 32-hex MD5 shapes (H = 3.39), RFC 3339 timestamps and argv — none of which
// puts a 32-character run on the board.
//
// # Rulings taken, and where they were challenged
//
// Numbered locally and consecutively (S1–S6). The WP-13 brief's own numbering
// skipped S3 and S7; neither numbering exists anywhere else in the repo, so the
// gaps below are not a missing ruling.
//
//   - S1 table order, first match wins, no map iteration anywhere. Recorded
//     reading: "an implementation MUST run every rule" (A2-9.4) means no row may
//     be omitted from the set, not that no row may be skipped once the decision
//     is taken — the table order fixes which id is reported, so scanning past a
//     match costs ten passes per field and cannot change the answer.
//   - S2 a rule id outside the table is not reportable: the ids are literals in
//     one slice, and the test compares them with a hand-transcribed list.
//   - S3 the message shape "secretscan.Scan: scanning field <field> for secret
//     material: rule <RULE-ID> matched: <len(value)> bytes (A2-9.4)" — an
//     attempt clause, then identifiers, then the clause cite, per ADR-0019 §2.
//     errs captures the "secretscan.Scan: " prefix itself (ADR-0019 §1), so
//     Scan's format string starts after it. Note the consequence: len(value) IS
//     information about the value — A2-9.5 permits exactly the byte length and
//     nothing else, and this package hands over no prefix, suffix, count of
//     symbols or entropy figure.
//   - S4 Scan never logs (log-or-return is ADR-0019 §3's last bullet; §4 is
//     granularity, not this). The caller that rejects the write logs the
//     returned error once.
//   - S5 reject, never redact: no redaction mode, no sanitized return value, no
//     bool found out-parameter. The value is also NOT passed through
//     errs.Redact — redaction presumes a value that is legitimate to log
//     elsewhere, and A2-9.5 forbids this one being reachable at all.
//   - S6 the planted corpus A2-9.4 assigns to the shared suite is a LOCAL
//     PROVISIONAL COPY in secretscan_test.go. Nothing here exports it; solving
//     the drift by exporting fixtures from the scanned package would let the
//     scanner own its own test oracle.
//
// # Notes for callers
//
// A2-9.5 covers three surfaces, not one: the message, an error attrs entry and
// any slog record. Scan controls the first and sets no attrs (it returns a bare
// errs error); the caller that builds an errs.Attrs or a slog record must not
// put the rejected value, or any prefix or digest of it, into engagement_id,
// run_id, job_id, node_id or a log attribute either.
//
// Scan is safe for concurrent use: the rule table is written once at package
// initialization and never mutated, [regexp.Regexp] is documented as
// concurrency-safe, and SEC-ENTROPY allocates its own counter array per call.
//
// A0-3.4 names its own tests for this rule, TestSecretScanNamesFieldNotValue
// and TestNoSecretValueOrDigestInError. Four of A2-9.4's five frozen test ids
// (its Tests: line lists five) are the ones this package implements, and they
// cover both A0-3.4 names as subtests of TestErrorMessageNamesFieldAndRuleIDOnly
// (exact_shape and the leak sweep). The fifth, TestGraphSecretFreeSerialization,
// is the shared suite's, assigned above.
// Whoever lands the shared suite (WP-20/WP-21) should either adopt those four
// names or add the two A0-3.4 names as thin aliases; the contract documents
// both sets and nothing reconciles them.
package secretscan
