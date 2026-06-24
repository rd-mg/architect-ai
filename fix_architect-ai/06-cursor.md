# Fix & Improvement Plan — `cursor` Agent

**Agent ID:** `cursor`
**Runtime:** Cursor IDE
**Config:** `.cursor/mcp.json`, `.cursor/rules/`, `.cursor/agents/`
**Assets:** `internal/assets/cursor/` (has both `agents/` and `sdd-phase-protocols/` subdirs)
**Go Adapter:** None (config files written directly)
**Priority:** 🟠 High — Dual asset structure (agents + phase-protocols) indicates partial migration

---

## 1. Current State

Cursor has a unique dual structure: `agents/` (for Cursor's native agent format) AND
`sdd-phase-protocols/` (for the architect-ai format). This suggests a legacy migration
path that was started but not completed.

Cursor supports MCP via `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "engram": { "command": "engram", "args": ["mcp", "--tools=agent"] },
    "context7": { "url": "https://mcp.context7.com/mcp" }
  }
}
```

### sequential-thinking — Current State
- Phase protocols mandate `sequential_thinking` ✅
- **NOT registered in `.cursor/mcp.json`** ❌

### context-mode — Current State
- Hook config written to `.cursor/hooks/context-mode.json` by installer ✅
  (`"preToolUse": [{ "command": "context-mode hook cursor pretooluse", ... }]`)
- context-mode NOT registered as MCP server in `.cursor/mcp.json` ❌
- Hook binary existence validated at install ✅

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| CU-01 | sequential-thinking absent from `.cursor/mcp.json` | 🔴 Critical |
| CU-02 | context-mode not in `.cursor/mcp.json` as MCP server | 🟠 High |
| CU-03 | codegraph absent | 🟠 High |
| CU-04 | Dual `agents/` + `sdd-phase-protocols/` structure — maintenance burden | 🟡 Medium |
| CU-05 | `thinking-agent.md` present but not linked in `.cursor/agents/` | 🟡 Medium |
| CU-06 | `context-mode hook cursor pretooluse` — hook format not verified against Cursor version | 🟡 Medium |

---

## 3. Fix Plan

### Fix CU-01/02/03 — MCP Config

**File:** `internal/assets/cursor/mcp.json` template

```json
{
  "mcpServers": {
    "engram": {
      "command": "engram",
      "args": ["mcp", "--tools=agent"]
    },
    "context7": {
      "url": "https://mcp.context7.com/mcp"
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

Written to `.cursor/mcp.json` at project root.

### Fix CU-04 — Consolidate Dual Structure

Deprecate `internal/assets/cursor/agents/` — migrate content into `sdd-phase-protocols/`.
The `agents/` subdirectory was Cursor's legacy format; modern Cursor uses standard MCP + rules.

### Fix CU-06 — Hook Format Verification

**File:** `internal/system/deps.go`

Add Cursor version check:
```go
func CursorVersion() string {
    out, _ := exec.Command("cursor", "--version").Output()
    return strings.TrimSpace(string(out))
}
// Validate hook format compatibility for detected version
```

---

## 4. sequential-thinking Detection & Configuration

```go
// In cursor config writer:
func cursorMCPConfig(npxAvailable, cmAvailable, cgAvailable bool) []byte {
    servers := map[string]any{
        "engram":   localServer("engram", []string{"mcp", "--tools=agent"}),
        "context7": remoteServer("https://mcp.context7.com/mcp"),
    }
    if npxAvailable {
        servers["sequential-thinking"] = localServer("npx",
            []string{"-y", "@modelcontextprotocol/server-sequential-thinking"})
    }
    if cmAvailable {
        servers["context-mode"] = localServer("context-mode", []string{"--mcp"})
    }
    if cgAvailable {
        servers["codegraph"] = localServer("npx",
            []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"})
    }
    cfg, _ := json.MarshalIndent(map[string]any{"mcpServers": servers}, "", "  ")
    return cfg
}
```

---

## 5. context-mode Detection & Configuration

Cursor uses BOTH:
1. Hook file (`.cursor/hooks/context-mode.json`) — existing, intercepts tool calls
2. MCP server entry — new, adds `ctx_*` tools to agent tool list

Both must be present for full context-mode functionality.

```go
func installCursorContextMode(projectDir string, dryRun bool) error {
    // 1. Hook file
    hookPath := filepath.Join(projectDir, ".cursor", "hooks", "context-mode.json")
    writeFile(hookPath, cursorContextModeHookJSON, dryRun)
    
    // 2. MCP entry — added to .cursor/mcp.json
    // (handled by cursorMCPConfig above)
    return nil
}
```

---

## 6. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | CU-01/02/03: Update `.cursor/mcp.json` template |
| 2 | CU-04: Deprecate `agents/` subdir |
| 2 | CU-05: Wire `thinking-agent.md` to cursor rules |
| 3 | CU-06: Hook format version check |

---

## David Kim CE — Coverage for `cursor` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `cursor`-specific deltas only.

| CE Topic | Status | `cursor`-Specific Action |
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

## Context7 — Coverage for `cursor` Agent

| Context7 Topic | Status | `cursor`-Specific Action |
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

## Code Verification Notes for `cursor`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `cursor` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
