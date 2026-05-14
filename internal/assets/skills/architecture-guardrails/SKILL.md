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

## When to Use
- Designing a new feature or subsystem.
- Moving responsibilities between Frontend, Backend, Database, or Plugins.
- Evaluating PRs that touch the base architecture.
- Executing system-level commands that mutate the environment.

## Core Guardrails (REQUIRED)
1. **Source of Truth**: Explicitly define where the actual state lives (e.g., DB, LocalStorage, Memory). Do not replicate state without a clear synchronization flow.
2. **Thin Adapters**: Keep integration layers (API/Plugins) as thin as possible. Business logic lives in the domain/core. External dependencies must ALWAYS be wrapped in adapters.
3. **Explicit Boundaries**: Respect the separation of concerns (Composition over Inheritance).
4. **Mental Model First**: New features must fit into the logical mental model BEFORE designing the UI.
5. **No Hidden Coupling**: NEVER hide cross-system coupling within generic helpers or utilities.

## Sandbox Execution Security (L1/L2 Delegation)
1. **Destructive Isolation:** L2 execution agents are FORBIDDEN from performing raw destructive file system mutations (e.g., recursive deletes, mass permission changes) without L0/L1 Orchestrator authorization.
2. **SandboxDriver Abstraction:** All critical terminal interactions and code execution routines MUST respect the system's ephemeral isolation containers.
3. **Graceful Failures:** If an operation requires breaking out of the designated workspace, the agent MUST stop, report the required permission escalation as a `RISK`, and defer to the human operator.

## Validation Protocol
- Add regression tests for EVERY change in system boundaries.
- If the change affects data synchronization, test both "push" and "pull" paths.
