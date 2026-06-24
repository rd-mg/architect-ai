# Fix & Improvement Plan — `codex` Agent

**Agent ID:** `codex`
**Runtime:** OpenAI Codex CLI (`codex` binary)
**Config:** `codex.toml` (project-root)
**Assets:** `internal/assets/codex/` (full set including `engram-compact-prompt.md`, `engram-instructions.md`)
**Go Adapter:** None
**Priority:** 🟡 Medium — Good asset coverage; TOML config format unique

---

## 1. Current State

Codex uses a `codex.toml` config. The assets include unique files:
- `engram-compact-prompt.md` — Codex-specific Engram compression instructions
- `engram-instructions.md` — Detailed Engram tool guidance

The TOML config has an MCP section:
```toml
[mcp_servers.context-mode]
command = "context-mode"
```

### sequential-thinking — Current State
- Phase protocols mandate it ✅
- **NOT in `codex.toml`** ❌

### context-mode — Current State
- `context-mode` entry in TOML ✅
- `context-mode` binary required to be pre-installed (no npx fallback in TOML)

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| CD-01 | sequential-thinking absent from `codex.toml` | 🔴 Critical |
| CD-02 | codegraph absent from `codex.toml` | 🟠 High |
| CD-03 | context-mode TOML entry has no fallback (no args, no env) | 🟡 Medium |
| CD-04 | engram-compact-prompt.md not referenced in other agents | Low |

---

## 3. Fix Plan

### Fix CD-01/02/03 — codex.toml MCP Section

```toml
[mcp_servers.engram]
command = "engram"
args = ["mcp", "--tools=agent"]

[mcp_servers.context7]
type = "http"
url = "https://mcp.context7.com/mcp"

[mcp_servers.context-mode]
command = "context-mode"
args = ["--mcp"]
env = {}

[mcp_servers.sequential-thinking]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-sequential-thinking"]

[mcp_servers.codegraph]
command = "npx"
args = ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
```

### Fix CD-04 — Share engram-compact-prompt.md

Move to `internal/assets/_shared/engram-compact-prompt.md` and reference from all agents.

---

## 4. sequential-thinking & context-mode Detection

```go
// internal/components/mcp/codex.go
func CodexTOMLMCPSection(available MCPAvailability) string {
    var b strings.Builder
    b.WriteString("[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\", \"--tools=agent\"]\n\n")
    if available.Context7 {
        b.WriteString("[mcp_servers.context7]\ntype = \"http\"\nurl = \"https://mcp.context7.com/mcp\"\n\n")
    }
    if available.ContextMode {
        b.WriteString("[mcp_servers.context-mode]\ncommand = \"context-mode\"\nargs = [\"--mcp\"]\n\n")
    }
    if available.SequentialThinking {
        b.WriteString("[mcp_servers.sequential-thinking]\ncommand = \"npx\"\nargs = [\"-y\", \"@modelcontextprotocol/server-sequential-thinking\"]\n\n")
    }
    if available.CodeGraph {
        b.WriteString("[mcp_servers.codegraph]\ncommand = \"npx\"\nargs = [\"-y\", \"@colbymchenry/codegraph\", \"serve\", \"--mcp\"]\n\n")
    }
    return b.String()
}
```

---

## 5. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | CD-01/02: Add sequential-thinking + codegraph to codex.toml |
| 2 | CD-03: Improve context-mode TOML with args and env |
| 3 | CD-04: Promote engram-compact-prompt.md to _shared |

---

## David Kim CE — Coverage for `codex` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `codex`-specific deltas only.

| CE Topic | Status | `codex`-Specific Action |
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

## Context7 — Coverage for `codex` Agent

| Context7 Topic | Status | `codex`-Specific Action |
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

## Code Verification Notes for `codex`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `codex` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
