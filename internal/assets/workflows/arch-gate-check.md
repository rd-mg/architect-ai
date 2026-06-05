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

---

## Completion Contract

This command uses structured completion tracking.

**When ALL steps above are complete and verified:**
  Emit: `<promise>GATE_CHECK_COMPLETE</promise>`
  Then emit the Return Envelope below.

**If a step cannot complete (blocker detected):**
  Emit: `<promise>GATE_CHECK_BLOCKED</promise>`
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
