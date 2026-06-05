---
description: Check and inject Caveman Register Firewall in sdd-apply and related targets
agent: sdd-orchestrator
subtask: false
---

TASK: Verify and fix Caveman Register Firewall.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`

STEP 1 — Check status:
  ctx_execute("shell", `architect-ai firewall check 2>&1 || go run ./cmd/firewall check 2>&1`)

STEP 2 — Inject where missing:
  ctx_execute("shell", `architect-ai firewall inject 2>&1 || go run ./cmd/firewall inject 2>&1`)

STEP 3 — Update work-unit-commits compact rule in registry:
  Edit .atl/skill-registry.md → section ### work-unit-commits
  Add the commit template from C-SPEC-02.
  Verify: grep -c "Anti-patterns" .atl/skill-registry.md

STEP 4 — Rebuild deployed configs:
  ctx_execute("shell", `architect-ai build 2>&1 || go run ./cmd/build 2>&1`)

STEP 5 — Return Envelope:
  status: success|partial
  executive_summary: {firewall status, N files patched}
  artifacts: [internal/assets/_shared/caveman-firewall.md, patched files]
  next_recommended: "Phase D — context-mode integration"

---

## Completion Contract

This command uses structured completion tracking.

**When ALL steps above are complete and verified:**
  Emit: `<promise>FIREWALL_COMPLETE</promise>`
  Then emit the Return Envelope below.

**If a step cannot complete (blocker detected):**
  Emit: `<promise>FIREWALL_BLOCKED</promise>`
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
