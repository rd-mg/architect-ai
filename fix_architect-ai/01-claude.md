# Fix & Improvement Plan — `claude` Agent

**Agent ID:** `claude`  
**Runtime:** Claude Code CLI / Anthropic API  
**Assets:** `internal/assets/claude/`  
**Go Adapter:** None (direct CLI install via `claude` binary)  
**Priority:** 🔴 Critical — Primary reference implementation; all other agents diverge from this baseline

---

## 1. Current State

Claude is the reference agent. It has the most complete asset set:
- `sdd-orchestrator.md` — full 9-phase pipeline
- `thinking-agent.md` — L0 super-orchestrator with D1–D4 classifier
- `persona-architect.md` — identity block

### MCP Configuration (Current)

Claude Code reads MCP config from `~/.claude/claude_code_config.json` (or scoped via `--scope`). architect-ai's installer writes:

```json
{
  "mcpServers": {
    "engram":              { "command": "engram", "args": ["mcp", "--tools=agent"] },
    "context7":            { "type": "http", "url": "https://mcp.context7.com/mcp" },
    "sequential-thinking": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"] }
  }
}
```

**context-mode** is wired separately via a hook file injected by `EnsureContextMode` at install time. The hook intercepts shell/read/grep tool calls and routes them through `context-mode` binary.

### sequential-thinking — Current State

- Registered in MCP config ✅  
- `thinking-agent.md` mandates `sequential_thinking` call BEFORE any tool use ✅  
- `sdd-design.md` mandates `sequential_thinking` with `≥2 branchId` for arch alternatives ✅  
- `sdd-explore.md` mandates `sequential_thinking` BEFORE search tools ✅  
- **Gap**: No graceful fallback prompt when `sequential_thinking` tool is unavailable (e.g. npm/npx offline). Sub-agent silently skips it.

### context-mode — Current State

- Installed via `npm install -g context-mode` at install time ✅  
- Hook file written to `.claude/hooks/context-mode.json` ✅  
- `context-mode-routing-policy.md` injected into all phase protocols ✅  
- **Gap 1**: Version detection (`context-mode --version`) is the only health check; no `ctx doctor` run in CI or post-install  
- **Gap 2**: `ctx_batch_execute` not referenced in any Claude phase protocol. `sdd-explore` still runs sequential ripgrep calls instead of batching  
- **Gap 3**: No per-session check that context-mode hooks are active (hooks can be disabled by Claude Code update)

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| C-01 | No `sequential_thinking` fallback when npx offline | Medium |
| C-02 | `ctx_batch_execute` unused in explore/verify phases | High |
| C-03 | No post-install `ctx doctor` health check | Medium |
| C-04 | Hook liveness not verified at session start | Medium |
| C-05 | `codegraph_*` tools absent from MCP config | High |
| C-06 | `notebooklm-mcp` absent from default MCP config (optional but undocumented) | Low |
| C-07 | `persona-architect.md` not referenced in `sdd-orchestrator.md` header | Low |
| C-08 | `tool_policy.yaml` mechanism absent (V4 requirement) | Medium |

---

## 3. Fix Plan

### Fix C-01 — sequential-thinking Fallback

**File:** `internal/assets/claude/thinking-agent.md`

Add after the mandatory `sequential_thinking` call block:

```markdown
> **Fallback (if sequential_thinking tool unavailable):**  
> Explicitly write a <thinking> block with:  
> (a) problem restatement, (b) D1-D4 classification, (c) 3 alternative approaches,  
> (d) selected approach + rationale.  
> This satisfies the sequential-thinking contract without the MCP tool.
```

### Fix C-02 — Batch Execute in Explore

**File:** `internal/assets/claude/sdd-phase-protocols/sdd-explore.md`

Replace sequential ripgrep calls:
```markdown
<!-- BEFORE -->
- rg "func Auth" --type go
- rg "AuthHandler" -l
- rg "^type.*Handler" --type go

<!-- AFTER -->
- ctx_batch_execute([
    "rg 'func Auth' --type go",
    "rg 'AuthHandler' -l",
    "rg '^type.*Handler' --type go"
  ])
  → single context-mode call, compressed output, no flooding
```

Apply same pattern to `sdd-verify.md` blast-radius checks.

### Fix C-03 — Post-install ctx doctor

**File:** `internal/app/install_cmd.go` (or `internal/components/engram/download.go`)

```go
// After EnsureContextMode returns success:
if _, err := exec.Command("ctx", "doctor").Output(); err != nil {
    fmt.Fprintf(stderr, "WARNING: ctx doctor failed — context-mode may not be fully configured\n")
    fmt.Fprintln(stderr, "  Run: ctx doctor --verbose")
}
```

