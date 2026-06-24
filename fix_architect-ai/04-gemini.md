# Fix & Improvement Plan — `gemini` Agent

**Agent ID:** `gemini`
**Runtime:** Gemini CLI (`gemini` binary) / Gemini API
**Assets:** `internal/assets/gemini/`
**Go Adapter:** `internal/agents/gemini/adapter.go`
**Priority:** 🟠 High — Second-largest asset set; SDD phase protocols most mature after claude

---

## 1. Current State

Gemini has the most detailed SDD phase protocols of any non-Claude agent, including the
`sdd-phase-protocols/` subdirectory with all 9 phases. The Gemini adapter handles
`gemini mcp add --scope user` for MCP server registration.

### MCP Configuration (Current)
```bash
gemini mcp add --scope user engram engram mcp --tools=agent
gemini mcp add --scope user context7 --url https://mcp.context7.com/mcp
```
sequential-thinking: **absent** ❌
context-mode: **absent** ❌
codegraph: **absent** ❌

### sequential-thinking — Current State
- Phase protocols mandate `sequential_thinking` calls ✅
- MCP not registered via `gemini mcp add` ❌ — tool will 404

### context-mode — Current State
- `context-mode-routing-policy.md` injected in phase protocols ✅
- Not registered as an MCP server ❌
- Gemini CLI supports `hooks.json` in `.gemini/hooks/` — not configured by architect-ai ❌

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| G-01 | sequential-thinking not registered via `gemini mcp add` | 🔴 Critical |
| G-02 | context-mode not registered as MCP server | 🔴 Critical |
| G-03 | codegraph not in Gemini MCP config | 🟠 High |
| G-04 | `.gemini/hooks/` not populated by architect-ai | 🟠 High |
| G-05 | `adapter_metering.go` missing context-mode version probe | 🟡 Medium |
| G-06 | notebooklm-mcp absent | 🟡 Medium |

---

## 3. Fix Plan

### Fix G-01/G-02/G-03 — MCP Registration

**File:** `internal/agents/gemini/adapter.go`

Add to the `Install()` method:

```go
mcpServers := []struct{ name, cmd string; args []string }{
    {"engram",              "engram",   []string{"mcp", "--tools=agent"}},
    {"context7",            "",         nil},  // --url flag
    {"sequential-thinking", "npx",      []string{"-y", "@modelcontextprotocol/server-sequential-thinking"}},
    {"context-mode",        "context-mode", []string{"--mcp"}},
    {"codegraph",           "npx",      []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"}},
}
for _, s := range mcpServers {
    if s.cmd == "" {
        exec.Command("gemini", "mcp", "add", "--scope", "user", s.name, "--url",
            "https://mcp.context7.com/mcp").Run()
    } else {
        args := append([]string{"mcp", "add", "--scope", "user", s.name, s.cmd}, s.args...)
        exec.Command("gemini", args...).Run()
    }
}
```

### Fix G-04 — Gemini Hooks

Write `.gemini/hooks/context-mode.json`:
```json
{
  "preToolUse": [{
    "command": "context-mode hook gemini pretooluse",
    "matcher": "Shell|Read|Grep|WebFetch"
  }]
}
```

### Fix G-05 — Adapter Metering

Add to `adapter_metering.go`:
```go
{ Name: "context-mode", Check: func() string {
    out, _ := exec.Command("context-mode", "--version").Output()
    return strings.TrimSpace(string(out))
}},
{ Name: "sequential-thinking", Check: func() string {
    out, _ := exec.Command("npx", "--yes", "@modelcontextprotocol/server-sequential-thinking", "--version").Output()
    return strings.TrimSpace(string(out))
}},
```

---

## 4. sequential-thinking Detection & Configuration

```go
// Detection
func (a *GeminiAdapter) DetectSequentialThinking() bool {
    _, err := exec.LookPath("npx")
    return err == nil
}

// Registration
func (a *GeminiAdapter) RegisterSequentialThinking(dryRun bool) error {
    if dryRun {
        fmt.Println("  DRY RUN: gemini mcp add --scope user sequential-thinking npx -y @modelcontextprotocol/server-sequential-thinking")
        return nil
    }
    return exec.Command("gemini", "mcp", "add", "--scope", "user",
        "sequential-thinking", "npx", "-y",
        "@modelcontextprotocol/server-sequential-thinking").Run()
}
```

---

## 5. context-mode Detection & Configuration

```go
func (a *GeminiAdapter) DetectContextMode() (string, error) {
    out, err := exec.Command("context-mode", "--version").Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}

func (a *GeminiAdapter) RegisterContextMode(dryRun bool) error {
    if dryRun {
        fmt.Println("  DRY RUN: gemini mcp add --scope user context-mode context-mode --mcp")
        return nil
    }
    return exec.Command("gemini", "mcp", "add", "--scope", "user",
        "context-mode", "context-mode", "--mcp").Run()
}
```

---

## 6. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | G-01/02/03: Register sequential-thinking, context-mode, codegraph via gemini mcp add |
| 1 | G-04: Write `.gemini/hooks/context-mode.json` |
| 2 | G-05: Extend adapter_metering with new probes |
| 2 | Add codegraph steps to `sdd-explore.md` |
| 3 | G-06: notebooklm-mcp optional registration |

---

## David Kim CE — Coverage for `gemini` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `gemini`-specific deltas only.

| CE Topic | Status | `gemini`-Specific Action |
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

## Context7 — Coverage for `gemini` Agent

| Context7 Topic | Status | `gemini`-Specific Action |
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

## Code Verification Notes for `gemini`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `gemini` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
