---
description: Rebuild architect-ai deployed configs and foundation from internal/assets sources
agent: sdd-orchestrator
subtask: false
---

TASK: Rebuild all architect-ai deployed configurations.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`

EXECUTION SEQUENCE (run in order, abort on any failure):

STEP 1 — Generate foundation.md:
  ctx_execute("shell", `architect-ai foundation 2>&1 || go run ./cmd/foundation 2>&1`)
  → If exit code != 0: report BUILD_BLOCKED with error. Stop.

STEP 2 — Build deployed configs:
  ctx_execute("shell", `architect-ai build 2>&1 || go run ./cmd/build 2>&1`)
  → If exit code != 0: report BUILD_FAILED with error. Stop.

STEP 3 — Validate:
  ctx_execute("shell", `architect-ai check all 2>&1 || go run ./cmd/architect-ai check all 2>&1`)

STEP 4 — Report result in Return Envelope format:
  status: success|blocked
  executive_summary: {N configs materialized | build failed at step N}
  artifacts: [CLAUDE.md, GEMINI.md, .antigravity/agent.md, .github/copilot-instructions.md, .atl/_generated/foundation.md]
  next_recommended: "Phase B — inject adaptive-reasoning gate"

---

## Completion Contract

This command uses structured completion tracking.

**When ALL steps above are complete and verified:**
  Emit: `<promise>BUILD_COMPLETE</promise>`
  Then emit the Return Envelope below.

**If a step cannot complete (blocker detected):**
  Emit: `<promise>BUILD_BLOCKED</promise>`
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
