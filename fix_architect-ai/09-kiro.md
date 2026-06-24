# Fix & Improvement Plan — `kiro` Agent

**Agent ID:** `kiro`
**Runtime:** Kiro (Amazon IDE)
**Config:** `.kiro/hooks/`, `.kiro/agents/`, `.kiro/settings.json`
**Assets:** `internal/assets/kiro/` (has `agents/` subdir with all 9 phases)
**Go Adapter:** `internal/agents/kiro/adapter.go`
**Priority:** 🟡 Medium — Has agents/ subdir (Kiro native format); no MCP config writer

---

## 1. Current State

Kiro uses a `.kiro/` directory with hooks, agents, and settings. architect-ai writes
`.kiro/hooks/context-mode.json` at install time (confirmed in inject.go). The `agents/`
subdir has all 9 SDD phases in Kiro's native format.

### sequential-thinking — Current State
- Phase protocols mandate `sequential_thinking` ✅
- **No MCP registration mechanism for Kiro in adapter.go** ❌
- Kiro MCP config format unknown/undocumented in architect-ai ❌

### context-mode — Current State
- Hook file written to `.kiro/hooks/context-mode.json` ✅
- context-mode binary install triggered ✅
- MCP server registration for Kiro: format not implemented ❌

---

## 2. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| KI-01 | Kiro MCP config format not implemented in adapter | 🔴 Critical |
| KI-02 | sequential-thinking not registered | 🔴 Critical |
| KI-03 | context-mode not registered as MCP server | 🟠 High |
| KI-04 | codegraph absent | 🟠 High |
| KI-05 | `errors.go` in adapter — unique error types suggest retry complexity | 🟡 Medium |
| KI-06 | No `thinking-agent.md` in assets | 🟡 Medium |

---

## 3. Fix Plan

### Fix KI-01/02/03/04 — Kiro MCP Config

Research Kiro's MCP config format. Based on `.kiro/settings.json` pattern:

```json
{
  "mcpServers": {
    "engram":              { "command": "engram", "args": ["mcp", "--tools=agent"] },
    "context7":            { "url": "https://mcp.context7.com/mcp" },
    "sequential-thinking": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"] },
    "context-mode":        { "command": "context-mode", "args": ["--mcp"] },
    "codegraph":           { "command": "npx", "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"] }
  }
}
```

Written to `.kiro/settings.json` or `.kiro/mcp_config.json` (verify with Kiro docs).

### Fix KI-06 — thinking-agent.md

Create `internal/assets/kiro/thinking-agent.md` (minimal):
Delegates to `_shared/adaptive-reasoning-gate-v2.md` + Kiro-specific tool availability note.

---

## 4–5. sequential-thinking & context-mode Detection

Same pattern as Windsurf — binary probe → npx fallback. Write to Kiro-specific config path.

---

## 6. Improvement Roadmap

| Week | Task |
|------|------|
| 1 | KI-01: Verify Kiro MCP config format, implement writer |
| 1 | KI-02/03/04: Register all MCPs |
| 2 | KI-06: thinking-agent.md |
| 3 | KI-05: Review errors.go retry logic for MCP failures |

---

## David Kim CE — Coverage for `kiro` Agent

| CE Topic | Status | kiro-Specific Action |
|----------|--------|---------------------|
| Protocol Shells | ❌ Missing | Add `/sdd.{phase}` header to all `.kiro/agents/` phase files |
| Token Budget Tracking | ❌ Missing | Add to `.atl/config.yaml` at install |
| Self-Refinement Engine | ❌ Missing | Add quality gate to `sdd-verify` agent file |
| Dynamic Assembly | ❌ Missing | Keyword-filter skill injection |
| Few-Shot Examples | ❌ Missing | Add to explore + design agent files |

> ⚠️ Kiro-specific: Verify that `agents/` subdir format supports Protocol Shell
> markdown headers — Kiro may parse YAML frontmatter only.

---

## Context7 — Coverage for `kiro` Agent

| Context7 Topic | Status | kiro-Specific Action |
|----------------|--------|---------------------|
| resolve-library-id + topic pattern | ❌ Missing | Add to `sdd-explore` agent file |
| Token cap | ❌ Missing | Add tokens: 5000 cap |

> ⚠️ VERIFY: Context7 MCP tool names against Kiro's MCP tool call syntax.

---

## Code Verification Notes for `kiro`

- Binary name/location for `kiro` IDE: verify at install time
- `.kiro/settings.json` and `.kiro/agents/` paths: verify against Kiro docs
- MCP config format: UNVERIFIED — research required before implementing
- All agent file format (YAML frontmatter + markdown): verify against Kiro docs
