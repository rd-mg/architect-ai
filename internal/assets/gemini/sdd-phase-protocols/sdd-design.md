# Phase Protocol: sdd-design

## Dependencies
- **Reads**: proposal artifact, spec artifact (if exists)
- **Writes**: `design` artifact

## Cognitive Posture
+++Critical + +++Systemic — Architecture needs both rigor and system view.

## Model
opus — architectural decisions

## Sub-Agent Launch Template

```
+++Critical
Evaluate objectively based on evidence. For each claim made or implied:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists?

+++Systemic
Analyze 2nd and 3rd order effects. What breaks elsewhere? What new
dependencies are created? What becomes harder to change later?

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-design

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Task: Produce the architecture design for "{change-name}". Based on proposal
and spec (if present), produce a design document covering:

## Mandatory Sections
- Architecture diagram (ASCII or Mermaid if supported)
- Module/component boundaries
- Data flow
- Interface contracts (what functions/methods are exposed)
- State management
- Error propagation model
- Integration points with existing code
- Migration path (if data model changes)
- Rollback strategy (if it fails in production)
- Alternative designs considered and why rejected
- Open questions (if any remain)

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/design",
  topic_key: "sdd/{change-name}/design",
  type: "architecture",
  project: "{project}",
  content: "{your design markdown}"
)

## Size Budget: 800 words max

## Return Envelope per sdd-phase-common.md Section D
Include: research_cache_hits: int, research_cache_misses: int
```

## Result Processing

- Validate all mandatory sections present
- Check `Alternative designs considered` is not empty (forces explicit decision-making)
- Check `Rollback strategy` is actionable (not "undo the changes")
- Update state: `specifying` → `designing`
- Next recommended: `sdd-tasks`

## Failure Handling

- If sub-agent cannot identify integration points → return `partial`, suggest sdd-explore round
- If design conflicts with active constraints in Context Pack → return `blocked`, escalate to user
- If Open Questions is non-empty → present to user, wait for resolution before sdd-tasks
