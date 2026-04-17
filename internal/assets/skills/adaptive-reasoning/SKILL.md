---
name: adaptive-reasoning
description: >
  Single-entry classifier and inline reasoning executor. Classifies a task
  across four observable dimensions and routes to one of three inline
  reasoning modes: direct-exec, adversarial-review (two-pass defect discovery),
  or bounded-synthesis (A/B/AB comparison for pre-implementation refinement).
  All reasoning executes INLINE — no delegation to sub-agents, no separate
  skills. This skill does NOT decide code acceptance (deterministic validators
  do) and does NOT own persistence.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "1.0"
---

# Adaptive Reasoning v1.0

## Why This Skill Exists

Previous versions of this framework shipped three separate skills:
`adaptive-reasoning` (classifier only), `judgment-day` (adversarial two-pass
review), and `autoreason-lite` (bounded A/B comparison). Using them together
required the orchestrator to delegate to external sub-agents, losing context
and consuming ~1500 extra tokens per invocation.

This v1.0 absorbs all three into a single skill. Classification and execution
happen inline in the same context. The three "modes" below are what used to
be three skills.

## Operating Contract

1. This skill **classifies and executes** reasoning inline.
2. It never delegates to sub-agents.
3. It never creates SDD phases.
4. It never decides code acceptance — deterministic validators (tests, builds,
   linters, `sdd-verify`) always outrank it for correctness claims.
5. It never owns persistence — phase skills persist artifacts.
6. Once classified, execute that mode fully in this same response.

## Classification: 4 Observable Dimensions

Score each dimension from strongest observable signal.

| Dimension | Values (weakest → strongest) |
|-----------|------------------------------|
| `scope` | atomic · bounded · multi-step · system-level |
| `ambiguity` | clear · partial · conflicting · unknown |
| `risk` | low · medium · high · critical |
| `verification` | syntax-only · testable · review-heavy · multi-gate |

Classification is mandatory and must score all four dimensions before routing.

## Routing Matrix

Apply these rules in order. Stop at the first match.

| Signal | Mode | Why |
|--------|------|-----|
| `verification=testable` or `verification=multi-gate` AND goal is code acceptance | **direct-exec → defer to validators** | Tests/builds/linters decide correctness; no reasoning overlay substitutes |
| `verification=review-heavy` or `verification=multi-gate` AND goal is DEFECT DISCOVERY | **Mode 2: adversarial-review** | Find what is wrong; two passes expose what one misses |
| `ambiguity` is `partial` or `conflicting` AND incumbent draft `A` exists AND competing draft `B` exists AND goal is pre-implementation refinement | **Mode 3: bounded-synthesis** | A/B/AB comparison produces the defensible final artifact |
| `scope=multi-step` or `scope=system-level` AND ambiguity is non-trivial | **native-sdd-first** | Decomposition before narrower reasoning modes |
| Everything else | **Mode 1: direct-exec** | No extra reasoning layer needed |

## Escalation Rules (priority order, stop at first match)

1. `verification=multi-gate` → always defer to deterministic validators
2. `risk=critical` → prefer Mode 2 or Mode 3 before Mode 1
3. `cost-constrained` context → prefer the simplest mode that preserves correctness

---

## Mode 1: direct-exec

**When to use**: atomic/bounded scope + clear ambiguity + low/medium risk + syntax-only/testable verification. Single credible approach. No synthesis or review needed.

**Action**: Proceed with the owning skill or phase. No extra reasoning overlay.

**Boundaries**: Do not use when defect discovery is the goal (use Mode 2). Do not use when two competing drafts exist (use Mode 3).

---

## Mode 2: adversarial-review (inline two-pass)

**When to use**: Review-heavy task. Goal is defect discovery in:
- code artifacts (PRs, patches, modules) with verification=review-heavy or multi-gate
- architecture artifacts (designs, SDDs, ADRs) looking for omissions/contradictions
- factual/research questions needing adversarial falsification

