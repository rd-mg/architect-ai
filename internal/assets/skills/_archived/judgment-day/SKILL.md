---
name: judgment-day
description: perform adaptive two-pass adversarial reasoning and synthesis for code, architecture, diffs, specifications, implementation plans, research questions, and technical decisions. use when the user asks for judgment day, judgment-day, dual review, adversarial review, two-judge reasoning, reasoning court, courteval, doble review, juzgar, or when a single-pass answer may miss hidden flaws. this skill defines how the agent should think, compare competing hypotheses, evaluate evidence quality and reasoning trajectories, and produce an analytical verdict without assuming any specific sdd phase, workflow, tool, or execution environment.
---

# Judgment Day

## Overview

Run disciplined adversarial analysis through **two independent reasoning passes**, then synthesize into one decision-ready result.

Defines **how to think**, not how to orchestrate work. Compatible with both **sdd** and **non-sdd** contexts. Must not impose lifecycle, tooling, repository, or delivery assumptions not already present in the task.

## Core Principle

Three stages:

1. **Pass A** — build one serious interpretation, answer, or critique
2. **Pass B** — build materially different interpretation, answer, or critique
3. **Synthesis** — compare both passes, audit evidence and reasoning paths, produce most defensible final result

Goal: create structured disagreement exposing hidden errors, weak evidence, missing assumptions, premature convergence. Not roleplay.

## Independence Rule

Keep passes independent until synthesis.

- Do not let Pass A bias Pass B before both complete.
- Do not reuse same framing sentence-for-sentence across both passes.
- When tools/retrieval available, prefer different search angles, candidate explanations, or falsification checks.
- When tools not available, still keep separate notes per pass until synthesis.

Sub-agents/delegation/parallel execution may be used if available. If not, perform both passes sequentially while preserving independence.

## Scope Compatibility

Must work whether target is in:
- sdd or non-sdd work
- discovery, design, implementation, refactor, review, debugging, research, or incident response
- code, diffs, specs, plans, migrations, tests, architecture, procedures, or factual questions

Do **not** assume:
- specific sdd phase exists
- registry, memory system, resolver exists
- sub-agents, delegate, delegation_read, mem_search, git, pr state, or branch state available
- user wants edits, fixes, commits, pushes, or status changes
- task is necessarily code review (may be reasoning, evaluation, or answer synthesis)

If project standards, architecture rules, contracts, or phase-specific constraints available in context, apply them. Otherwise proceed with generic engineering/analytical judgment, state assumption explicitly.

## Adaptive Lens Selection

Choose passes to create **meaningful tension**, not superficial duplication.

### For code, diffs, bugs, tests, implementations
- **Pass A — local correctness lens**: correctness, edge cases, error handling, safety, state transitions, direct implementation risk.
- **Pass B — system impact lens**: contracts, integration assumptions, invariants, operability, performance, maintainability, downstream effects.

### For architecture, specs, implementation plans
- **Pass A — feasibility lens**: internal coherence, completeness, assumptions, sequencing, whether plan can work as written.
- **Pass B — failure-mode lens**: hidden coupling, migration hazards, rollback gaps, scalability limits, observability gaps, where design breaks under realistic stress.

### For factual/research questions
- **Pass A — best-supported answer lens**: strongest answer supported by available evidence.
- **Pass B — adversarial falsification lens**: disprove, narrow, or qualify Pass A by identifying contradictory evidence, missing links, alternative interpretations, unsupported jumps.

### For policy/process/operational guidance
- **Pass A — rule and compliance lens**: what stated policy/contract/process explicitly requires.
- **Pass B — execution reality lens**: ambiguity, edge conditions, operator behavior, practical failure points, how rule behaves in real workflows.

### For ambiguous targets
If target spans multiple categories, choose pair producing strongest disagreement surface, state chosen lenses explicitly.

## Pass Procedure

### 1. Confirm target
Identify actual scope first. If scope ambiguous enough to invalidate review, ask once. Otherwise proceed with most defensible bounded interpretation, state assumption.

### 2. Run Pass A
Produce serious analysis, not strawman. Capture: main conclusion, supporting evidence, key assumptions, critical reasoning steps, uncertainty/open gaps.

### 3. Run Pass B
Use materially different lens. Try to expose: contradictory evidence, missing assumptions, broken causal links, overconfident claims, alternative explanations, unexamined edge cases, invalid generalization from narrow evidence.

### 4. Evaluate trajectories, not just outputs
Inspect **reasoning trajectory** of each pass:
- Evidence relevant?
- Key assumptions explicit?
- Skip necessary inference step?
- Confuse correlation with causation?
- Rely on stale/weak/indirect support?
- Ignore known contract/invariant/stated requirement?
- Collapse uncertainty into unjustified definitive claim?

Pass with polished answer but weak trajectory should lose to rougher answer with stronger evidence and cleaner logic.

## Agreement Trap Check

If both passes converge quickly, do **not** assume correctness. Run anti-convergence check:

- What shared assumption could make both passes wrong?
- What evidence would most likely overturn both?
- Did both passes inherit same framing error from prompt/context?
- Are both repeating same unsupported claim in different words?

When both passes agree on weak basis, reduce confidence and state so.

## Synthesis Rule

Not majority voting. Choose final result by weighing:

1. evidence quality
2. reasoning quality
3. contract and requirement alignment
4. realism of failure scenarios
5. remaining uncertainty

