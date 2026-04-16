---
name: adaptive-reasoning
description: >
  First-pass task classification and method routing for Architect-AI.
  Trigger: When a task needs an explicit choice between direct execution,
  deterministic validation, adversarial review, or bounded artifact refinement.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "0.3"
---

## When to Use

- User explicitly asks for "adaptive-reasoning", "adaptive-reasoning", or equivalent trigger phrases

---

## Purpose

`adaptive-reasoning` is the classifier, not the finisher.

Its only job is to select one fixed route from the routing vocabulary:

- `native-owner`
- `deterministic-validators`
- `judgment-day`
- `autoreason-lite`
- `native-sdd-first`

It MUST NOT approve implementation correctness and MUST NOT replace deterministic acceptance.

This skill shapes **how the agent thinks about a task** — it sets the reasoning posture before action,
regardless of which SDD phase the agent is or is not currently inside. It does not coordinate phases
at runtime and does not require phase-state awareness to operate.

---

## Routing Principles

### Pattern 1: Classify Before You Optimize

Decide what kind of problem this is before choosing a reasoning overlay.

Task classes:
- implementation task
- deterministic acceptance task
- defect-finding review task
- artifact-refinement task

If the task is already clear and owned by a native skill, keep the native path.

Classification is mandatory and must evaluate all six dimensions.

### Pattern 2: Keep Deterministic Acceptance Sacred

Do NOT route code acceptance to a reasoning overlay.

- tests
- builds
- lints
- structural validators
- `sdd-verify` acceptance

These always route to `deterministic-validators`. No override exists for this rule.

### Pattern 3: Route to `autoreason-lite` Only When All Five Conditions Are Met

Route to `autoreason-lite` if and only if **all** of the following are true:

1. The task is proposal/spec/design-class work.
2. There is an incumbent draft `A`.
3. There is at least one serious competing draft `B`.
4. The goal is pre-implementation refinement.
5. Keeping the incumbent unchanged is still an acceptable outcome.

If **any** condition is absent, stay with the owning SDD skill (`native-owner` or `native-sdd-first`
depending on scope). Do not route to `autoreason-lite` with four of five conditions — treat partial
satisfaction as a miss, not a partial pass.

### Pattern 4: Use `judgment-day` for Defects, Not Synthesis

Route to `judgment-day` when the goal is discovery — finding bugs, omissions, or regressions in:

- **code artifacts** (PRs, patches, modules) — verification burden is `review-heavy` or `multi-gate`.
- **architecture artifacts** (design docs, SDDs, ADRs) — but only when looking for _omissions and
  contradictions_, not when synthesizing a stronger design from competing options.

Do not use `judgment-day` for:
- Synthesis or comparison of competing proposal drafts → use `autoreason-lite`.
- Code correctness acceptance → use `deterministic-validators`.

The distinguishing question: _Is the goal to find what is wrong, or to produce something better?_
Wrong-finding → `judgment-day`. Better-producing → `autoreason-lite`.

### Pattern 5: Use Observable Features, Not Vibes

Classify from observable signals. Score every dimension before choosing a route.

| Dimension | Values (weakest → strongest) |
|-----------|------------------------------|
| scope size | `atomic` → `bounded` → `multi-step` → `system-level` |
| ambiguity | `clear` → `partial` → `conflicting` → `unknown` |
| dependency shape | `standalone` → `sequential` → `branching` → `graph-shaped` |
| risk level | `low` → `medium` → `high` → `critical` |
| verification burden | `syntax-only` → `testable` → `review-heavy` → `multi-gate` |
| cost sensitivity | `low-cost-ok` → `balanced` → `cost-constrained` |

Do not route from trigger words alone. Score first.

---

## Computed Classification Matrix

### Step 1: Classify the Task

Use the strongest matching value in each dimension.

| Dimension | Low side | High side | What to inspect |
|-----------|----------|-----------|-----------------|
| Scope size | `atomic` | `system-level` | files, moving parts, or artifacts touched |
| Ambiguity | `clear` | `unknown` | whether the request defines outcome and constraints |
| Dependency shape | `standalone` | `graph-shaped` | whether work fans across phases, branches, or subsystems |
| Risk level | `low` | `critical` | blast radius, rollback pain, policy or user sensitivity |
| Verification burden | `syntax-only` | `multi-gate` | whether machine checks, review, or multiple validators are needed |
| Cost sensitivity | `low-cost-ok` | `cost-constrained` | whether heavier reasoning is justified |

