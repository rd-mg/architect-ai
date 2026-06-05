---
description: Setup context-mode for current platform
agent: sdd-orchestrator
subtask: false
---

TASK: Setup context-mode for current platform.

STEP 1: Detect platform and check context-mode status
  ctx_execute("shell", "architect-ai setup --dry-run 2>&1 || go run ./cmd/setup --dry-run 2>&1")

STEP 2: Install and configure
  ctx_execute("shell", "architect-ai setup --yes 2>&1 || go run ./cmd/setup --yes 2>&1")

STEP 3: Verify
  ctx_execute("shell", "context-mode doctor 2>&1")

STEP 4: Read platform config
  ctx_execute("shell", "cat .atl/platform-config.yaml 2>&1")

STEP 5: Return Envelope
  status: success|blocked
  executive_summary: {platform, context-mode version, hooks active, doctor pass}
  artifacts: [.atl/platform-config.yaml, platform config files written]
  next_recommended: "architect-ai build && architect-ai gate inject && architect-ai firewall inject"

---

## Completion Contract

This command uses structured completion tracking.

**When ALL steps above are complete and verified:**
  Emit: `<promise>SETUP_COMPLETE</promise>`
  Then emit the Return Envelope below.

**If a step cannot complete (blocker detected):**
  Emit: `<promise>SETUP_BLOCKED</promise>`
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
