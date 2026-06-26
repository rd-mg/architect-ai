<!-- architect-ai:architecture-guardrails:full-reference -->
# Architecture Guardrails (Full Reference)

## When to Apply
- Designing new feature or subsystem (D1 ≥ 2)
- Moving responsibilities across layers (Frontend, Backend, Database, Plugins)
- Evaluating changes to shared infrastructure
- Executing system-level mutations
- Reviewing PRs that touch base architecture

## Core Rules (5 — Inviolable)

### 1. Source of Truth
State lives in ONE place. No replication without sync.
- Explicitly define where actual state lives (DB, LocalStorage, Memory, Engram)
- Do not replicate state without a clear synchronization flow
- If a change affects data synchronization, test both "push" and "pull" paths

### 2. Thin Adapters
Business logic in domain/core. Integrations are thin wrappers.
- Keep integration layers (API/Plugins) as thin as possible
- External dependencies MUST be wrapped in adapters
- Never leak infrastructure concerns into domain logic

### 3. Explicit Boundaries
No hidden cross-system coupling in helpers/utilities.
- Respect separation of concerns (Composition over Inheritance)
- No hidden coupling within generic helpers or utilities
- Each module's public API should be its only contract

### 4. Mental Model First
Fit new features into logical model BEFORE designing implementation.
- New features must fit into the logical mental model BEFORE designing the UI
- Validate the mental model against existing architecture before committing
- If the model doesn't fit, refine the model — don't force-fit the feature

### 5. Sandbox Security
L2 agents CANNOT perform destructive mutations without L0/L1 authorization.
- L2 execution agents FORBIDDEN from raw destructive file system mutations (recursive deletes, mass permission changes) without L0/L1 authorization
- Critical terminal interactions and code execution MUST respect ephemeral isolation containers
- If operation requires breaking out of designated workspace: STOP, report RISK, defer to human

## Validation Protocol
- Add regression tests for EVERY change in system boundaries
- If change affects data synchronization, test both "push" and "pull" paths
- Verify no new hidden coupling introduced (grep for cross-module imports in helpers/utils)

## Compact Form (for L0 injection — ~150 tokens)

When injected into L0 or sub-agent Tier 0, use this compressed form:

```
Constitution: Honor all 5 rules:
1. Source of Truth — state in ONE place, no unsynced replication
2. Thin Adapters — business logic in core, integrations as wrappers
3. Explicit Boundaries — no hidden cross-system coupling
4. Mental Model First — validate model before implementation
5. Sandbox Security — L2 cannot destruct without L0/L1 auth
Report violations as RISK in return envelope.
```