### Step 2: Derive the Method Family

| Classification signal | Route | Why |
|-----------------------|-------|-----|
| `atomic` + `clear` + `standalone` + `low` | `native-owner` | Direct execution is cheaper and clearer |
| implementation work with `testable` or `multi-gate` verification | `deterministic-validators` | Machine-checkable acceptance must decide |
| review-heavy task where defect discovery in code or architecture matters more than synthesis | `judgment-day` | Independent adversarial review is the right tool |
| proposal/spec/design work with `partial` or `conflicting` ambiguity and at least two plausible drafts | `autoreason-lite` | Bounded synthesis improves planning artifacts without replacing the owner |
| `multi-step` or `system-level` change with non-trivial ambiguity | `native-sdd-first` | Planning/decomposition must stabilize before choosing a narrower overlay |
| `bounded` scope with a clear owner and no competing drafts | `native-owner` | No extra routing layer needed |

`bounded` is a valid and common outcome. It resolves to `native-owner` unless another signal overrides.

### Step 3: Apply Escalation Rules

Escalation rules are applied after the base route is derived. When rules conflict, the priority order is:

1. `verification burden = multi-gate` → deterministic validators outrank any judge-style preference.
2. `risk = critical` → prefer explicit planning or adversarial review over direct execution.
3. `cost sensitivity = cost-constrained` → prefer the simplest route that still preserves correctness.

Rule 1 beats Rule 2. Rule 2 beats Rule 3. Do not apply all three independently and then pick — apply in
order and stop at the first match that changes the base route.

---

## Route Specification

### `native-owner`

Direct execution by the owning skill or phase. No extra routing overlay.

Use when:
- The task is a stable implementation request with a clear owner.
- Scope is `atomic` or `bounded`, ambiguity is `clear`, verification is `syntax-only` or `testable`.
- Only one credible draft exists and no review or synthesis is needed.

Do not use when deterministic acceptance, defect-finding, or system-level decomposition is required.

### `deterministic-validators`

Machine-checkable evidence decides. This route is the terminal for all code acceptance work.

Use when:
- Tests, builds, lints, structural validators, or `sdd-verify` acceptance are the acceptance gate.
- Verification burden is `testable` or `multi-gate`.

No reasoning overlay can substitute for this route on code acceptance. Multi-gate verification burden
always escalates here, regardless of other signals.

### `judgment-day`

Independent adversarial review. The goal is discovery of defects, omissions, or regressions.

Use when:
- The task is a PR review, architecture audit, or SDD review where defect-finding is the primary goal.
- Verification burden is `review-heavy` or higher.
- The artifact under review is either code or an architecture document (not a proposal being refined).

Do not use when:
- The goal is to synthesize a stronger artifact from competing drafts.
- The task is code correctness acceptance.

### `autoreason-lite`

Bounded artifact synthesis with incumbency protection. Refines planning artifacts pre-implementation.

Use when all five Pattern 3 conditions are met (see above).

Do not use when:
- Only one credible draft exists.
- The work is implementation (`sdd-apply`) or acceptance (`sdd-verify`).
- The goal is defect-finding rather than synthesis.

### `native-sdd-first`

Planning and decomposition before method selection. Stabilizes scope before narrower overlays are chosen.

Use when:
- Scope is `multi-step` or `system-level`.
- Ambiguity is `partial`, `conflicting`, or `unknown`.
- The task fans across phases, subsystems, or branches.

After `native-sdd-first` produces a stable decomposition, re-classify each resulting sub-task
independently. Some may route to `native-owner`, others to `deterministic-validators`, and so on.

Do not use `native-sdd-first` as a catch-all for ambiguous tasks that are actually `bounded` in scope.
Score the scope dimension first.

---

## Routing Record (Required Output)

Every routing decision must emit this compact record:

```text
owner: <native owner or capability>
scope: atomic|bounded|multi-step|system-level
ambiguity: clear|partial|conflicting|unknown
dependency_shape: standalone|sequential|branching|graph-shaped
risk: low|medium|high|critical
verification_burden: syntax-only|testable|review-heavy|multi-gate
cost_sensitivity: low-cost-ok|balanced|cost-constrained
route: native-owner|deterministic-validators|judgment-day|autoreason-lite|native-sdd-first
confidence: high|medium|low
escalation_flag: none|risk-override|cost-constrained|multi-gate-override
reason: <1-2 lines>
```

