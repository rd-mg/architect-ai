# Proposal: Phase 2 - SDD v3: Phase DAG + Circuit Breaker + Result Contract + Apply Continuity


## Intent
Upgrade SDD to v3: enforce the Phase DAG mechanically (not just in prompts), add a JSON Result Contract, a Circuit Breaker for Ralph Loop prevention, and Apply Continuity checkpointing. Retains all v2 additions (branch isolation, Odoo detection, semantic audit, FMEA gate).

## Problem Context
1. **Worktree Friction (v2):** Replaced `git worktree` with universal `git checkout -b` temp branch.
2. **Context Blindness (v2):** Odoo detection missing — now auto-routes overlay skills.
3. **Spec vs Code Drift (v2):** Semantic Audit Step 0 in `sdd-verify` closes this gap.
4. **Happy Path Designs (v2):** FMEA + FSM gate in `sdd-design` mandated.
5. **Phase DAG bypassable (v3):** Agent could run `sdd-apply` without `sdd-design` completing — prompts are not enforcement.
6. **No Result Contract (v3):** Phase agents return freeform text — orchestrator cannot validate completion mechanically.
7. **Ralph Loops (v3):** Phase agent can retry failed approaches indefinitely, burning all quota.
8. **Apply Interruption (v3):** If `sdd-apply` is interrupted mid-run, no checkpoint exists — full restart required.

## Proposed Approach
1. **Apply Branch Protocol (v2):** Temp branch `apply/{change_name}` + ff-only/no-ff merge strategy.
2. **Odoo Auto-Detection (v2):** `__manifest__.py` / `requirements.txt` detection → auto-load Odoo skills.
3. **Semantic Audit Protocol (v2):** Step 0 in `sdd-verify` — `rg` assertions vs spec/design.
4. **FMEA + FSM Gate (v2):** Exit conditions in `sdd-design` before passing to `sdd-tasks`.
5. **sdd-hotfix (v2):** 5-step micro-cycle for low-risk urgent changes.
6. **Phase DAG via `.atl/sdd-state.yaml` (v3):** YAML state file is single source of truth. Each agent reads prerequisites before executing. Bypassing a phase = BLOCKED status.
7. **Result Contract JSON (v3):** Every phase agent emits a validated JSON envelope — orchestrator checks before advancing.
8. **Circuit Breaker (v3):** Max 3 attempts per phase. Exit code 2 = ABANDONED. Prevents Ralph Loops.
9. **Apply Continuity via `.atl/apply-progress.yaml` (v3):** Task-level checkpoint file. Resume from last completed task on restart.
