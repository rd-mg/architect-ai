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


## REVIEW WORKLOAD GUARD [L1a sdd-orchestrator — MANDATORY between sdd-tasks and sdd-apply]

Location in orchestrator flow: AFTER sdd-tasks result received, BEFORE launching sdd-apply.

### Check flow

```
sdd-tasks returns result_contract
       │
       ▼
READ result_contract.review_workload
       │
       ├── budget_risk = "low" AND chained_prs_recommended = false
       │     → Proceed to sdd-apply normally
       │     → Pass delivery_strategy = session.delivery_strategy
       │
       ├── budget_risk = "medium" OR "high"
       │     ├── delivery_strategy = "ask-on-risk"
       │     │     → Emit forecast to user (LITE)
       │     │     → Ask: split / single-pr+exception / stop
       │     │     → Wait for user response
       │     │     → Cache choice in session.current_change_delivery_decision
       │     │
       │     ├── delivery_strategy = "auto-chain"
       │     │     → Compute slice 1 (SliceForDelivery(tasks, maxLines=400))
       │     │     → Launch sdd-apply with slice 1 ONLY
       │     │     → After PR for slice 1: continue to slice 2, etc.
       │     │
       │     ├── delivery_strategy = "single-pr"
       │     │     → Ask user for size:exception rationale
       │     │     → mem_save("sdd/{change}/size-exception", rationale)
       │     │     → Proceed with single PR
       │     │
       │     └── delivery_strategy = "exception-ok"
       │           → mem_save("sdd/{change}/size-exception", "auto-approved by exception-ok strategy")
       │           → Proceed with single PR
       │
       └── No review_workload in result (old format)
             → WARN: sdd-tasks did not include workload forecast
             → Estimate manually: count task files × avg 50 lines
             → Apply same decision logic
```

### IMPORTANT: Automatic mode does NOT skip this guard.
Even with execution_mode = automatic, the Review Workload Guard ALWAYS runs.
Automatic means "no pause between phases", not "skip reviewer protection".

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
