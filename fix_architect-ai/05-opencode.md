# Fix & Improvement Plan — `opencode` Agent

**Agent ID:** `opencode`
**Runtime:** OpenCode TUI (`opencode` binary)
**Config:** `opencode.json` (project-root) + `.atl/agents/*.md`
**Assets:** `internal/assets/opencode/`
**Go Adapter:** `internal/agents/opencode/adapter.go`
**Priority:** 🟠 High — Most complete MCP config of all agents (already has sequential-thinking + context-mode in JSON)

---

## 1. Current State

OpenCode is the most MCP-forward agent. Its `opencode.json` already includes:
```json
{
  "mcp": {
    "context7":           { "enabled": true, "type": "remote", "url": "https://mcp.context7.com/mcp" },
    "context-mode":       { "type": "local", "command": ["npx", "-y", "@mksglu/context-mode"], "enabled": true },
    "engram":             { "type": "local", "command": ["${ENGRAM_BIN}", "mcp", "--tools=agent"] },
    "sequential-thinking":{ "type": "local", "command": ["npx", "-y", "@modelcontextprotocol/server-sequential-thinking"], "enabled": true }
  }
}
```

This is the **reference MCP config**. Other agents need to converge on this pattern.

### sequential-thinking — Current State ✅
Registered with `npx -y @modelcontextprotocol/server-sequential-thinking`. Phase protocols mandate calls. **COMPLETE**.

### context-mode — Current State ✅ (mostly)
Registered via `@mksglu/context-mode` (note: different package name from `context-mode` binary).
**Gap**: Using `@mksglu/context-mode` (npm package) vs `context-mode` (global binary). Need version parity check.

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| OC-01 | `codegraph` absent from `opencode.json` MCP section | 🟠 High |
| OC-02 | `@mksglu/context-mode` vs `context-mode` binary — version drift risk | 🟡 Medium |
| OC-03 | `notebooklm-mcp` absent | 🟡 Medium |
| OC-04 | `sdd-overlay-multi.json` uses `background-agents.ts` plugin — not yet integrated with V4 tool_policy | 🟡 Medium |
| OC-05 | `persona-architect.md` present but not wired to `opencode.json` agent config | Low |
| OC-06 | No `codegraph init -i` post-install step | 🟠 High |

---

## 3. Fix Plan

### Fix OC-01 — Add codegraph

**File:** `internal/assets/opencode/opencode.json`

```json
"codegraph": {
  "type": "local",
  "command": ["npx", "-y", "@colbymchenry/codegraph", "serve", "--mcp"],
  "enabled": true
}
```

Post-install: `codegraph init -i --quiet` (run in project directory).

### Fix OC-02 — context-mode Package Consistency

**File:** `internal/assets/opencode/opencode.json`

Change to use global binary (consistent with other agents):
```json
"context-mode": {
  "type": "local",
  "command": ["context-mode", "--mcp"],
  "enabled": true
}
```

If `context-mode` binary not found, fall back to `npx -y @mksglu/context-mode`.

**Detection in adapter:**
```go
func (a *OpenCodeAdapter) ContextModeCommand() []string {
    if _, err := exec.LookPath("context-mode"); err == nil {
        return []string{"context-mode", "--mcp"}
    }
    return []string{"npx", "-y", "@mksglu/context-mode"}
}
```

### Fix OC-06 — codegraph init Post-install

**File:** `internal/agents/opencode/adapter.go`

```go
func (a *OpenCodeAdapter) PostInstall(projectDir string, dryRun bool) error {
    if !dryRun {
        cmd := exec.Command("codegraph", "init", "-i", "--quiet")
        cmd.Dir = projectDir
        cmd.Run()  // non-fatal if codegraph not installed
    }
    return nil
}
```

---

## 4. sequential-thinking Detection & Configuration

Already complete in OpenCode. Document the detection path for parity:

```go
// internal/agents/opencode/adapter.go
func (a *OpenCodeAdapter) VerifySequentialThinking() error {
    // Check 1: entry exists in opencode.json
    cfg := a.readOpenCodeJSON()
    if _, ok := cfg.MCP["sequential-thinking"]; !ok {
        return fmt.Errorf("sequential-thinking missing from opencode.json mcp section")
    }
    // Check 2: npx available
    if _, err := exec.LookPath("npx"); err != nil {
        return fmt.Errorf("npx not found — sequential-thinking requires Node.js/npm")
    }
    return nil
}
```

---

## 5. context-mode Detection & Configuration

```go
func (a *OpenCodeAdapter) VerifyContextMode() (string, error) {
    // Check binary
    out, err := exec.Command("context-mode", "--version").Output()
    if err != nil {
        // Try npm package
        out2, err2 := exec.Command("npx", "--yes", "@mksglu/context-mode", "--version").Output()
        if err2 != nil {
            return "", fmt.Errorf("context-mode not available: neither binary nor @mksglu/context-mode found")
        }
        return "@mksglu:" + strings.TrimSpace(string(out2)), nil
    }
    return strings.TrimSpace(string(out)), nil
}
```

---

## 6. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | OC-01: Add codegraph to opencode.json |
| 1 | OC-06: codegraph init post-install |
| 2 | OC-02: Normalize context-mode command |
| 2 | OC-04: Wire tool_policy.yaml to background-agents.ts |
| 3 | OC-03: notebooklm-mcp optional |

---

## David Kim CE — Coverage for `opencode` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `opencode`-specific deltas only.

| CE Topic | Status | `opencode`-Specific Action |
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

## Context7 — Coverage for `opencode` Agent

| Context7 Topic | Status | `opencode`-Specific Action |
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

## Code Verification Notes for `opencode`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `opencode` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