Synthesizer may:
- **Select Pass A** when clearly stronger
- **Select Pass B** when clearly stronger
- **Merge both** when each contributes valid, complementary insight
- **Synthesize new result** when both incomplete/partially wrong but evidence supports better conclusion

Do not pretend both equally valid when one materially stronger. Do not force symmetry when evidence asymmetric.

## Finding Buckets

When task is review-oriented:

- **Confirmed** — both passes agree, or one has strong direct evidence and other does not materially refute
- **Suspect** — raised by one pass only, not yet strongly evidenced
- **Contradiction** — both passes materially disagree on same claim
- **Info** — notable but non-blocking observations

## Severity Model

Classify each finding as:

- **CRITICAL** — likely production-breaking, security-relevant, corrupting, or fundamentally incorrect
- **WARNING (real)** — realistic bug or operational problem under normal use
- **WARNING (theoretical)** — requires contrived, unsupported, or highly artificial conditions
- **SUGGESTION** — useful improvement but not required for correctness

### Reality test for warnings
> Can a normal user, system state, or expected workflow trigger this without contrivance?

- **yes** → `WARNING (real)`
- **no** → `WARNING (theoretical)`

Treat theoretical warnings as informational by default unless user explicitly asks for hardening.

## Confidence Rule

Communicate confidence based on evidence quality, not tone:

- **high confidence** — direct evidence, coherent trajectory, no material contradiction
- **medium confidence** — plausible conclusion with limited or indirect evidence
- **low confidence** — unresolved contradiction, missing evidence, strong dependence on assumptions

Never present confidence as certainty.

## Evidence Rule

Every non-trivial claim should include strongest available support:
- file and line, when available
- exact contract/invariant/requirement involved
- concrete failure scenario or falsification path
- why it matters in practice
- smallest safe correction direction, when remediation requested

Do not inflate severity without evidence. Do not downgrade real risk because output looks clean/familiar.

## Action Policy

By default, produces **analysis first**. Does **not** authorize edits, refactors, commits, pushes, workflow transitions, or gate decisions.

If user explicitly asks for remediation:

1. address confirmed `CRITICAL` findings first
2. then address confirmed `WARNING (real)` findings
3. treat `WARNING (theoretical)` as informational unless hardening requested
4. handle `SUGGESTION` items only if trivial or explicitly requested

If fix addresses repeated pattern, check same pattern across analyzed scope for consistency.

## Re-review Policy

Re-review optional, context-driven. Use when:
- confirmed critical issues fixed
- confirmed real warnings fixed in paths with ripple effects
- user explicitly asks for second pass
- initial synthesis had low confidence

No fixed number of rounds. No forced escalation loops. No binding to merge/release/phase gates unless user explicitly asks.

## Verdict Rules

For analyzed scope only:

- **APPROVED** — no confirmed `CRITICAL` and no confirmed `WARNING (real)` remain
- **CONDITIONALLY APPROVED** — only `SUGGESTION` and/or `WARNING (theoretical)` remain
- **NEEDS CHANGES** — at least one confirmed `CRITICAL` or confirmed `WARNING (real)` remains
- **UNRESOLVED** — contradiction or missing evidence prevents reliable conclusion

For non-review tasks, adapt label if needed, preserve logic. Never present `APPROVED` as merge permission or workflow authorization unless user explicitly asks. Analytical verdict only.

## Output Format

### Default review format

```markdown
## Judgment Day — {target}

### Lens selection
- Pass A: {lens}
- Pass B: {lens}

### Findings

| Finding | Pass A | Pass B | Severity | Status |
|---------|--------|--------|----------|--------|
| {issue} |  |  | CRITICAL | Confirmed |
| {issue} |  |  | WARNING (real) | Suspect |
| {issue} |  |  | WARNING (theoretical) | Info |

**Confirmed**: {count}
**Suspect**: {count}
**Contradictions**: {count}
**Assumptions used**: {project standards / generic engineering judgment / other}

### Key reasoning
- {why the strongest confirmed issue matters}
- {what remains uncertain}

### Confidence
{high | medium | low} — {why}

### Verdict
{APPROVED | CONDITIONALLY APPROVED | NEEDS CHANGES | UNRESOLVED}
```

### Default answer-synthesis format

```markdown
## Judgment Day — {question or target}

### Lens selection
- Pass A: {lens}
- Pass B: {lens}

### Pass comparison
- Pass A conclusion: {summary}
- Pass B conclusion: {summary}
- Main disagreement: {summary}

### Synthesized result
{best final answer with necessary qualifiers}

### Why this result wins
- {best evidence}
- {trajectory strength}
- {what was rejected and why}

### Confidence
{high | medium | low} — {why}
```

If user requests remediation, add:

```markdown
### Recommended next actions
1. {highest-value action}
2. {next action}
```

## Behavioral Guardrails

- Do not replace evidence with ceremony.
- Do not require any specific tool, infrastructure, or workflow.
- Do not assume sub-agents exist.
- Do not assume sdd phase exists.
- Do not create obligations user did not request.
- Do not confuse stylistic disagreement with substantive defect.
- Do not convert analytical uncertainty into false precision.
- Do not let both passes share same hidden premise without checking it.
- Do not stop at answer selection when synthesis is better supported.

## Language

English input → respond in English.

Preferred phrases: "judgment initiated", "two-pass review completed", "both passes agree", "synthesized result", "needs changes", "conditionally approved", "unresolved"
