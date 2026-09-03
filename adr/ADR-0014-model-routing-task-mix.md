# ADR-0014: Model routing — task-based model mix, staged providers

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Different agent roles have different needs (orchestration reasoning vs. fast
tool execution vs. summarization). Operator currently runs LM Studio locally
but wants to **start with Qwen Cloud** and use a **mix of models per task**,
moving to local inference as the proof point later (ADR-0006).

## Decision

1. Model configuration is **per-role/per-task**: each agent profile declares
   which model it uses (e.g. strong reasoning model for the orchestrator,
   small fast models for routine worker tasks, capable model for research).
2. Providers are staged: **Qwen Cloud first** (OpenAI-compatible endpoint),
   **LM Studio / other local servers later** — no code change expected
   beyond endpoint/model configuration, since everything speaks the
   OpenAI-compatible contract (ADR-0006).
3. The supported-model list (ADR-0006) records per model: context size,
   tool-calling reliability, recommended roles.

## Consequences

- **Positive:** cost/latency control; cloud → local migration is config, not
  code; matches how the operator already works.
- **Negative:** behavior varies with model mix → findings reproducibility
  depends on recording the exact model config per run (must land in the
  event log).
- **Follow-ups:** default model-role matrix for v1; per-run model snapshot
  in evidence.
