# Fix & Improvement Plan — `generic` Agent

**Agent ID:** `generic`
**Runtime:** Any agent not explicitly supported
**Config:** Standard markdown files in `.atl/agents/`
**Assets:** `internal/assets/generic/` (full set — used as fallback baseline)
**Priority:** 🟢 Low-Medium — Fallback; must be kept current as it feeds all other agents

---

## 1. Current State

Generic is the master fallback. It includes:
- All 9 SDD phase protocols
- `context-mode-routing-policy.md` ✅
- `engram-tool-routing-policy.md` ✅
- `general-orchestrator.md` ✅
- `sdd-orchestrator.md` ✅
- `thinking-agent.md` ✅
- Persona files (`persona-architect.md`, `persona-neutral.md`)

### sequential-thinking — Current State
- Phase protocols mandate `sequential_thinking` ✅
- No MCP config (generic has no specific IDE config writer) — depends on the agent runtime

### context-mode — Current State
- `context-mode-routing-policy.md` present ✅
- Routing policy is text-only — no MCP tool registration (by design for generic)

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| GEN-01 | sequential-thinking fallback block missing from `thinking-agent.md` | 🟠 High |
| GEN-02 | `ctx_batch_execute` not used in `sdd-explore.md` (multiple sequential greps) | 🟠 High |
| GEN-03 | CodeGraph steps absent from `sdd-explore.md` and `sdd-verify.md` | 🟠 High |
| GEN-04 | `context-mode-routing-policy.md` does not mention `ctx_batch_execute` usage pattern | 🟡 Medium |
| GEN-05 | Persona files not referenced in `sdd-orchestrator.md` header | Low |
| GEN-06 | `engram-tool-routing-policy.md` lacks L4 configurable TTL note | Low |

---

## 3. Fix Plan

### Fix GEN-01 — sequential-thinking Fallback

Add to `internal/assets/generic/thinking-agent.md`:
```markdown
**sequential_thinking tool unavailable fallback:**
Write an explicit <thinking> block containing:
1. Problem restatement (1 sentence)
2. D1–D4 classification with values
3. Three candidate approaches (bullet each)
4. Selected approach + rationale
This satisfies the sequential-thinking protocol contract without the MCP tool.
```

### Fix GEN-02 — Batch Execute in Explore

In `internal/assets/generic/sdd-phase-protocols/sdd-explore.md`:
```markdown
**Step 2 — Batch Code Search** (use ctx_batch_execute if available):
ctx_batch_execute([
  "rg '{change_topic}' --type {lang} -l",
  "rg '^func|^type|^class' {primary_file}",
  "rg '{change_topic}' --type {lang} -A 3 -B 1"
])
Fallback (no ctx_batch_execute): run each rg call sequentially, stopping when
enough evidence exists to proceed to Step 3.
```

### Fix GEN-03 — CodeGraph in Explore/Verify

In `sdd-explore.md` — add after batch search:
```markdown
**Step 2b — Semantic Graph** (if codegraph_context tool available):
codegraph_context(query: "{change_topic}", maxNodes: 25, format: "markdown")
→ Replace/supplement 3–5 of the above ripgrep calls with semantic results
```

In `sdd-verify.md` — add blast-radius check:
```markdown
**Step 1b — Impact Verification** (if codegraph_impact available):
codegraph_impact(nodeId: "{primary_node_from_explore}", depth: 3)
→ Confirms no unintended callers will break
```

### Fix GEN-04 — Routing Policy Update

Add to `context-mode-routing-policy.md`:
```markdown
## ctx_batch_execute Pattern

For multiple related searches in one phase step:
  ctx_batch_execute(["cmd1", "cmd2", "cmd3"])
  → Single MCP call, all outputs compressed, returned as one block
  
Better than: sequential rg calls that each flood the context window.
Threshold: use batch_execute when running ≥3 shell commands in the same step.
```

---

## 4. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | GEN-01: sequential-thinking fallback block |
| 1 | GEN-02: batch execute in sdd-explore.md |
| 1 | GEN-03: CodeGraph steps in explore + verify |
| 2 | GEN-04: Update context-mode-routing-policy.md |
| 3 | GEN-05/06: Minor doc improvements |

---

## David Kim CE — Coverage for `generic` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `generic`-specific deltas only.

| CE Topic | Status | `generic`-Specific Action |
|----------|--------|--------------------------|
| Protocol Shells | ❌ Missing | Add `/sdd.{phase}` header to all phase protocols |
| Token Budget Tracking | ❌ Missing | Add `token_budget` to `.atl/config.yaml` at install |
| Self-Refinement Engine | ❌ Missing | Add quality gate to `sdd-verify.md` |
| Dynamic Assembly | ❌ Missing | Keyword-filter skill injection in orchestrator |
| Few-Shot Examples | ❌ Missing | Add positive/negative output examples to explore + design |
| Progressive Disclosure | ❌ Missing | Add paging protocol to `sdd-explore.md` |
| Pareto-lang Operations | 🟡 Partial | Caveman compression present; add `/compress.summary` |
| Multi-agent Orchestration | ✅ Present | SDD phase DAG satisfies this |

> ⚠️ All code/config added for CE topics must be verified against actual runtime
> behavior — quality thresholds (0.85), token counts, and keyword lists need
> empirical tuning per agent/model combination.

---

## Context7 — Coverage for `generic` Agent

| Context7 Topic | Status | `generic`-Specific Action |
|----------------|--------|--------------------------|
| resolve-library-id before get-library-docs | ❌ Missing | Add two-step pattern to `sdd-explore.md` |
| topic parameter in get-library-docs | ❌ Missing | Enforce topic in all context7 calls |
| Token cap on docs fetch | ❌ Missing | Add tokens: 5000 cap to all get-library-docs calls |
| context7 for verify phase | ❌ Missing | Add API signature check to `sdd-verify.md` |
| context7 NOT for memory | ✅ Correct | Engram used for memory |

> ⚠️ VERIFY tool names: `mcp__context7__resolve-library-id` and
> `mcp__context7__get-library-docs` — confirm against actual Context7 MCP schema
> before adding to phase protocol instructions.

---

## Code Verification Notes for `generic`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `generic` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
