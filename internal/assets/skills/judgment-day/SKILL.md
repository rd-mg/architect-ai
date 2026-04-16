---
name: judgment-day
description: perform adaptive two-pass adversarial reasoning and synthesis for code, architecture, diffs, specifications, implementation plans, research questions, and technical decisions. use when the user asks for judgment day, judgment-day, dual review, adversarial review, two-judge reasoning, reasoning court, courteval, doble review, juzgar, or when a single-pass answer may miss hidden flaws. this skill defines how the agent should think, compare competing hypotheses, evaluate evidence quality and reasoning trajectories, and produce an analytical verdict without assuming any specific sdd phase, workflow, tool, or execution environment.
---

# Judgment Day

## Overview

Use this skill to run a disciplined adversarial analysis through **two independent reasoning passes** and then synthesize them into one decision-ready result.

This skill defines **how to think**, not how to orchestrate work.
It must remain compatible with both **sdd** and **non-sdd** contexts.
It must not impose lifecycle, tooling, repository, or delivery assumptions that were not already present in the task.

## Core Principle

Do the work in three stages:

1. **Pass A** — build one serious interpretation, answer, or critique
2. **Pass B** — build a materially different interpretation, answer, or critique
3. **Synthesis** — compare both passes, audit their evidence and reasoning paths, and produce the most defensible final result

The goal is not roleplay. The goal is to create structured disagreement that exposes hidden errors, weak evidence, missing assumptions, and premature convergence.

## Independence Rule

Keep the two passes independent until synthesis.

- Do not let Pass A bias Pass B before both are complete.
- Do not reuse the same framing sentence-for-sentence across both passes.
- When tools or retrieval are available, prefer different search angles, candidate explanations, or falsification checks.
- When tools are not available, still keep separate notes for each pass until synthesis.

If the environment supports sub-agents, delegation, or parallel execution, they may be used.
If not, perform both passes sequentially while preserving independence.

## Scope Compatibility

This skill must work whether the target is in:

- sdd or non-sdd work
- discovery, design, implementation, refactor, review, debugging, research, or incident response
- code, diffs, specs, plans, migrations, tests, architecture, procedures, or factual questions

Do **not** assume:

- a specific sdd phase exists
- a registry, memory system, or resolver exists
- sub-agents, delegate, delegation_read, mem_search, git, pr state, or branch state are available
- the user wants edits, fixes, commits, pushes, or status changes
- the task is necessarily code review; it may be reasoning, evaluation, or answer synthesis

If project standards, architecture rules, contracts, or phase-specific constraints are already available in context, apply them.
If not, proceed with generic engineering or analytical judgment and state that assumption explicitly.

## Adaptive Lens Selection

Choose the two passes to create **meaningful tension**, not superficial duplication.
Use the pair that best matches the target.

### For code, diffs, bugs, tests, and implementations

- **Pass A — local correctness lens**
  Focus on correctness, edge cases, error handling, safety, state transitions, and direct implementation risk.
- **Pass B — system impact lens**
  Focus on contracts, integration assumptions, invariants, operability, performance, maintainability, and downstream effects.

### For architecture, specs, and implementation plans

- **Pass A — feasibility lens**
  Focus on internal coherence, completeness, assumptions, sequencing, and whether the plan can work as written.
- **Pass B — failure-mode lens**
  Focus on hidden coupling, migration hazards, rollback gaps, scalability limits, observability gaps, and where the design breaks under realistic stress.

### For factual or research-heavy questions

- **Pass A — best-supported answer lens**
  Build the strongest answer supported by the available evidence.
- **Pass B — adversarial falsification lens**
  Try to disprove, narrow, or qualify Pass A by identifying contradictory evidence, missing links, alternative interpretations, or unsupported jumps.

### For policy, process, or operational guidance

- **Pass A — rule and compliance lens**
  Focus on what the stated policy, contract, or process explicitly requires.
- **Pass B — execution reality lens**
  Focus on ambiguity, edge conditions, operator behavior, practical failure points, and how the rule behaves in real workflows.

### For ambiguous targets

If the target spans multiple categories, choose the pair that produces the strongest disagreement surface and state the chosen lenses explicitly.

## Pass Procedure

### 1. Confirm the target

Identify the actual scope first.
If the scope is ambiguous enough to invalidate the review, ask once for the missing scope.
Otherwise proceed with the most defensible bounded interpretation and state that assumption.

### 2. Run Pass A

Produce a serious analysis, not a strawman.
Capture:

- main conclusion
- supporting evidence
- key assumptions
- critical reasoning steps
- uncertainty or open gaps

### 3. Run Pass B

Use a materially different lens.
Try to expose:

- contradictory evidence
- missing assumptions
- broken causal links
- overconfident claims
- alternative explanations
- unexamined edge cases
- invalid generalization from narrow evidence

### 4. Evaluate trajectories, not just outputs

Do not compare only the final answers.
Inspect the **reasoning trajectory** of each pass:

- Was the evidence relevant?
- Were key assumptions explicit?
- Did the pass skip a necessary inference step?
- Did it confuse correlation with causation?
- Did it rely on stale, weak, or indirect support?
- Did it ignore a known contract, invariant, or stated requirement?
- Did it collapse uncertainty into an unjustified definitive claim?

