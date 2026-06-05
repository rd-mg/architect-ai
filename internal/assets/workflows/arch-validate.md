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

---

## Completion Contract

This command uses structured completion tracking.

**When ALL steps above are complete and verified:**
  Emit: `<promise>VALIDATE_COMPLETE</promise>`
  Then emit the Return Envelope below.

**If a step cannot complete (blocker detected):**
  Emit: `<promise>VALIDATE_BLOCKED</promise>`
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
