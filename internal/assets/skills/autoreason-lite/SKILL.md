---
name: autoreason-lite
description: bounded competitive refinement for code and non-code artifacts using an incumbent state or draft, one plausible competing alternative, and one synthesis candidate. use when comparing or refining an implementation approach, patch, refactor, specification, requirement, or technical design with meaningful tradeoffs, especially when keeping the current state unchanged must remain a valid outcome. this skill defines how the agent should think; it does not by itself authorize edits, refactors, commits, pushes, workflow transitions, phase transitions, or code acceptance, and deterministic validators outrank it for runtime decisions.
---

# Autoreason Lite

## Overview

Use this skill to structure reasoning for one bounded comparison round over a target that may be code or non-code.

Treat it as a support protocol, not as an artifact owner, workflow owner, router, validator, execution phase, or gate. It may assist work that happens inside SDD-like flows, but it must not assume that such a flow exists and it must not create, rename, or replace phases.

## Operating Contract

Apply these rules first:

1. Confirm that the target is concrete enough to compare meaningfully.
2. Confirm that there is an incumbent `A` and one plausible competing alternative `B`.
3. Confirm that the task involves real ambiguity, tradeoffs, or competing structures.
4. Confirm that the goal is refinement, remediation planning, or conservative solution selection rather than open-ended ideation.

If any condition is false, do not use this skill. Fall back to the host workflow or the agent's normal reasoning.

## Hard Boundaries

Never use this skill to do any of the following:

- replace deterministic acceptance evidence
- treat tests, builds, linters, parsers, AST checks, type checks, or compilers as subordinate to preference-based reasoning
- declare merge approval, release approval, or workflow advancement unless the user explicitly asks for that interpretation
- claim ownership of persistence, task decomposition, routing, or phase control
- expand into unbounded brainstorming tournaments

If deterministic validators exist, they outrank this skill for runtime or code decisions.

## Target Types

This skill may be used for any of these targets when bounded comparison is the right reasoning shape:

- current code versus a competing fix or refactor
- an existing patch versus an alternative patch
- an incumbent implementation approach versus a different one
- a proposal, specification, requirement set, task plan, or technical design

Do not use this skill for broad adversarial defect-hunting when there is no real competing option to compare. Use a review-oriented method for that.

## Candidate Model

Use exactly one bounded comparison round with these candidates:

- `A`: incumbent state, artifact, implementation, patch, or approach
- `B`: one serious competing alternative that challenges omissions, structure, sequencing, scope, constraints, correctness, or risk distribution
- `AB`: one synthesis candidate that preserves the strongest material traits of `A` and `B`

Do not create tournaments, brackets, or repeated generations of alternatives unless the user explicitly asks for another round.

## Evaluation Rubric

Compare `A`, `B`, and `AB` against the smallest useful set of criteria for the target.

For code or implementation problems, prioritize:

- correctness against the stated behavior
- safety of state transitions, side effects, and failure handling
- contract compatibility and invariant preservation
- blast radius, rollback difficulty, and change surface
- operability, readability, and maintainability
- simplicity relative to the problem being solved
- testability and validator alignment

For non-code artifacts, prioritize:

- completeness against the stated goal
- clarity and testability of the claims
- internal consistency and absence of contradiction
- alignment with already-approved constraints and neighboring artifacts
- blast radius, rollback difficulty, and change surface
- simplicity relative to the problem being solved

Prefer material improvements over stylistic churn.

## Conservative Selection Rule

Treat "no change" as a valid and often preferable outcome.

- Keep `A` if it is already strongest overall.
- Keep `A` on ties.
- Prefer the option that introduces the least unnecessary churn.
- Adopt `B` or `AB` only when the gain is substantive, not cosmetic.
- If `AB` wins, carry forward only the material deltas that improve the target.

## Comparison Procedure

Follow this sequence:

1. Restate the decision target, constraints, and success condition.
2. Normalize `A` and `B` to the same comparison frame.
3. Produce one synthesis candidate `AB`.
4. Compare `A`, `B`, and `AB` using the rubric above.
5. Select one winner using the conservative selection rule.
6. Return the chosen option plus a brief rationale and explicit delta summary.

Do not let the comparison expand into open-ended brainstorming.

## Action Policy

By default, this skill produces **analysis first**.
It does **not** authorize edits, refactors, commits, pushes, workflow transitions, or gate decisions by itself.

If the user explicitly asks for remediation after analysis:

1. address confirmed `CRITICAL` findings first
2. then address confirmed `WARNING (real)` findings
3. treat `WARNING (theoretical)` as informational unless hardening is requested
4. handle `SUGGESTION` items only if trivial or explicitly requested

If a fix addresses a repeated pattern, check the same pattern across the analyzed scope so the correction is consistent.

## Re-review Policy

Re-review is optional and context-driven.
Use it when:

- confirmed critical issues were fixed
- confirmed real warnings were fixed in paths with ripple effects
- the user explicitly asks for a second pass
- the initial synthesis had low confidence

Do not require a fixed number of rounds.
Do not force escalation loops.
Do not bind the review to merge gates, release gates, or phase gates unless the user explicitly asks for that interpretation.

## Output Contract

When surfacing results, organize them in this order:

1. applicability check
2. summary of `A`, `B`, and `AB`
3. comparison findings by criterion
4. decision: keep `A`, adopt `B`, or adopt `AB`
5. minimal delta list
6. unresolved risks or assumptions

Keep the explanation concise and decision-oriented.

## Relationship To Other Systems

This skill only shapes how the agent thinks.

- If a host workflow has named SDD phases, let that workflow keep phase ownership.
- If a host workflow has artifact owners, let those owners keep authorship and persistence.
- If an adaptive router exists, let it decide when to invoke this skill; do not assume one exists.
- If a defect-review skill exists, reserve that skill for adversarial review rather than bounded synthesis.

## Anti-Patterns

Avoid these failure modes:

- using this skill when there is no real competing alternative
- generating many alternatives just because ambiguity exists
- preferring novelty over stability
- rewriting accepted structure without material gain
- using reasoning preference as a substitute for deterministic code evidence
- turning an analytical decision into an implicit merge, release, or phase gate
