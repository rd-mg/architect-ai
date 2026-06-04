---
description: Execute all architect-ai validation checks
agent: sdd-orchestrator
subtask: false
---

TASK: Run all architect-ai checks.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`

STEP 1: Run check all
  ctx_execute("shell", "architect-ai check all 2>&1 || go run ./cmd/architect-ai check all 2>&1")

STEP 2: Report
  status: success|blocked
  executive_summary: {Summary of passed and failed checks}
  artifacts: []
  next_recommended: ""
