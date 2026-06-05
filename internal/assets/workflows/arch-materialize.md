---
description: Regenerate all deployed platform configs using LLM synthesis + validation
agent: sdd-orchestrator
subtask: false
---

TASK: Regenerate deployed platform configs from internal/assets sources.

+++Pragmatic
[MODE 1 | D1=1, D2=0, D3=0, D4=0] Maintenance task — LLM synthesis mode.

STEP 1 — Read sources in parallel:
  ctx_batch_execute([
    {type:"read", path:"internal/assets/opencode/architect.md"},
    {type:"read", path:"internal/assets/opencode/sdd-orchestrator.md"},
    {type:"read", path:"internal/assets/_shared/adaptive-reasoning-gate-v2.md"},
    {type:"read", path:"internal/assets/_shared/caveman-firewall.md"},
    {type:"read", path:".atl/_generated/foundation.md"},
    {type:"read", path:"internal/assets/source-map.json"}
  ])

STEP 2 — Synthesize each platform config:
  For CLAUDE.md:
    Combine: architect (L0) + sdd-orchestrator (L1a, Section 0 IntentGate first)
    + foundation.md standards + adaptive-reasoning-gate-v2 + caveman-firewall
    + KEYWORD ROUTING table (for degraded sessions)
    Format: per CLAUDE.md template structure (read current CLAUDE.md as reference)
  
  For GEMINI.md:
    Same content, add KEYWORD ROUTING table as FIRST section (D-SPEC-04)
    Single-thread framing (no Task() delegation tool)
  
  For .antigravity/agent.md:
    Single-thread version: KEYWORD ROUTING first, then persona-architect content
    No multi-agent delegation sections
  
  For .github/copilot-instructions.md:
    Simplified version: skills referenced, no full orchestrator content

STEP 3 — Atomic write each file (write to .tmp, rename):
  For each synthesized file:
    Write content to {target}.tmp
    Rename {target}.tmp → {target}

STEP 4 — Validate (ralph-loop: repeat until PASS):
  ctx_execute("shell", `go run ./cmd/architect-ai check all 2>&1`)
  → If CHECK: FAIL: identify residual placeholders → fix inline → re-run Step 3 → re-run Step 4
  → Repeat until CHECK: PASS

STEP 5 — Emit completion:
  <promise>MATERIALIZE_COMPLETE</promise>

Return Envelope:
  status: success|blocked
  files_written: [list]
  check_result: PASS|FAIL
  iterations_needed: N
