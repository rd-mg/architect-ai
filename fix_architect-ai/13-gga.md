# Fix & Improvement Plan — `gga` Agent

**Agent ID:** `gga`
**Runtime:** GGA — Gentleman Guardian Agent (by Gentleman-Programming)
**Install:** `brew install gga` / GitHub releases
**Config:** GGA project-level config files
**Assets:** `internal/assets/gga/` (has `AGENTS.md`, `sdd-orchestrator.md`, `thinking-agent.md`, `general-orchestrator.md`)
**Go Adapter:** None (managed as a tool via `update.ToolInfo`)
**Component:** `model.ComponentGGA`
**Priority:** 🟡 Medium — Full asset set but MCP config unknown; GGA is an AI provider switcher

---

## 1. Current State

GGA ("Gentleman Guardian Agent") is an AI provider switcher — it routes prompts to
different AI backends. architect-ai manages GGA as a **tool** (update/upgrade tracking)
rather than as a direct config target.

GGA has a full asset set: `AGENTS.md`, full SDD orchestrator, thinking-agent, and
general orchestrator. The `general-orchestrator.md` is particularly rich.

### sequential-thinking — Current State
- Phase protocols mandate `sequential_thinking` ✅
- GGA MCP configuration mechanism **unknown** to architect-ai ❌
- GGA likely proxies MCP from its underlying backend (Claude/Gemini/etc.)

### context-mode — Current State
- No hook configuration written for GGA ❌
- context-mode routing policy is injected in assets ✅ (textual, not tool-based)

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| GG-01 | GGA MCP config format/location unknown | 🟠 High |
| GG-02 | sequential-thinking registration unknown | 🟠 High |
| GG-03 | context-mode not hooked | 🟡 Medium |
| GG-04 | codegraph absent | 🟡 Medium |
| GG-05 | Update failure hint points to GitHub releases — no auto-recovery | Low |

---

## 3. Fix Plan

### Fix GG-01/02 — Research GGA Config

Action: Read GGA documentation at `https://github.com/Gentleman-Programming/gga`.
GGA likely uses a `gga.json` or similar config at project root. Once confirmed:

```go
var ggaMCPConfig = []byte(`{
  "mcpServers": {
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    },
    "context-mode": {
      "command": "context-mode",
      "args": ["--mcp"]
    },
    "codegraph": {
      "command": "npx",
      "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
    }
  }
}`)
```

### Fix GG-03 — context-mode for GGA

GGA underlying backend (e.g., Claude) will have context-mode hooked. GGA itself
may pass through tool calls to the backend. Verify this assumption by testing
`sequential_thinking` tool call via GGA → Claude backend.

---

## 4. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | GG-01: Research GGA config format (external docs) |
| 2 | GG-02/03/04: Implement MCP registration if format confirmed |
| 3 | GG-05: Add fallback download URL to update hint |

---

## David Kim CE — Coverage for `gga` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `gga`-specific deltas only.

| CE Topic | Status | `gga`-Specific Action |
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

## Context7 — Coverage for `gga` Agent

| Context7 Topic | Status | `gga`-Specific Action |
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

## Code Verification Notes for `gga`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `gga` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
