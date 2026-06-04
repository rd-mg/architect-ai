---
description: Verify health of context-mode and estimate D4 Context Pressure
agent: sdd-orchestrator
subtask: false
---

TASK: Check context-mode health and update D4 estimate.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`

STEP 1: Check stats
  ctx_execute("shell", "ctx_stats 2>&1")
  → extract context_health score

STEP 2: Map to D4
  Map context_health to D4 estimate per D-SPEC-05 rules.

STEP 3: Report
  status: success|blocked
  executive_summary: {platform, context_health, D4_estimate, savings_ratio}
  artifacts: []
  next_recommended: ""
