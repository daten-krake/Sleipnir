# ADR-0006: LLM integration — OpenAI-compatible, local-first

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

A key goal of Sleipnir is to **prove the platform works with local AI
(Qwen models)** — pentest data leaving the building is unacceptable for many
customers. At the same time the LLM layer must stay flexible.

## Decision

1. The platform talks to models through an **OpenAI-compatible chat
   completions API** (works for OpenAI itself and for local servers such as
   Ollama, vLLM, llama.cpp server, LMDeploy, etc.).
2. **Local-first:** the reference deployment uses self-hosted Qwen models;
   cloud APIs are an opt-in configuration.
3. Internally a small provider/model **registry** with a curated
   **supported-model list** (added later): model id, context size,
   capabilities, tool-calling support, recommended roles.
4. **Closed-door agent config:** agent definitions/prompts are locked while
   the platform is running a job — no live tampering of agent behavior
   mid-engagement (integrity + auditability).

## Consequences

- **Positive:** stdlib-only feasible (plain HTTP+JSON); local inference keeps
  data in-house; vendor-neutral.
- **Negative:** tool-calling quality varies across local models; Qwen-level
  function-calling support must be validated per model → supported-model
  list is a real gate, not decoration. Hand-rolled streaming JSON parsing.
- **Follow-ups:**
  - Verify Qwen tool-calling reliability for agent loops; define fallback
    (text-based protocol) if a model lacks native function calling.
  - Model role mapping (which model for orchestrator vs. workers) to control
    cost/latency on local hardware.
