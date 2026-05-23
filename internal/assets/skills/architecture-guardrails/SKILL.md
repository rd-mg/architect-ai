---
name: architecture-guardrails
description: >
  Global architecture guardrails.
  Trigger: Any change that affects system boundaries, state flow, or cross-package responsibilities.
metadata:
  domain: "Strategy"
  capabilities: ["System Boundaries", "State Flow Verification", "Coupling Prevention", "Execution Security"]
  version: "2.0"
---

# Architecture Guardrails (Global)

## Triggers
- Designing new feature or subsystem
- Moving responsibilities between Frontend, Backend, Database, or Plugins
- Evaluating PRs touching base architecture
- Executing system-level commands that mutate environment

## Core Guardrails (REQUIRED)
1. **Source of Truth**: Explicitly define where actual state lives (DB, LocalStorage, Memory). No state replication without clear synchronization flow.
2. **Thin Adapters**: Keep integration layers (API/Plugins) thin. Business logic in domain/core. External dependencies ALWAYS wrapped in adapters.
3. **Explicit Boundaries**: Composition over Inheritance. Respect separation of concerns.
4. **Mental Model First**: New features must fit logical mental model BEFORE designing UI.
5. **No Hidden Coupling**: NEVER hide cross-system coupling within generic helpers or utilities.

## Sandbox Execution Security (L1/L2 Delegation)
1. **Destructive Isolation:** L2 agents FORBIDDEN from raw destructive file system mutations (recursive deletes, mass permission changes) without L0/L1 authorization.
2. **SandboxDriver Abstraction:** All critical terminal interactions and code execution MUST respect ephemeral isolation containers.
3. **Graceful Failures:** If operation requires breaking out of designated workspace, agent MUST stop, report required permission escalation as `RISK`, defer to human operator.

## Validation Protocol
- Add regression tests for EVERY change in system boundaries
- If change affects data synchronization, test both "push" and "pull" paths
