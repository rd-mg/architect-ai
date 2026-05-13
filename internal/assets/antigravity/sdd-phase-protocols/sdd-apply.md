<!-- architect-ai:prompt-caching-anchor:start -->
# SDD Phase Protocol: sdd-apply
Project: architect-ai
Adapter: Antigravity
Version: 1.1
<!-- architect-ai:prompt-caching-anchor:end -->

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

## Execution Graph Awareness (MANDATORY)
**BEFORE** starting implementation, read `sdd/{change-name}/tasks` and verify the **Execution Graph**. 
- Identify parallel-safe tasks vs sequence-locked tasks.
- If current batch violates graph order, STOP and report.

## Atomic Commit Protocol
- **MANDATORY**: Each [x] task MUST be accompanied by a clean commit (if using Git).
- Format: `type(scope): message` (Conventional Commits).
- No attribution or "Co-authored-by" allowed.

## Phase: sdd-apply

Task: Implement batch {N} of tasks for "{change-name}". Batch size: {size}.

{if apply-progress exists:}
## Previous Progress
Topic key `sdd/{change-name}/apply-progress` contains prior state.
READ it via mem_search + mem_get_observation, MERGE with new progress,
SAVE the combined state. DO NOT overwrite — MERGE.

## Batch Scope
Tasks to complete in this batch: {list from tasks artifact}

## Constraints
- **Atomic Commits**: Verify each task has its own commit.
- Update tasks.md: mark each completed task with [x]
- If a task cannot be completed, mark as BLOCKED and note reason
- Follow the compact rules in Project Standards EXACTLY
- Do not modify files outside the scope of assigned tasks
- Do not start new tasks until assigned ones are done or blocked


## Empirical Verification Loop (+++Empirical)
- **MANDATORY**: Before concluding, you MUST perform an empirical verification of your findings/artifacts.
- Examples: run a script, check a file, verify a tool output, or perform a manual check of the logic.
- Record the evidence in the `empirical_proof` field of the return handshake.

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

## Return Envelope & Compliance per sdd-phase-common.md (Sections A-F)
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
