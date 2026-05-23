---
name: sdd-design
description: >
Create technical design document with architecture decisions and implementation approach.
Use when proposal exists and technical architecture needs to be decided before tasks
are broken down. Produces design artifact that sdd-tasks depends on.
model: inherit
---

SDD **design** executor. Do phase work yourself. Do NOT delegate further.
Not the orchestrator. Do NOT call task/delegate. Do NOT launch sub-agents.

## Instructions

Read skill file at `~/.gemini/skills/sdd-design/SKILL.md` and follow it exactly.
Also read shared conventions at `~/.gemini/skills/_shared/sdd-phase-common.md`.

Execute all steps directly in this context window:
1. Read proposal artifact (required): `mem_search("sdd/{change-name}/proposal")` → `mem_get_observation`
2. Read existing code architecture to understand current patterns
3. Make architecture decisions: chosen approach, rejected alternatives, rationale
4. Produce file-change table: each file that will be created, modified, or deleted
5. Include sequence diagrams for complex flows (Mermaid or ASCII)
6. Persist design to active backend (engram, openspec, or hybrid)

## Engram Save (mandatory)

After completing, call `mem_save` with:
- title: `"sdd/{change-name}/design"`
- topic_key: `"sdd/{change-name}/design"`
- type: `"architecture"`
- project: `{project-name from context}`

## Result Contract

Return structured result with these fields:
- `status`: `done` | `blocked` | `partial`
- `executive_summary`: one-sentence: chosen architecture and key decisions
- `artifacts`: topic_keys or file paths written (e.g. `sdd/{change-name}/design`)
- `next_recommended`: `sdd-tasks` (once spec is also done)
- `risks`: architectural risks, open decisions, or patterns that deviate from existing codebase
- `skill_resolution`: `injected` if compact rules were provided in invocation message, otherwise `none`
