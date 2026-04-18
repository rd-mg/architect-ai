# Phase Protocol: sdd-apply

## Dependencies
- **Reads**: tasks artifact, spec (if exists), design (if exists), apply-progress (if continuation)
- **Writes**: `apply-progress` artifact (updated after each batch)

## Cognitive Posture
+++Pragmatic — Execute the spec. Don't freelance.

## Model
sonnet — implementation work

## Sub-Agent Launch Template

```
+++Pragmatic
Execute the task with the minimum viable approach. Ship the smallest correct
change that satisfies the spec. Do exactly what was asked — no scope creep,
no speculative additions.

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

{if strict_tdd true:}
## STRICT TDD MODE IS ACTIVE
Test runner: {test-command}
Follow strict-tdd.md procedure. Do NOT fall back to Standard Mode.
Write failing test → verify red → implement → verify green → refactor.

## Phase: sdd-apply

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Task: Implement batch {N} of tasks for "{change-name}". Batch size: {size}.

{if apply-progress exists:}
## Previous Progress
Topic key `sdd/{change-name}/apply-progress` contains prior state.
READ it via mem_search + mem_get_observation, MERGE with new progress,
SAVE the combined state. DO NOT overwrite — MERGE.

## Batch Scope
Tasks to complete in this batch: {list from tasks artifact}

## Constraints
- Update tasks.md: mark each completed task with [x]
- If a task cannot be completed, mark as BLOCKED and note reason
- Follow the compact rules in Project Standards EXACTLY
- Do not modify files outside the scope of assigned tasks
- Do not start new tasks until assigned ones are done or blocked

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/apply-progress",
  topic_key: "sdd/{change-name}/apply-progress",
  type: "implementation-state",
  project: "{project}",
  content: "{batch progress with status per task}"
)

## Size Budget: 400 words (progress report). Code changes themselves are separate.

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Check that tasks.md was updated (completed tasks marked [x])
- Verify blocked tasks have reason + next action
- Update state: `tasks-ready` → `applying`
- Next recommended: `sdd-verify` (if all tasks done) or `sdd-apply` (next batch)

## Failure Handling

- If sub-agent reports errors in file operations → return `blocked`, escalate
- If tests fail in STRICT TDD mode → sub-agent must stop and report (do NOT force green)
- If apply-progress merge conflicts with previous state → return `blocked`, route to context-guardian
- If tasks.md cannot be found → return `blocked`, state integrity broken
