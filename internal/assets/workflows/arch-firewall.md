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
