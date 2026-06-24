# Fix & Improvement Plan — `kilocode` Agent

**Agent ID:** `kilocode`
**Runtime:** Kilocode (VS Code extension)
**Config:** `.kilocode/mcp.json` or VS Code settings
**Assets:** Generic (no dedicated `internal/assets/kilocode/` dir)
**Go Adapter:** `internal/agents/kilocode/adapter.go` + `paths.go`
**Priority:** 🟡 Medium — Go adapter exists; no dedicated assets (uses generic)

---

## 1. Current State

Kilocode has a Go adapter and `paths.go` (indicating path-resolution logic) but NO
dedicated assets directory. It falls back to generic templates. `adapter_metering.go`
exists, suggesting active monitoring.

### sequential-thinking: ABSENT ❌
### context-mode: ABSENT ❌ (no hook file written per inject.go search)
### codegraph: ABSENT ❌

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| KC-01 | No dedicated assets — falls back to generic with no customization | 🟠 High |
| KC-02 | MCP config not written by adapter | 🔴 Critical |
| KC-03 | sequential-thinking absent | 🔴 Critical |
| KC-04 | context-mode absent | 🟠 High |
| KC-05 | codegraph absent | 🟠 High |
| KC-06 | `paths.go` path resolution not validated against Kilocode's actual config locations | 🟡 Medium |

---

## 3. Fix Plan

### Fix KC-02/03/04/05 — MCP Config

Kilocode is a VS Code extension — uses `.kilocode/mcp.json` or VS Code `settings.json`.

**File:** `internal/agents/kilocode/adapter.go`

```go
var kilococeMCPJSON = []byte(`{
  "mcpServers": {
    "engram":              { "command": "engram", "args": ["mcp", "--tools=agent"] },
    "context7":            { "url": "https://mcp.context7.com/mcp" },
    "sequential-thinking": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"] },
    "context-mode":        { "command": "context-mode", "args": ["--mcp"] },
    "codegraph":           { "command": "npx", "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"] }
  }
}`)

func (a *KilocodeAdapter) Install(projectDir string, dryRun bool) error {
    cfgPath := filepath.Join(projectDir, ".kilocode", "mcp.json")
    return writeJSON(cfgPath, kilocodeMCPJSON, dryRun)
}
```

### Fix KC-01 — Minimal Dedicated Assets

Create `internal/assets/kilocode/` with:
- `sdd-orchestrator.md` — copy of generic with Kilocode tool-availability note
- No phase protocols needed (use generic via skill-registry)

---

## 4–5. sequential-thinking & context-mode Detection

```go
func (a *KilocodeAdapter) detectMCPs() map[string]bool {
    return map[string]bool{
        "npx":          checkBinary("npx"),
        "context-mode": checkBinary("context-mode"),
        "codegraph":    checkBinary("codegraph") || checkBinary("npx"),
    }
}
```

---

## 6. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | KC-02/03/04/05: MCP config writer in adapter |
| 2 | KC-01: Minimal assets dir |
| 3 | KC-06: Validate paths.go against actual Kilocode config locations |

---

## David Kim CE — Coverage for `kilocode` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `kilocode`-specific deltas only.

| CE Topic | Status | `kilocode`-Specific Action |
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

## Context7 — Coverage for `kilocode` Agent

| Context7 Topic | Status | `kilocode`-Specific Action |
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

## Code Verification Notes for `kilocode`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `kilocode` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
