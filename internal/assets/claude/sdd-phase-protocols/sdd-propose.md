# Phase Protocol: sdd-propose

## Dependencies
- **Reads**: exploration artifact (optional)
- **Writes**: `proposal` artifact

## Cognitive Posture
+++Critical — Evaluate feasibility with rigor. Identify biases and unproven assumptions before committing.

## Model
opus — architectural decisions

## Sub-Agent Launch Template

```
+++Critical
Evaluate objectively based on evidence. For each claim made or implied:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists? Do not accept aesthetic preferences
as evidence.

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-propose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Task: Create a change proposal for "{change-name}". Read exploration (if any).
Produce: proposal.md with scope, approach, affected areas, rollback plan,
success criteria, capabilities section.

## Mandatory Sections
- Scope (what's in, what's out)
- Approach (high-level strategy)
- Affected Areas (concrete file paths where possible)
- Rollback Plan (how to undo if this fails)
- Success Criteria (observable conditions for "done")
- Capabilities (contract with sdd-spec — new/modified/none)

## Artifact Store: {mode}

## Persistence (MANDATORY)

If mode is `engram` or `hybrid`, call:
```
mem_save(
  title: "sdd/{change-name}/proposal",
  topic_key: "sdd/{change-name}/proposal",
  type: "architecture",
  project: "{project}",
  content: "{your proposal markdown}"
)
```

If mode is `openspec` or `hybrid`:
- Write `openspec/changes/{change-name}/proposal.md`
- Write `openspec/changes/{change-name}/state.yaml` (MUST create initial version on new change)

### Atomic Write Pattern (state.yaml)
1. Write to `state.yaml.tmp`
2. Rename to `state.yaml`

### Validation
After every write to `state.yaml`, call `architect-ai sdd-status {change-name}`. If it fails, fix the file immediately.

## Size Budget: 450 words max. Use bullets and tables over prose.

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Check `executive_summary` length — must be < 100 words
- Validate `Capabilities` section is filled (not "TODO")
- Update state: `exploring` → `proposing`
- Next recommended: `sdd-spec` or `sdd-design`

## Failure Handling

- If proposal lacks rollback plan → return `partial`, ask sub-agent to complete
- If Capabilities section says "unclear" → route to sdd-explore for more investigation
