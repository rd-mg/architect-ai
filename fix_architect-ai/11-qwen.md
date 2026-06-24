# Fix & Improvement Plan — `qwen` Agent

**Agent ID:** `qwen`
**Runtime:** Qwen (Alibaba) — Qwen-Coder via API/local
**Config:** `.atl/agents/` generic markdown files
**Assets:** `internal/assets/qwen/`
**Go Adapter:** `internal/agents/qwen/adapter.go`
**Priority:** 🟡 Medium — Only `sdd-orchestrator.md` in assets; limited MCP ecosystem

---

## 1. Current State

Qwen has `sdd-orchestrator.md` only. The Go adapter handles API-based installation
(no local IDE binary). Qwen's MCP support depends on the client wrapper used
(typically a REST API caller).

### sequential-thinking: ABSENT ❌
### context-mode: ABSENT ❌ (API-based agent — hook mechanism unclear)
### codegraph: ABSENT ❌

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| QW-01 | No MCP config mechanism (API-based agent) | 🟠 High |
| QW-02 | sequential-thinking: API doesn't natively support MCP | 🟠 High |
| QW-03 | context-mode: no hook mechanism for API-based agents | 🟡 Medium |
| QW-04 | codegraph: tool use depends on API capabilities | 🟡 Medium |
| QW-05 | No `thinking-agent.md` | 🟡 Medium |
| QW-06 | Phase protocols not in dedicated assets | 🟠 High |

---

## 3. Fix Plan

Qwen is API-based. MCP tools are injected as tool definitions in the API request,
NOT as separate MCP servers. architect-ai must:

1. **Translate MCP tool schemas → Qwen function-calling format** at dispatch time
2. Include `sequential_thinking` as a function definition (emulated — not real MCP)
3. For context-mode: use the context-mode binary to pre-process shell commands before
   injecting results into the Qwen API call

### Fix QW-02 — sequential-thinking Emulation

**File:** `internal/agents/qwen/adapter.go`

```go
// Inject sequential_thinking as a Qwen function definition
func qwenThinkingToolDef() map[string]any {
    return map[string]any{
        "type": "function",
        "function": map[string]any{
            "name": "sequential_thinking",
            "description": "Break down a complex problem into ordered thought steps",
            "parameters": map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "thought":     map[string]any{"type": "string"},
                    "nextThought": map[string]any{"type": "string"},
                    "thoughtNumber": map[string]any{"type": "integer"},
                    "totalThoughts": map[string]any{"type": "integer"},
                    "isRevision":    map[string]any{"type": "boolean"},
                    "branchId":      map[string]any{"type": "string"},
                },
            },
        },
    }
}
```

### Fix QW-03 — context-mode Pre-processing

For API-based agents, context-mode runs as a pre-processing step:
```go
func (a *QwenAdapter) ExecuteShell(cmd string) (string, error) {
    // Route through context-mode if available
    if contextModeAvailable() {
        return exec.Command("ctx", "execute", cmd).Output()
    }
    return exec.Command("sh", "-c", cmd).Output()
}
```

---

## 4. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | QW-02: sequential_thinking as emulated function definition |
| 2 | QW-05: thinking-agent.md |
| 2 | QW-06: Phase protocols (link to generic) |
| 3 | QW-03: context-mode pre-processing for shell calls |

---

## David Kim CE — Coverage for `qwen` Agent

> See `00-common-implementation-plan.md` §David Kim CE for full topic definitions.
> Items below are `qwen`-specific deltas only.

| CE Topic | Status | `qwen`-Specific Action |
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

## Context7 — Coverage for `qwen` Agent

| Context7 Topic | Status | `qwen`-Specific Action |
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

## Code Verification Notes for `qwen`

> All code in this file is **REFERENCE/PSEUDOCODE ONLY**.

- Binary names (e.g., `context-mode`, `codegraph`, `npx`) must be verified
  with `which <binary>` on a clean target machine
- npm package names must be verified with `npm view <package>`
- Config file paths (e.g., MCP config location) must be verified against
  vendor documentation for each `qwen` version
- JSON/YAML field names must be verified against each agent's current config schema
- Go function signatures must match the actual version in `go.mod`
- Test all changes with `architect-ai install --dry-run` first