**Hard boundary**: DO NOT use for synthesis from competing drafts (that's Mode 3). DO NOT use for code correctness acceptance (deterministic validators do that).

### Procedure (execute inline, do not delegate)

**Step 1: Confirm target and scope**

If scope is ambiguous enough to invalidate the review, ask once for clarification. Otherwise proceed with the most defensible interpretation and state your assumption.

**Step 2: Run Pass A**

Build one serious analysis. Capture:
- Main conclusion
- Supporting evidence
- Key assumptions
- Critical reasoning steps
- Uncertainty or open gaps

Use the "local correctness lens" for code, "feasibility lens" for specs, "best-supported lens" for research.

**Step 3: Run Pass B (different lens)**

Use a materially different lens. Try to expose:
- Contradictory evidence
- Missing assumptions
- Broken causal links
- Overconfident claims
- Alternative explanations
- Unexamined edge cases

Use "system impact lens" for code, "failure-mode lens" for specs, "adversarial falsification" for research.

Keep passes independent — do NOT let Pass A bias Pass B.

**Step 4: Agreement Trap Check**

If both passes converge quickly, ask:
- What shared assumption could make both wrong?
- What evidence would overturn both?
- Did both inherit the same framing error?

If convergence is on weak basis, reduce confidence.

**Step 5: Synthesis**

Choose the final result by weighing:
1. Evidence quality
2. Reasoning quality
3. Contract/requirement alignment
4. Realism of failure scenarios
5. Remaining uncertainty

Options:
- Select Pass A if clearly stronger
- Select Pass B if clearly stronger
- Merge when both contribute complementary valid insight
- Synthesize a new result when both are incomplete but evidence supports better

Do NOT force symmetry when evidence is asymmetric.

**Step 6: Classify findings**

- **Confirmed** — both passes agree or one has strong evidence the other doesn't refute
- **Suspect** — raised by one pass only, not yet strongly evidenced
- **Contradiction** — passes materially disagree
- **Info** — notable but non-blocking

**Step 7: Apply severity**

- **CRITICAL** — production-breaking, security-relevant, corrupting, fundamentally incorrect
- **WARNING (real)** — realistic bug under normal use
- **WARNING (theoretical)** — requires contrived conditions
- **SUGGESTION** — improvement not required for correctness

Reality test: _Can a normal user, system state, or expected workflow trigger this without contrivance?_ If yes → real. If no → theoretical.

**Step 8: Verdict**

- **APPROVED** — no confirmed CRITICAL or WARNING (real) remain
- **CONDITIONALLY APPROVED** — only SUGGESTION and/or WARNING (theoretical) remain
- **NEEDS CHANGES** — confirmed CRITICAL or WARNING (real) remain
- **UNRESOLVED** — contradiction or missing evidence prevents conclusion

Verdict is ANALYTICAL only. Never present APPROVED as merge permission.

### Mode 2 Output Template

```markdown
## Adversarial Review — {target}

### Lens selection
- Pass A: {lens}
- Pass B: {lens}

### Findings

| Finding | Pass A | Pass B | Severity | Status |
|---------|--------|--------|----------|--------|
| {issue} | ✅ | ✅ | CRITICAL | Confirmed |
| {issue} | ✅ | ❌ | WARNING (real) | Suspect |

**Confirmed**: {count} · **Suspect**: {count} · **Contradictions**: {count}

### Key reasoning
- {why the strongest confirmed issue matters}

### Confidence
{high|medium|low} — {why}

### Verdict
{APPROVED|CONDITIONALLY APPROVED|NEEDS CHANGES|UNRESOLVED}
```

---

## Mode 3: bounded-synthesis (A/B/AB comparison)

**When to use**: Pre-implementation refinement. ALL FIVE conditions must hold:
1. Task is proposal/spec/design-class work (not implementation)
2. Incumbent draft `A` exists
3. At least one serious competing draft `B` exists
4. Goal is refinement, not open-ended ideation
5. Keeping `A` unchanged is still a valid outcome

If any condition fails, fall back to Mode 1.

**Hard boundary**: DO NOT use for implementation work (sdd-apply). DO NOT use for code acceptance. DO NOT use for defect discovery (that's Mode 2).

### Procedure (execute inline, do not delegate)

**Step 1: Confirm applicability**

Check all five conditions above. If any fails, route to Mode 1 and explain why.

**Step 2: Restate target**

State decision target, constraints, and success condition in one paragraph.

**Step 3: Normalize candidates**

- `A`: incumbent state/draft/approach
- `B`: one serious competing alternative
- Normalize both to the same comparison frame (same format, same evaluation criteria)

**Step 4: Produce synthesis candidate AB**

Create one synthesis combining the strongest material traits of A and B. Do not create tournaments or multiple synthesis candidates unless the user explicitly requests another round.

**Step 5: Evaluate A, B, AB against rubric**

For code/implementation targets:
- Correctness against stated behavior
- Safety of state transitions and side effects
- Contract compatibility and invariant preservation
- Blast radius, rollback difficulty, change surface
- Operability, readability, maintainability
- Simplicity relative to problem
- Testability

For non-code artifacts:
- Completeness against stated goal
- Clarity and testability of claims
- Internal consistency
- Alignment with approved constraints
- Simplicity relative to problem

Prefer material improvements over stylistic churn.

**Step 6: Apply conservative selection**

- Keep `A` if strongest overall
- Keep `A` on ties
- Prefer the option that introduces least unnecessary churn
- Adopt `B` or `AB` only when gain is substantive
- If `AB` wins, carry forward only material deltas

**Step 7: Return result**

### Mode 3 Output Template

```markdown
## Bounded Synthesis — {target}

### Applicability check
- [ ] Target is proposal/spec/design class
- [ ] Incumbent `A` exists
- [ ] Competing draft `B` exists
- [ ] Goal is pre-implementation refinement
- [ ] "No change" is acceptable

### Candidates
- **A (incumbent)**: {summary}
- **B (competing)**: {summary}
- **AB (synthesis)**: {summary}

### Comparison

| Criterion | A | B | AB |
|-----------|---|---|-----|
| {criterion} | {score/note} | {score/note} | {score/note} |

### Decision
{Keep A | Adopt B | Adopt AB}

### Minimal delta list
1. {change 1}
2. {change 2}

### Unresolved risks or assumptions
- {risk}
```

---

## Common Output Record

Every routing decision MUST emit this record:

```text
mode: direct-exec | adversarial-review | bounded-synthesis | native-sdd-first
scope: atomic | bounded | multi-step | system-level
ambiguity: clear | partial | conflicting | unknown
risk: low | medium | high | critical
verification: syntax-only | testable | review-heavy | multi-gate
confidence: high | medium | low
escalation_flag: none | risk-override | cost-constrained | multi-gate-override
reason: <1-2 lines>
```

---

## Boundaries Recap

- This skill does NOT create SDD phases.
- This skill does NOT replace `sdd-propose`, `sdd-spec`, `sdd-design`, or `sdd-verify`.
- This skill does NOT decide code acceptance (deterministic validators do).
- This skill does NOT own persistence of the final artifact.
- Modes 2 and 3 execute INLINE in the same response — never delegate, never spawn sub-agents.

## Anti-Patterns

- Using Mode 2 (adversarial-review) when the goal is synthesis, not defect discovery
- Using Mode 3 (bounded-synthesis) when only one credible draft exists
- Using Mode 3 for implementation or acceptance work
- Converting analytical uncertainty into false precision
- Letting both passes in Mode 2 share a hidden premise without checking it
- Generating many alternatives in Mode 3 just because ambiguity exists
- Preferring novelty over stability in Mode 3
- Claiming Mode 2 verdict = merge permission

## Compatibility with SDD

- If `adaptive-reasoning` is invoked before delegation, the skill resolver MAY consume the routing record directly to inform delegation choice.
- Routing record is stable contract surface. Field names and mode values are not renamed.

## Legacy Note

Previous standalone skills `judgment-day` and `autoreason-lite` are archived
under `internal/assets/skills/_archived/`. Their logic is absorbed here. If
you see references to them in legacy prompts, treat them as equivalent to
Mode 2 and Mode 3 of this skill respectively.
