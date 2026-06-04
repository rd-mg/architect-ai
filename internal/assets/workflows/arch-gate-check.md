---
description: Verify Adaptive Reasoning Gate status (read-only)
agent: sdd-orchestrator
subtask: false
---

TASK: Check Adaptive Reasoning Gate status without modifying files.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`

STEP 1: Run gate check
  ctx_execute("shell", "architect-ai gate check 2>&1 || go run ./cmd/gate check 2>&1")

STEP 2: Report
  status: success|blocked
  executive_summary: {gate check status, missing targets if any}
  artifacts: []
  next_recommended: ""
