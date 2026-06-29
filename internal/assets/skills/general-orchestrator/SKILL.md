---
name: general-orchestrator
description: >
  Manage non-SDD general tasks and orchestrate general sub-agents.
  Trigger: Manages general user requests.
license: MIT
metadata:
  author: rd-mg
  version: "3.0"
---

## Purpose

L1a General Orchestrator — handles non-SDD general tasks (brainstorm, solver, research). Coordinates sub-agents. Never executes inline.

## Skill Digestion Harness Protocol

Execute BEFORE every L2 delegation:

### 1. Identify Skills
Identify from project registry or `.atl/skill-manifest.yaml`.
- Enforce **max 3 tier-2 skills** (excluding Tier 1 execution fundamentals).
- If > 3 matched, select top 3 by relevance.

### 2. Digest Rules
For each selected skill, locate `SKILL.md`. Extract **ONLY** `## Compact Rules` or `## Rules` section.
- **FORBIDDEN**: Do NOT inject full `SKILL.md` into sub-agent prompt.

### 3. Deliver Compact
Format extracted compact rules → inject under `## Project Standards (auto-resolved)`.

### 4. Validate Return Contract
Check sub-agent response for `skill_resolution`:
- **`injected`**: Success.
- **`fallback-registry`**: Local registry fallback. Retry or log.
- **`none`**: No skills processed. Escalate + report diagnostic warning.