`confidence` reflects how cleanly the task matched the classification matrix:
- `high` — all dimensions scored clearly, route is unambiguous.
- `medium` — one dimension was uncertain; route is the strongest match.
- `low` — multiple dimensions conflicted or were unknown; route is provisional and should be reviewed.

`escalation_flag` records whether an escalation rule changed the base route:
- `none` — base route was kept.
- `risk-override` — Rule 2 (critical risk) changed the route.
- `cost-constrained` — Rule 3 constrained the route to a simpler option.
- `multi-gate-override` — Rule 1 (multi-gate verification) overrode a judge-style preference.

The field names and route values are stable contract surface.

---

## Route Examples

### `judgment-day` positive — code review

- Input: large PR review where the goal is defect-finding and regression detection.
- Route: `judgment-day`.
- Reason: review-heavy verification burden on a code artifact; discovery, not synthesis.

### `judgment-day` positive — architecture audit

- Input: reviewing a finalized SDD for contradictions and missing failure modes.
- Route: `judgment-day`.
- Reason: architecture artifact under adversarial review for omissions; not a proposal refinement.

### `judgment-day` blocked — competing proposals

- Input: choosing between two competing design proposals.
- Route: not `judgment-day`.
- Why blocked: the goal is synthesis/refinement, not defect discovery. Use `autoreason-lite` if both
  Pattern 3 conditions for incumbency and competition are met.

### `autoreason-lite` positive

- Input: proposal refinement with incumbent draft `A` and serious competing draft `B`.
  All five Pattern 3 conditions are satisfied.
- Route: `autoreason-lite`.
- Reason: bounded artifact synthesis before implementation; incumbency protection applies.

### `autoreason-lite` blocked — single draft

- Input: one proposal draft with no real competitor.
- Route: not `autoreason-lite`.
- Why blocked: no bounded comparison target exists. Route to `native-owner`.

### `autoreason-lite` blocked — code work

- Input: `sdd-apply` implementation or `sdd-verify` acceptance.
- Route: not `autoreason-lite`.
- Why blocked: code execution and verification must converge through deterministic validators.

### `native-sdd-first` positive

- Input: system-wide refactor touching six subsystems; scope is unclear and dependencies are
  graph-shaped.
- Route: `native-sdd-first`.
- Reason: decomposition required before any narrower overlay can be selected.

### `native-sdd-first` blocked — bounded scope

- Input: single-module change with partial ambiguity around one edge case.
- Route: not `native-sdd-first`. Score scope as `bounded`, not `system-level`.
- Why blocked: `native-sdd-first` is for system-level decomposition, not for hedging on
  bounded tasks with minor uncertainty.

### `bounded` scope resolves to `native-owner`

- Input: update a single config file to add a new environment variable; owner is known; no competing
  options; verification is syntax-only.
- Scope: `bounded`. Route: `native-owner`.
- Reason: bounded scope with a clear owner; no extra layer needed.

### Escalation conflict example

- Input: critical-risk task that is also cost-constrained and requires multi-gate verification.
- Apply escalation rules in priority order:
  1. `multi-gate-override` fires first → route to `deterministic-validators`.
  2. Rules 2 and 3 are not evaluated further.
- `escalation_flag: multi-gate-override`.

---

## Minimal Workflow

1. Identify the owning phase or capability.
2. Score all six observable dimensions: scope, ambiguity, dependency shape, risk, verification burden,
   and cost sensitivity.
3. Determine the task class: implementation, acceptance, defect-finding, or artifact-refinement.
4. Check whether deterministic evidence must decide (verification burden `testable` or `multi-gate`).
5. Apply the Step 2 matrix to derive a base route.
6. Apply escalation rules in priority order (multi-gate → critical risk → cost-constrained). Stop at
   the first rule that changes the route.
7. Set `confidence` and `escalation_flag`.
8. Emit the routing record.

---

## Boundaries

- This skill does not create a new SDD phase.
- This skill does not replace `sdd-propose`, `sdd-spec`, `sdd-design`, or `sdd-verify`.
- This skill does not decide code acceptance.
- This skill does not own persistence of the final artifact.

---

## Resources

- `internal/assets/skills/autoreason-lite/SKILL.md`
- `internal/assets/skills/judgment-day/SKILL.md`