### Fix C-04 — Hook Liveness Check

**File:** `internal/assets/claude/thinking-agent.md` (session startup sequence)

Add as Step 0 in tool availability probe:
```markdown
0. **Hook liveness**: confirm context-mode hook is active.  
   Execute a tiny shell command and verify it routes through ctx:  
   `ctx doctor` OR check `.claude/hooks/context-mode.json` exists.  
   If missing: surface warning, continue with raw tool calls (degraded mode).
```

### Fix C-05 — CodeGraph MCP Config

**File:** `internal/components/mcp/codegraph.go` (new, see blueprint §3.7)

Add to Claude MCP config writer:
```json
{
  "mcpServers": {
    "codegraph": {
      "command": "npx",
      "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
    }
  }
}
```

Install condition: `codegraph` binary present OR `npx` available.

### Fix C-08 — Tool Policy YAML

**File:** `.atl/tool_policy.yaml` (generated at `architect-ai install` time)

```yaml
pre_tool_use:
  - matcher: "Bash|Shell"
    decision: "ask"
    condition: "phase == sdd-apply AND mode != tmux"
  - matcher: "mcp__codegraph__*"
    decision: "allow"
  - matcher: "mcp__engram__*"
    decision: "allow"
  - matcher: "WebFetch|WebSearch"
    decision: "ask"
    condition: "posture == production"
```

---

## 4. sequential-thinking Detection & Configuration

### Detection Logic (Go)

```go
// internal/components/mcp/sequential_thinking.go
func DetectSequentialThinking() (bool, error) {
    // Check 1: npx available
    if _, err := exec.LookPath("npx"); err != nil {
        return false, fmt.Errorf("npx not found — sequential-thinking requires Node.js")
    }
    // Check 2: package resolvable (offline check via npm cache)
    out, _ := exec.Command("npm", "list", "-g",
        "@modelcontextprotocol/server-sequential-thinking").Output()
    if strings.Contains(string(out), "empty") {
        return false, nil  // not pre-installed; will run via npx -y (online)
    }
    return true, nil
}
```

### MCP Config Block (Claude)

```json
"sequential-thinking": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"],
  "env": {}
}
```

**Scope:** Written to `~/.claude/claude_code_config.json` at `--scope user`.

---

## 5. context-mode Detection & Configuration

### Detection Logic (Go) — existing + improvements

```go
// internal/components/engram/download.go — extend EnsureContextMode
func DetectContextMode() ContextModeCapability {
    // Check binary
    out, err := exec.Command("context-mode", "--version").Output()
    if err != nil {
        return ContextModeCapability{Installed: false}
    }
    version := strings.TrimSpace(string(out))
    
    // Check hook file exists
    hookPath := filepath.Join(os.Getenv("HOME"), ".claude", "hooks", "context-mode.json")
    _, hookErr := os.Stat(hookPath)
    
    // Check doctor (new)
    doctorOut, doctorErr := exec.Command("ctx", "doctor").Output()
    
    return ContextModeCapability{
        Installed:   true,
        Version:     version,
        HookActive:  hookErr == nil,
        DoctorClean: doctorErr == nil && !strings.Contains(string(doctorOut), "ERROR"),
    }
}
```

### Hook Config (Claude Code)

Written to `.claude/hooks/context-mode.json` (per-project) or `~/.claude/hooks/` (global):

```json
{
  "preToolUse": [{
    "command": "context-mode hook claude pretooluse",
    "matcher": "Shell|Read|Grep|WebFetch|Task|MCP:ctx_execute|MCP:ctx_execute_file|MCP:ctx_batch_execute"
  }]
}
```

---

## 6. Improvement Roadmap

| Week | Task | Files |
|------|------|-------|
| 1 | C-02: batch execute in explore | `sdd-explore.md`, `sdd-verify.md` |
| 1 | C-05: codegraph MCP config | `internal/components/mcp/codegraph.go` |
| 2 | C-01: sequential-thinking fallback | `thinking-agent.md` |
| 2 | C-03: ctx doctor post-install | `internal/app/install_cmd.go` |
| 3 | C-04: hook liveness check | `thinking-agent.md` |
| 3 | C-08: tool_policy.yaml | `internal/app/skills_cmd.go` |
| 4 | Add codegraph steps to `sdd-explore.md` | Phase A: Mode A exploration |

---

## David Kim CE — Coverage for `claude` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `claude`-specific deltas only.

| CE Topic | Status | `claude`-Specific Action |
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

## Context7 — Coverage for `claude` Agent

| Context7 Topic | Status | `claude`-Specific Action |
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

## Code Verification Notes for `claude`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `claude` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
