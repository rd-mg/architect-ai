# Fix & Improvement Plan — `vscode` Agent

**Agent ID:** `vscode`
**Runtime:** VS Code + GitHub Copilot
**Config:** `.vscode/mcp.json`, `.github/copilot-instructions.md`
**Assets:** `internal/assets/vscode/`
**Go Adapter:** `internal/agents/vscode/adapter.go`
**Priority:** 🟠 High — Only agent with HTTP MCP transport; has general-orchestrator.md

---

## 1. Current State

VS Code uses HTTP-type MCP servers (native VS Code MCP format):
```json
{
  "servers": {
    "context7": { "type": "http", "url": "https://mcp.context7.com/mcp" },
    "engram":   { "type": "http", "url": "http://localhost:3000/mcp" }
  }
}
```

The `adapter.go` handles `vsCodeContext7OverlayJSON` (written to `.vscode/mcp.json`).

### sequential-thinking — Current State
- Phase protocols reference `sequential_thinking` ✅
- **NOT in `.vscode/mcp.json`** ❌
- VS Code MCP does NOT support stdio-type servers in `.vscode/mcp.json` (HTTP only)
- Sequential-thinking server has no HTTP transport — requires stdio wrapper ❌

### context-mode — Current State
- Hook config written to `.github/hooks/context-mode.json` ✅
  (`"preToolUse": [{ "command": "context-mode hook cursor pretooluse", "matcher": "..." }]`)
- context-mode NOT in `.vscode/mcp.json` ❌
- context-mode needs an HTTP wrapper for VS Code MCP ❌

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| VS-01 | sequential-thinking has no HTTP transport for VS Code MCP | 🔴 Critical |
| VS-02 | context-mode has no HTTP transport for VS Code MCP | 🔴 Critical |
| VS-03 | codegraph absent from `.vscode/mcp.json` | 🟠 High |
| VS-04 | `.github/copilot-instructions.md` lacks KEYWORD ROUTING (E-08 check) | 🟠 High |
| VS-05 | `general-orchestrator.md` is present but the general-orchestrator is not in `.vscode/mcp.json` as an agent | 🟡 Medium |
| VS-06 | codegraph serve supports `--mcp` flag with HTTP option — untested for VS Code | 🟡 Medium |

---

## 3. Fix Plan

### Fix VS-01/VS-02 — HTTP Wrappers

Create a thin HTTP proxy that wraps stdio MCP servers for VS Code:

**File:** `internal/components/tmux/http_proxy.go` (or new `internal/components/mcp/http_proxy.go`)

```go
// MCPHTTPProxy wraps a stdio MCP server command as an HTTP server on a local port
type MCPHTTPProxy struct {
    Command string
    Args    []string
    Port    int
}

func (p *MCPHTTPProxy) Start() error {
    // Start subprocess, pipe stdio to HTTP endpoint at localhost:Port
}
```

**VS Code MCP config with HTTP proxies:**
```json
{
  "servers": {
    "context7":            { "type": "http", "url": "https://mcp.context7.com/mcp" },
    "engram":              { "type": "http", "url": "http://localhost:3000/mcp" },
    "sequential-thinking": { "type": "http", "url": "http://localhost:3001/mcp" },
    "context-mode":        { "type": "http", "url": "http://localhost:3002/mcp" },
    "codegraph":           { "type": "http", "url": "http://localhost:7340/mcp" }
  }
}
```

**Launcher:** `architect-ai mcp-proxy start` — starts all proxy processes, writes ports to `.atl/mcp-ports.json`.

### Fix VS-03 — codegraph in VS Code

`codegraph serve --mcp --transport http --port 7340` — codegraph natively supports HTTP transport.
Add to VS Code MCP config:
```json
"codegraph": { "type": "http", "url": "http://localhost:7340/mcp" }
```

Install hook: `codegraph serve --mcp --transport http --port 7340 --daemon`.

### Fix VS-04 — KEYWORD ROUTING in copilot-instructions

**File:** `internal/assets/vscode/sdd-orchestrator.md` (used as copilot-instructions content)

Add KEYWORD ROUTING block (satisfying E-08):
```markdown
## KEYWORD ROUTING

Analyze the first user message. Match against:
- SDD_KEYWORDS: ["use sdd", "sdd-", "/sdd", "software design", "architecture change"]
  → Route to SDD Orchestrator sub-agent
- GENERAL_KEYWORDS: [default]
  → Route to General Orchestrator

DO NOT invoke both orchestrators for the same request.
```

---

## 4. sequential-thinking Detection & Configuration

VS Code-specific challenge: stdio MCP not natively supported.

**Strategy A (Preferred):** HTTP proxy wrapper
```go
func VSCodeSequentialThinkingConfig(port int) map[string]any {
    return map[string]any{
        "type": "http",
        "url":  fmt.Sprintf("http://localhost:%d/mcp", port),
    }
}
// Port written to .atl/mcp-ports.json after proxy start
```

**Strategy B (Fallback):** Ship a small Node.js HTTP-to-stdio bridge:
```javascript
// tools/seq-thinking-http-bridge.js
const { createServer } = require('http');
const { spawn } = require('child_process');
// Bridge HTTP ↔ stdio for @modelcontextprotocol/server-sequential-thinking
```

**Detection:**
```go
func DetectVSCodeMCPRequirements() VSCodeMCPStatus {
    return VSCodeMCPStatus{
        CanRunHTTP:          checkPort(3000),  // engram already on 3000
        CodeGraphHTTPReady:  checkBinary("codegraph") && checkPort(7340),
        ProxyDaemonRunning:  checkPIDFile(".atl/mcp-proxy.pid"),
    }
}
```

---

## 5. context-mode Detection & Configuration

context-mode needs HTTP transport for VS Code. Two options:

**Option 1:** `context-mode --mcp --transport http --port 3002` (if supported by context-mode version)
```go
out, _ := exec.Command("context-mode", "--help").Output()
httpSupported := strings.Contains(string(out), "--transport")
```

**Option 2:** HTTP proxy (same as sequential-thinking proxy, port 3002).

Both options verified at install time. If neither available, VS Code agent runs without context-mode MCP (degraded — relies only on hook file in `.github/hooks/`).

---

## 6. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | VS-03: codegraph HTTP in `.vscode/mcp.json` |
| 1 | VS-04: KEYWORD ROUTING in sdd-orchestrator.md |
| 2 | VS-01: sequential-thinking HTTP proxy |
| 2 | VS-02: context-mode HTTP proxy or HTTP transport |
| 3 | VS-05: Wire general-orchestrator as VS Code agent |
| 3 | `architect-ai mcp-proxy start` command |

---

## David Kim CE — Coverage for `vscode` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `vscode`-specific deltas only.

| CE Topic | Status | `vscode`-Specific Action |
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

## Context7 — Coverage for `vscode` Agent

| Context7 Topic | Status | `vscode`-Specific Action |
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

## Code Verification Notes for `vscode`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `vscode` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
