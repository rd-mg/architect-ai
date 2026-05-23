---
name: autoreason-lite
description: bounded competitive refinement for code and non-code artifacts using an incumbent state or draft, one plausible competing alternative, and one synthesis candidate. use when comparing or refining an implementation approach, patch, refactor, specification, requirement, or technical design with meaningful tradeoffs, especially when keeping the current state unchanged must remain a valid outcome. this skill defines how the agent should think; it does not by itself authorize edits, refactors, commits, pushes, workflow transitions, phase transitions, or code acceptance, and deterministic validators outrank it for runtime decisions.
---

# Autoreason Lite

## Overview

Use to structure reasoning for one bounded comparison round over a target (code or non-code).

Support protocol, not artifact owner, workflow owner, router, validator, execution phase, or gate. May assist inside SDD-like flows, but must not assume such flow exists, must not create/rename/replace phases.

## Operating Contract

Apply these rules first:

1. Confirm target concrete enough to compare meaningfully.
2. Confirm incumbent `A` and one plausible competing alternative `B`.
3. Confirm task involves real ambiguity, tradeoffs, or competing structures.
4. Confirm goal is refinement, remediation planning, or conservative solution selection rather than open-ended ideation.

If any condition false, do not use. Fall back to host workflow or agent's normal reasoning.

## Hard Boundaries

Never use to:
- replace deterministic acceptance evidence
- treat tests, builds, linters, parsers, AST checks, type checks, compilers as subordinate to preference-based reasoning
- declare merge approval, release approval, workflow advancement unless user explicitly asks
- claim ownership of persistence, task decomposition, routing, or phase control
- expand into unbounded brainstorming tournaments

Deterministic validators outrank this skill for runtime/code decisions.

## Target Types

May use for any of these targets when bounded comparison is right shape:
- current code vs competing fix/refactor
- existing patch vs alternative patch
- incumbent approach vs different one
- proposal, specification, requirement set, task plan, technical design

Do NOT use for broad adversarial defect-hunting when no real competing option. Use review-oriented method instead.

## Candidate Model

Exactly one bounded comparison round:

- `A`: incumbent state, artifact, implementation, patch, or approach
- `B`: one serious competing alternative that challenges omissions, structure, sequencing, scope, constraints, correctness, or risk distribution
- `AB`: one synthesis candidate preserving strongest material traits of `A` and `B`

No tournaments, brackets, or repeated generations unless user explicitly asks.

## Evaluation Rubric

Compare `A`, `B`, `AB` against smallest useful criteria set.

For code/implementation:
- correctness against stated behavior
- safety of state transitions, side effects, failure handling
- contract compatibility and invariant preservation
- blast radius, rollback difficulty, change surface
- operability, readability, maintainability
- simplicity relative to problem
- testability and validator alignment

For non-code artifacts:
- completeness against stated goal
- clarity and testability of claims
- internal consistency, absence of contradiction
- alignment with approved constraints, neighboring artifacts
- blast radius, rollback difficulty, change surface
- simplicity relative to problem

Prefer material improvements over stylistic churn.

## Conservative Selection Rule

"No change" is valid and often preferable.

- Keep `A` if already strongest overall.
- Keep `A` on ties.
- Prefer option introducing least unnecessary churn.
- Adopt `B` or `AB` only when gain is substantive, not cosmetic.
- If `AB` wins, carry forward only material deltas improving the target.

## Comparison Procedure

1. Restate decision target, constraints, success condition.
2. Normalize `A` and `B` to same comparison frame.
3. Produce one synthesis candidate `AB`.
4. Compare `A`, `B`, `AB` using rubric above.
5. Select one winner using conservative selection rule.
6. Return chosen option + brief rationale + explicit delta summary.

Do not expand into open-ended brainstorming.

## Action Policy

By default, produces **analysis first**. Does **not** authorize edits, refactors, commits, pushes, workflow transitions, or gate decisions.

If user explicitly asks for remediation:

1. address confirmed `CRITICAL` findings first
2. then address confirmed `WARNING (real)` findings
3. treat `WARNING (theoretical)` as informational unless hardening requested
4. handle `SUGGESTION` items only if trivial or explicitly requested

If fix addresses repeated pattern, check same pattern across analyzed scope for consistency.

## Re-review Policy

Re-review is optional, context-driven. Use when:
- confirmed critical issues fixed
- confirmed real warnings fixed in paths with ripple effects
- user explicitly asks for second pass
- initial synthesis had low confidence

No fixed number of rounds. No forced escalation loops. No binding to merge/release/phase gates unless user explicitly asks.

## Output Contract

Organize results in this order:

1. applicability check
2. summary of `A`, `B`, `AB`
3. comparison findings by criterion
4. decision: keep `A`, adopt `B`, or adopt `AB`
5. minimal delta list
6. unresolved risks or assumptions

Keep explanation concise and decision-oriented.

## Relationship To Other Systems

Shapes only how agent thinks.

- Host workflow with named SDD phases → let workflow keep phase ownership.
- Host workflow with artifact owners → let owners keep authorship/persistence.
- Adaptive router exists → let it decide when to invoke; don't assume one exists.
- Defect-review skill exists → reserve for adversarial review, not bounded synthesis.

## Anti-Patterns

Avoid:
- using when no real competing alternative
- generating many alternatives just because ambiguity exists
- preferring novelty over stability
- rewriting accepted structure without material gain
- using reasoning preference as substitute for deterministic code evidence
- turning analytical decision into implicit merge/release/phase gate
