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

## Step 0: Deep Code Exploration (Sequential Thinking)
- **MANDATORY**: Call `sequential_thinking` to map the target modules and identify dependencies BEFORE running any search tools.

## ADR Pre-check (MANDATORY)
**BEFORE** performing any code search, check for existing Architecture Decision Records:
- `mem_search(query: "arch/_global/decision", project: "{project}")`
- `mem_search(query: "sdd/{project}/design/main", project: "{project}")`

## Code Investigation (Section B — 5-Step Skim Protocol)
1. **Ripgrep Discovery**: Identify candidate files using specific keywords.
2. **Structural Skim**: List functions and types (e.g., `rg "^func|^type|^var|^const" {file}`).
3. **Boundary Check**: Identify imports and dependencies to see what OTHER files are affected.
4. **Logic Isolation**: Read only the specific blocks of code (functions/methods) identified in step 2.
5. **Pattern Comparison**: Compare found implementation with established project patterns.

Do NOT `cat` entire files unless they are under 50 lines. Identify constraints. Do NOT modify code.

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

## Return Envelope & Compliance per sdd-phase-common.md (Sections A-F)
```

## Result Processing

- Check `skill_resolution` — if not `injected`, trigger re-read of registry
- Store `executive_summary` only; discard verbose output
- Extract any `questions` returned by Socratic mode and present to user
- Update state: `idle` → `exploring`

## Failure Handling

- If sub-agent returns `status: blocked` with unanswered questions → present to user, wait
- If sub-agent can't find enough information → record as `partial`, suggest next steps
