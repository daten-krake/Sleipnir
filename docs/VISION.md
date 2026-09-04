# Sleipnir — Product Vision (working doc)

- **What:** Sleipnir is an automated, agent-driven pentest platform.
- **Users:** internal security teams and pentest consultants; self-deployable,
  single-tenant. SaaS is a possible much-later phase.
- **Primary target domain (v1):** **internal pentesting** — recon,
  services, credentials, lateral movement, privilege escalation on
  internal networks — with **Active Directory as the main emphasis**.
  Web application pentesting follows later.
- **Long-term vision:** evolve into a **purple team platform** (attack
  execution paired with detection validation).
- **Proof point:** the platform must run on **local AI (Qwen models)** via
  OpenAI-compatible endpoints — no data exfiltration required.
- **Test ground:** HackTheBox machines during development.
- **Differentiators:** tiered human approval for dangerous actions,
  full command/evidence log, and **revertable actions** (client environment
  cleanup), remote on-site agent node (e.g. Raspberry Pi).
