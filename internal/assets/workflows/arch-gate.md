---
description: Check and inject Adaptive Reasoning Gate v2 in all orchestrator targets
agent: sdd-orchestrator
subtask: false
---

TASK: Verify and fix Adaptive Reasoning Gate in architect-ai orchestrators.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`

STEP 1 — Check current gate status:
  ctx_execute("shell", `architect-ai gate check 2>&1 || go run ./cmd/gate check 2>&1`)

STEP 2 — Inject gate where missing (idempotent):
  ctx_execute("shell", `architect-ai gate inject 2>&1 || go run ./cmd/gate inject 2>&1`)

STEP 3 — Purge L2 auto-scoring:
  ctx_execute("shell", `architect-ai gate l2-purge 2>&1 || go run ./cmd/gate l2-purge 2>&1`)

STEP 4 — Update adaptive-reasoning compact rule in .atl/skill-registry.md:
  Apply the B-SPEC-06 canonical compact rule replacement.
  Verify: architect-ai check registry

STEP 5 — Rebuild deployed configs (gate is in templates):
  ctx_execute("shell", `architect-ai build 2>&1 || go run ./cmd/build 2>&1`)

STEP 6 — Return Envelope:
  status: success|partial|blocked
  executive_summary: {N targets gated | N L2 files purged | gate check status}
  artifacts: [modified orchestrator files]
  next_recommended: "Phase C — caveman firewall"
