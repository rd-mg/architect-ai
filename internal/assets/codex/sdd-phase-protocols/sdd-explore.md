# Phase Protocol: sdd-explore

## Dependencies
- **Reads**: nothing (optional: prior context)
- **Writes**: `explore` artifact

## Cognitive Posture
+++Socratic — Reveal assumptions. Explore the problem space. Formulate questions.

## Model
sonnet — structural investigation, not architectural decisions

## Sub-Agent Launch Template

```
+++Socratic
Before producing artifacts, formulate 3 questions about unstated assumptions
in the request. Reveal what has NOT been said.

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-explore

Task: Investigate the topic "{topic}". Read the codebase. Compare approaches.
Identify constraints. Do NOT modify code. Do NOT create non-exploration files.

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/explore",
  topic_key: "sdd/{change-name}/explore",
  type: "architecture",
  project: "{project}",
  content: "{your exploration markdown}"
)

## Size Budget: 600 words max

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Check `skill_resolution` — if not `injected`, trigger re-read of registry
- Store `executive_summary` only; discard verbose output
- Extract any `questions` returned by Socratic mode and present to user
- Update state: `idle` → `exploring`

## Failure Handling

- If sub-agent returns `status: blocked` with unanswered questions → present to user, wait
- If sub-agent cannot find enough information → record as `partial`, suggest next steps