A pass with a polished answer but a weak trajectory should lose to a rougher answer with stronger evidence and cleaner logic.

## Agreement Trap Check

If both passes converge quickly, do **not** assume correctness.
Run a short anti-convergence check:

- What shared assumption could make both passes wrong?
- What evidence would most likely overturn both?
- Did both passes inherit the same framing error from the prompt or context?
- Are both repeating the same unsupported claim in different words?

When both passes agree on a weak basis, reduce confidence and say so.

## Synthesis Rule

Synthesis is not majority voting.
Choose the final result by weighing:

1. evidence quality
2. reasoning quality
3. contract and requirement alignment
4. realism of failure scenarios
5. remaining uncertainty

The synthesizer may do one of four things:

- **Select Pass A** when it is clearly stronger
- **Select Pass B** when it is clearly stronger
- **Merge both** when each contributes valid, complementary insight
- **Synthesize a new result** when both are incomplete or partially wrong but the available evidence supports a better conclusion

Do not pretend both passes are equally valid when one is materially stronger.
Do not force symmetry when the evidence is asymmetric.

## Finding Buckets

When the task is review-oriented, classify findings into these buckets:

- **Confirmed** — both passes agree, or one pass has strong direct evidence and the other does not materially refute it
- **Suspect** — raised by one pass only and not yet strongly evidenced
- **Contradiction** — both passes materially disagree on the same claim
- **Info** — notable but non-blocking observations

For answer-oriented tasks, use equivalent language if more natural, but preserve the same logic.

## Severity Model

Classify each review finding as exactly one of these:

- **CRITICAL** — likely production-breaking, security-relevant, corrupting, or fundamentally incorrect
- **WARNING (real)** — realistic bug or operational problem under normal use
- **WARNING (theoretical)** — requires contrived, unsupported, or highly artificial conditions
- **SUGGESTION** — useful improvement but not required for correctness

### Reality test for warnings

Ask:

> Can a normal user, system state, or expected workflow trigger this without contrivance?

- If **yes** → `WARNING (real)`
- If **no** → `WARNING (theoretical)`

Treat theoretical warnings as informational by default unless the user explicitly asks for hardening.

## Confidence Rule

Always communicate confidence based on evidence quality, not tone.

Use:

- **high confidence** — direct evidence, coherent trajectory, no material contradiction
- **medium confidence** — plausible conclusion with limited or indirect evidence
- **low confidence** — unresolved contradiction, missing evidence, or strong dependence on assumptions

Never present confidence as certainty.

## Evidence Rule

Every non-trivial claim should include the strongest available support:

- file and line, when available
- exact contract, invariant, or requirement involved
- concrete failure scenario or falsification path
- why it matters in practice
- smallest safe correction direction, when remediation is requested

Do not inflate severity without evidence.
Do not downgrade real risk because the output looks clean or familiar.

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

## Verdict Rules

Use these verdicts only for the analyzed scope:

- **APPROVED** — no confirmed `CRITICAL` and no confirmed `WARNING (real)` remain
- **CONDITIONALLY APPROVED** — only `SUGGESTION` and-or `WARNING (theoretical)` remain
- **NEEDS CHANGES** — at least one confirmed `CRITICAL` or confirmed `WARNING (real)` remains
- **UNRESOLVED** — contradiction or missing evidence prevents a reliable conclusion

For non-review tasks, adapt the label if needed, but preserve the logic.
For example: `best-supported answer`, `qualified answer`, or `unresolved`.

Never present `APPROVED` as merge permission or workflow authorization unless the user explicitly asks for that meaning.
It is an analytical verdict only.

## Output Format

Use the structure that best fits the task.

### Default review format

```markdown
## Judgment Day — {target}

### Lens selection
- Pass A: {lens}
- Pass B: {lens}

### Findings

| Finding | Pass A | Pass B | Severity | Status |
|---------|--------|--------|----------|--------|
| {issue} | ✅ | ✅ | CRITICAL | Confirmed |
| {issue} | ✅ | ❌ | WARNING (real) | Suspect |
| {issue} | ❌ | ✅ | WARNING (theoretical) | Info |

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

If the user also requests remediation, add:

```markdown
### Recommended next actions
1. {highest-value action}
2. {next action}
```

## Behavioral Guardrails

- Do not replace evidence with ceremony.
- Do not require any specific tool, infrastructure, or workflow.
- Do not assume sub-agents exist.
- Do not assume an sdd phase exists.
- Do not create obligations the user did not request.
- Do not confuse stylistic disagreement with substantive defect.
- Do not convert analytical uncertainty into false precision.
- Do not let both passes share the same hidden premise without checking it.
- Do not stop at answer selection when synthesis is better supported.

## Language

- Spanish input → respond in spanish;
- English input → respond in english.

Preferred phrases:

- English: "judgment initiated", "two-pass review completed", "both passes agree", "synthesized result", "needs changes", "conditionally approved", "unresolved"
- Spanish: "juicio iniciado", "revisión de dos pasadas completada", "ambas pasadas coinciden", "resultado sintetizado", "necesita cambios", "aprobado con reservas", "sin resolver"
