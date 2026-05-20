<!-- architect-ai:generated v2 -->
<!-- PLATFORM: antigravity -->
<!-- RUNTIME: single-thread — all orchestration is sequential and simulated -->

# architect-ai — Antigravity Agent

## RUNTIME NOTICE

Antigravity is single-threaded. No real sub-agent delegation.
All L0/L1/L2 transitions are SEQUENTIAL and INLINE-SIMULATED.
Each "agent switch" uses ULTRA caveman framing + identity clear.

## Sequential Thinking — Always Inline

sequential_thinking MCP not available on Antigravity.
ALWAYS use Inline Hypothesis Branching for D1+D2 >= 5:
```
[SEQUENTIAL THINKING — inline]
Branch A: {approach} | Tradeoffs: {pros/cons} | Risk: {risk}
Branch B: {approach} | Tradeoffs: {pros/cons} | Risk: {risk}
[If D5>=2: Branch C: adversarial approach]
Decision: Branch {X} — {specific evidence}
[END SEQUENTIAL THINKING]
```

## Simulated Delegation Protocol

For EVERY sub-agent invocation:
```
Step 1: ULTRA: "[{from}→{to}] task: {one-line task description}"
Step 2: Load {to} agent's compact rules from .atl/skill-manifest.yaml
Step 3: Execute task inline following {to}'s contract and postures
Step 4: ULTRA: "[{to}→{from}] result: {summary}"
Step 5: Clear {to} identity — do NOT carry it forward
Step 6: Resume {from} identity
```

## Context Management (Antigravity — no /compact)

On D4 >= 2 (high context pressure):
```
Step 1: Save checkpoint to .atl/session.yaml (if available)
Step 2: mem_save("session/context-pack/{timestamp}", checkpoint) — if Engram available
Step 3: If no Engram: emit LITE "Context limit. Start new chat: 'resume {change} from .atl/'"
Step 4: Include in output: next_action, critical_facts, files_modified
```

## Phase DAG Enforcement (Antigravity)

Even in single-thread mode, enforce Phase DAG:
```bash
# Before any SDD phase
STATE=".atl/sdd-state.yaml"
[ -f "${STATE}" ] || { echo "BLOCKED: sdd-init required first"; exit 1; }
```

<!-- architect-ai:L0:start -->
<!-- architect-ai:L0:end -->

<!-- architect-ai:foundation:start -->
<!-- architect-ai:foundation:end -->
