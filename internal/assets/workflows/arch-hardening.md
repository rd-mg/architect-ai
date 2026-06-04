---
description: Full architect-ai T-0 hardening sequence — run all phases A, B, C, D
agent: sdd-orchestrator
subtask: false
---

TASK: Execute full T-0 hardening sequence for architect-ai.

+++Pragmatic
[MODE 1 | D1=1, D2=0, D3=0, D4=0] Executing sequential hardening operations.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`
- Current project: !`echo -n "$(basename $(pwd))"`

PRE-CHECK: Verify Go 1.22+ installed
  ctx_execute("shell", `go version 2>&1`)
  → If error: BLOCKED — install Go 1.22+ from golang.org/dl

PHASE A — Config Materialization:
  STEP A1: Generate foundation.md
    ctx_execute("shell", `go run ./cmd/foundation 2>&1`)
    → Exit != 0: ABORT with phase A blocked message

  STEP A2: Build deployed configs
    ctx_execute("shell", `go run ./cmd/build 2>&1`)
    → Exit != 0: ABORT with phase A blocked message

  STEP A3: Validate build
    ctx_execute("shell", `go run ./cmd/architect-ai check all 2>&1`)

PHASE B — Adaptive Reasoning Gate:
  STEP B1: Inject gate into orchestrators
    ctx_execute("shell", `go run ./cmd/gate inject 2>&1`)

  STEP B2: Purge L2 auto-scoring
    ctx_execute("shell", `go run ./cmd/gate l2-purge 2>&1`)

  STEP B3: Verify gate
    ctx_execute("shell", `go run ./cmd/gate check 2>&1`)

PHASE C — Caveman Firewall:
  STEP C1: Inject firewall
    ctx_execute("shell", `go run ./cmd/firewall inject 2>&1`)

  STEP C2: Verify firewall
    ctx_execute("shell", `go run ./cmd/firewall check 2>&1`)

PHASE D — context-mode Setup:
  STEP D1: Check if already configured
    ctx_execute("shell", `cat .atl/platform-config.yaml 2>/dev/null || echo "NOT_CONFIGURED"`)
    → If NOT_CONFIGURED: run setup

  STEP D2: Setup context-mode (if needed)
    ctx_execute("shell", `go run ./cmd/setup --yes 2>&1`)

  STEP D3: Verify context-mode
    ctx_execute("shell", `context-mode doctor 2>&1`)

FINAL: Full system check
  ctx_execute("shell", `go run ./cmd/architect-ai check all 2>&1`)

RETURN ENVELOPE:
  status: success|partial|blocked
  executive_summary: {
    phase_a: {configs_built: N, check: PASS|FAIL},
    phase_b: {gate_targets: N, l2_purged: N},
    phase_c: {firewall_targets: N},
    phase_d: {platform: X, context_mode: Y, hooks: Z},
    final_check: PASS|FAIL
  }
  artifacts: [CLAUDE.md, GEMINI.md, .antigravity/agent.md, .github/copilot-instructions.md, .atl/_generated/foundation.md, .atl/platform-config.yaml]
  next_recommended: "git add -A && git commit -m 'chore(architect-ai): T-0 hardening v2 complete'"
  risks: "Review modified L2 SKILL.md files manually to verify purge was clean"
