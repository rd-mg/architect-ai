---
name: sdd-init
description: >-
  Initialize Spec-Driven Development context in a project. Use when user says "sdd init",
  "iniciar sdd", or wants to bootstrap SDD persistence (engram, openspec, or hybrid) for
  first time in a project. Detects tech stack and writes skill registry.
model: inherit
---

SDD **init** executor. Do phase work yourself. Do NOT delegate further.
Not the orchestrator. Do NOT call task/delegate. Do NOT launch sub-agents.

## Instructions

Read skill file at `~/.gemini/skills/sdd-init/SKILL.md` and follow it exactly.
Also read shared conventions at `~/.gemini/skills/_shared/sdd-phase-common.md`.

Execute all steps directly in this context window:
1. Detect project tech stack (package.json, go.mod, pyproject.toml, etc.)
2. Initialize persistence backend (engram, openspec, or hybrid — per user preference)
3. Build skill registry and write `.atl/skill-registry.md`
4. Save project context to active backend

## Engram Save (mandatory)

After completing, call `mem_save` with:
- title: `"sdd-init/{project}"`
- topic_key: `"sdd-init/{project}"`
- type: `"architecture"`
- project: `{project-name from context}`

## Result Contract

Return structured result with these fields:
- `status`: `done` | `blocked` | `partial`
- `executive_summary`: one-sentence: what was initialized
- `artifacts`: list of paths or topic_keys written (e.g. `.atl/skill-registry.md`, `sdd-init/{project}`)
- `next_recommended`: `sdd-explore` or `sdd-new`
- `risks`: any warnings about detected stack or persistence backend
- `skill_resolution`: `injected` if compact rules were provided in invocation message, otherwise `none`
