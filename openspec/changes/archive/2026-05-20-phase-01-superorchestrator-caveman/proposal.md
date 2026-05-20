# Proposal: Phase 1 - Super-Orchestrator v2.0: Inline Execution + Delegation Triggers + Caveman

## Intent
Upgrade L0 Super-Orchestrator to v2.0. Replace the rigid "router-only" v1 model with a 3-Mode architecture (Mode A: Inline, Mode B: SDD, Mode C: General). Introduce 6 Mandatory Delegation Triggers to prevent context inflation, session-wide Execution Modes (Interactive/Automatic), and per-phase Model Routing. Maintain Caveman compression and section injection from v1.

## Scope
### In Scope
- L0 3-Mode architecture (Inline, SDD, General)
- 6 Mandatory Delegation Triggers (4-file rule, multi-file write, PR rule, etc.)
- Session-wide Execution Mode (Interactive vs Automatic)
- Model Routing by phase (Opus, Sonnet, Haiku)
- Caveman dual-mode compression in all adapters (OpenCode, Claude, VSCode, Antigravity, Gemini CLI)
- Go: `internal/install/adapter/injector.go` with section injection

### Out of Scope
- GGA pre-commit hook (Phase 7)
- Odoo L3 agents (Phases 6, 9)

## Impact
- L0 can execute trivial tasks (git status, 1-file read) without L1 latency.
- Strict delegation triggers prevent the L0 from burning context on complex tasks.
- Model routing reduces costs by using smaller models (Haiku) for mechanical phases (Archive).
- Caveman compression reduces token waste across all platforms.

