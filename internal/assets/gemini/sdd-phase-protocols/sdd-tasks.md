# Phase Protocol: sdd-tasks

## Dependencies
- **Reads**: proposal, spec (if exists), design (if exists)
- **Writes**: `tasks` artifact (numbered checklist)

## Cognitive Posture
+++Pragmatic — Mechanical breakdown. No over-engineering. No speculative tasks.

## Model
sonnet — structured breakdown

## Sub-Agent Launch Template

```
+++Pragmatic
Execute the task with the minimum viable approach. No gold-plating. No
over-engineering. Break down ONLY what the spec and design require — do not
add speculative tasks.

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-tasks

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Task: Break down the approved proposal + spec + design for "{change-name}"
into an ordered, numbered checklist of implementable tasks.

## Format
- Use hierarchical numbering: 1.1, 1.2, 1.3, 2.1, ...
- Each top-level group corresponds to a domain or file area
- Each task is atomic (one developer can complete in < 30 minutes)
- Each task has a clear acceptance criterion

## Mandatory Structure
1. **Setup** (new files, configs, dependencies)
1.1 Create file X
1.2 Add dependency Y to manifest
2. **Implementation** (core logic)
2.1 Implement function Z per spec capability A
2.2 ...
3. **Tests**
3.1 Unit test for capability A
3.2 Integration test for flow B
4. **Documentation**
4.1 Update README
4.2 Add changelog entry
5. **Migration** (if applicable)
5.1 Pre-migrate script for field X
5.2 Post-migrate script for data Y

## Task Format
Each task: `- [ ] {number} {action verb} {target}. Acceptance: {condition}.`

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/tasks",
  topic_key: "sdd/{change-name}/tasks",
  type: "task-list",
  project: "{project}",
  content: "{your tasks markdown}"
)

## Size Budget: 530 words max

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Validate tasks are atomic (reject tasks that say "implement entire feature X")
- Check each task has acceptance criterion
- Check numbering is sequential and consistent
- Update state: `designing` → `tasks-ready`
- Next recommended: `sdd-apply`

## Failure Handling

- If sub-agent produces tasks larger than 30 minutes → return with "Break down task N.N further"
- If acceptance criteria missing → return `partial`, request completion
- If tasks reference non-existent spec capabilities → return `blocked`, route to sdd-spec for clarification
