# Fix & Improvement Plan — `windsurf` Agent

**Agent ID:** `windsurf`
**Runtime:** Windsurf IDE (Codeium)
**Config:** `~/.codeium/windsurf/mcp_config.json`
**Assets:** `internal/assets/windsurf/`
**Go Adapter:** `internal/agents/windsurf/adapter.go`
**Priority:** 🟡 Medium — Only `sdd-orchestrator.md` in assets; thin coverage

---

## 1. Current State

Windsurf's MCP config is minimal. The adapter writes to `~/.codeium/windsurf/mcp_config.json`.
Assets: only `sdd-orchestrator.md` (no `thinking-agent.md`, no phase protocols in assets dir).

### sequential-thinking — Current State: ABSENT ❌
### context-mode — Current State: ABSENT ❌
### codegraph — Current State: ABSENT ❌

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| WS-01 | sequential-thinking absent from MCP config | 🔴 Critical |
| WS-02 | context-mode absent | 🔴 Critical |
| WS-03 | codegraph absent | 🟠 High |
| WS-04 | No `thinking-agent.md` for Windsurf | 🟠 High |
| WS-05 | No phase protocols in `internal/assets/windsurf/` | 🟠 High |
| WS-06 | `notebooklm-mcp` absent | Low |

---

## 3. Fix Plan

### Fix WS-01/02/03 — MCP Config

**File:** `internal/agents/windsurf/adapter.go`

Write to `~/.codeium/windsurf/mcp_config.json`:
```json
{
  "mcpServers": {
    "engram": {
      "command": "engram",
      "args": ["mcp", "--tools=agent"]
    },
    "context7": {
      "serverUrl": "https://mcp.context7.com/mcp"
    },
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
}
```

### Fix WS-04/05 — Thin Asset Coverage

Windsurf uses the generic/shared assets for phase protocols. Explicitly document this:

```go
// In adapter.go — Windsurf uses generic phase protocols via skill-registry
// No Windsurf-specific phase protocol files needed (YAGNI)
// thinking-agent.md: copy from _shared/adaptive-reasoning-gate-v2.md
```

Add `thinking-agent.md` for Windsurf (minimal — delegate to generic):
```markdown
---
name: windsurf-thinking-agent
model: inherit
---
{content of _shared/adaptive-reasoning-gate-v2.md}
{content of _shared/general-orchestrator-identity.md}
```

---

## 4. sequential-thinking Detection & Configuration

```go
func (a *WindsurfAdapter) Install(dryRun bool) error {
    available := struct {
        npx         bool
        contextMode bool
        codegraph   bool
    }{
        npx:         checkBinary("npx"),
        contextMode: checkBinary("context-mode"),
        codegraph:   checkBinary("codegraph") || checkBinary("npx"),
    }
    
    cfg := buildWindsurfMCPConfig(available)
    return writeJSON("~/.codeium/windsurf/mcp_config.json", cfg, dryRun)
}
```

---

## 5. context-mode Detection & Configuration

Same pattern as OpenCode — global binary or npx fallback:
```go
func windsurfContextModeCommand() []string {
    if _, err := exec.LookPath("context-mode"); err == nil {
        return []string{"context-mode", "--mcp"}
    }
    return []string{"npx", "-y", "@mksglu/context-mode"}
}
```

---

## 6. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | WS-01/02/03: Full MCP config in adapter.go |
| 2 | WS-04: thinking-agent.md (thin, delegates to _shared) |
| 2 | WS-05: Link generic phase protocols explicitly |
| 3 | WS-06: notebooklm optional |

---

## David Kim CE — Coverage for `windsurf` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `windsurf`-specific deltas only.

| CE Topic | Status | `windsurf`-Specific Action |
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

## Context7 — Coverage for `windsurf` Agent

| Context7 Topic | Status | `windsurf`-Specific Action |
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

## Code Verification Notes for `windsurf`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `windsurf` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
