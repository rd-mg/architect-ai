---
name: sdd-orchestrator
description: >
  Manage the SDD pipeline and orchestrate specialized sub-agents.
  Trigger: Coordinates SDD lifecycle phases.
license: MIT
metadata:
  author: rd-mg
  version: "3.0"
---

## Purpose

L1a SDD Orchestrator. Manages entire SDD pipeline from initialization to archival. Coordinates specialized sub-agents. Never does execution work inline.

## Skill Digestion Harness Protocol

Before every L2 delegation, execute Skill Digestion Harness:

### 1. Identify Skills
Identify required skills from project registry or `.atl/skill-manifest.yaml`.
- Enforce **max 3 tier-2 skills limit** (excluding Tier 1 execution fundamentals).
- If more than 3 matched, select top 3 by relevance to target task.

### 2. Digest Rules
For each selected skill, locate its `SKILL.md`. Extract **only** `## Compact Rules` or `## Rules` section.
- **FORBIDDEN**: Do NOT inject entire `SKILL.md` into sub-agent prompt.

### 3. Deliver Compact
Format extracted compact rules, inject directly into sub-agent prompt under `## Project Standards (auto-resolved)`.

### 4. Validate Return Contract
When sub-agent responds, check return contract for `skill_resolution` field:
- **`injected`**: Successful loading and matching.
- **`fallback-registry`**: Local skill registry used as fallback. Retry or log.
- **`none`**: No skills processed. Escalate and report diagnostic warning.
