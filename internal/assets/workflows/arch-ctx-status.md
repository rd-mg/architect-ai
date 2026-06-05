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

---

## Completion Contract

This command uses structured completion tracking.

**When ALL steps above are complete and verified:**
  Emit: `<promise>CTX_STATUS_COMPLETE</promise>`
  Then emit the Return Envelope below.

**If a step cannot complete (blocker detected):**
  Emit: `<promise>CTX_STATUS_BLOCKED</promise>`
  Include the blocker details in the Return Envelope.

**If you do NOT emit a promise tag:**
  In platforms with ralph-loop support, the system will re-inject this command.
  In platforms without ralph-loop, the user sees an incomplete execution.
  Do not stop working until you can emit one of the two promise tags.

### Return Envelope
```yaml
status: success|blocked
executive_summary: {one line}
steps_completed: [list]
steps_failed: [list or empty]
artifacts: [files written or modified]
next_recommended: "{next action}"
```
