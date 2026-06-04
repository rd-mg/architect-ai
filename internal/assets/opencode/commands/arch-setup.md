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
