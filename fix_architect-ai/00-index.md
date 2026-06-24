# architect-ai V4 — Per-Agent Fix & Improvement Plans Index

**Date:** 2026-06-22  
**Scope:** All coding agents in architect-ai V3.3.0 → V4 upgrade  
**Guiding principle:** Pareto/YAGNI — 20% changes, 80% value  

---

## Key Finding: Antigravity CLI vs IDE Are Separate Agents

The Antigravity documentation confirms two distinct products with different hook formats,
config structures, and capabilities. Current architect-ai merges them incorrectly.

| Feature | Antigravity CLI (`agy`) | Antigravity IDE (extension) |
|---------|------------------------|----------------------------|
| Hook output format | `decision/reason/permissionOverrides` | `{}` (empty) |
| Config root | `~/.gemini/antigravity-cli/` | `mcp_config.json` + `.agents/` |
| MCP registration | Plugin `mcp_config.json` | `mcp_config.json` with OAuth |
| Session management | `/fork`, `/branch`, `agy <uuid>` | IDE-managed |
| Parallelism | Session forks (user-driven) | Sidecars (process-level) |
| Context window | 1M tokens (Gemini) | 1M tokens (Gemini) |
| Unique feature | Artifact picker TUI, status line | `ephemeralMessage` injection, Stop hook |
| architect-ai coverage | ❌ ZERO (critical gap) | ✅ Partial (wrong hook format) |

---

## MCP Universal Status Across All Agents

| MCP | Required By | Claude | Ant-IDE | Ant-CLI | Gemini | OpenCode | Cursor | VSCode | Windsurf | Kiro | Kilocode | Qwen | Codex | GGA | Generic |
|-----|------------|--------|---------|---------|--------|----------|--------|--------|----------|------|----------|------|-------|-----|---------|
| engram | All | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | — |
| context7 | All | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | — |
| **sequential-thinking** | All | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | — |
| **context-mode** | All | ✅ hook | ❌ | ❌ | ✅ hook | ✅ MCP | ✅ hook | ✅ hook | ❌ | ✅ hook | ❌ | ❌ | ✅ TOML | ❌ | — |
| **codegraph** | All | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | — |
| notebooklm | Optional | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | — |

**Legend:** ✅ = implemented, ❌ = missing, — = N/A (no IDE config), hook = via hook file only (not MCP tool), MCP = as MCP server tool

---

## Agent Plans

| # | File | Agent | Priority | Critical Gaps |
|---|------|-------|---------|---------------|
| 01 | `01-claude.md` | Claude | 🔴 | C-02 batch execute, C-05 codegraph |
| 02 | `02-antigravity-ide.md` | Antigravity IDE | 🟠 | AI-01 wrong hook format, AI-02 no seq-think |
| 03 | `03-antigravity-cli.md` | Antigravity CLI | 🟠 | AC-01 zero assets, AC-02 no plugin installer |
| 04 | `04-gemini.md` | Gemini | 🟠 | G-01 seq-think not registered, G-02 ctx-mode not registered |
| 05 | `05-opencode.md` | OpenCode | 🟠 | OC-01 codegraph absent (otherwise best MCP coverage) |
| 06 | `06-cursor.md` | Cursor | 🟠 | CU-01 seq-think absent, CU-02 ctx-mode MCP absent |
| 07 | `07-vscode.md` | VS Code | 🟠 | VS-01 seq-think no HTTP transport, VS-02 ctx-mode HTTP gap |
| 08 | `08-windsurf.md` | Windsurf | 🟡 | WS-01/02/03 all MCPs absent |
| 09 | `09-kiro.md` | Kiro | 🟡 | KI-01 MCP format unknown, KI-02/03/04 all absent |
| 10 | `10-kilocode.md` | Kilocode | 🟡 | KC-01 no assets, KC-02/03/04 MCPs absent |
| 11 | `11-qwen.md` | Qwen | 🟡 | QW-01/02 API-based — no native MCP; emulation needed |
| 12 | `12-codex.md` | Codex | 🟡 | CD-01 seq-think absent from TOML, CD-02 codegraph absent |
| 13 | `13-gga.md` | GGA | 🟡 | GG-01 MCP config format unknown |
| 14 | `14-generic.md` | Generic (fallback) | 🟢 | GEN-02 no batch execute, GEN-03 no codegraph |

---

## Global Fix Priority Queue (cross-agent Pareto)

### Sprint 1 — Week 1-2 (Critical fixes, all agents)

1. **Sequential-thinking MCP registration** — 12 of 14 agents missing → single config
   template, applied per-agent by config writers
2. **Antigravity split** — rename `antigravity` → `antigravity-ide`, create
   `antigravity-cli` from scratch (AC-01 through AC-10)
3. **Fix IDE hook format** (AI-01) — currently writes CLI format to IDE → silent breakage
4. **codegraph MCP** — zero agents have it; add to all config writers simultaneously
5. **`ctx_batch_execute` in sdd-explore** — generic + all agents pull from generic →
   one change propagates to all

### Sprint 2 — Week 3-4 (High-value improvements)

6. **VS Code HTTP proxy** for stdio MCPs (seq-think, ctx-mode) — unique to VS Code
7. **Gemini MCP registration** via `gemini mcp add` CLI
8. **OpenCode codegraph** + `codegraph init` post-install
9. **CodeGraph exploration mode** in all `sdd-explore.md` files

### Sprint 3 — Week 5-6 (Completion)

10. Kiro/Kilocode/Windsurf MCP configs (low user volume, high fix ROI)
11. Qwen emulated `sequential_thinking` function definition
12. GGA config format research + implementation
13. Sidecar configs for archive cleanup (Antigravity CLI + generic)

---

## sequential-thinking MCP — Universal Reference

All agents should use this server definition (adapted to their config format):

```
Binary: npx -y @modelcontextprotocol/server-sequential-thinking
Detection: exec.LookPath("npx") + npm cache check
Fallback: explicit <thinking> block in prompt when tool unavailable
Install: npm install -g @modelcontextprotocol/server-sequential-thinking (optional pre-cache)
```

| Config Format | Entry |
|---|---|
| JSON mcpServers | `"sequential-thinking": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"] }` |
| TOML | `[mcp_servers.sequential-thinking]\ncommand="npx"\nargs=["-y","@modelcontextprotocol/server-sequential-thinking"]` |
| VS Code HTTP | `"sequential-thinking": { "type": "http", "url": "http://localhost:3001/mcp" }` (via proxy) |
| Gemini CLI | `gemini mcp add --scope user sequential-thinking npx -y @modelcontextprotocol/server-sequential-thinking` |

---

## context-mode MCP — Universal Reference

```
Binary: context-mode (npm install -g context-mode)
NPX fallback: npx -y @mksglu/context-mode
Detection: exec.Command("context-mode", "--version")
MCP mode: context-mode --mcp
Hook mode: context-mode hook <agent> pretooluse (written to agent hook dir)
Post-install check: ctx doctor
```

| Config Format | Entry |
|---|---|
| JSON mcpServers | `"context-mode": { "command": "context-mode", "args": ["--mcp"] }` |
| TOML | `[mcp_servers.context-mode]\ncommand="context-mode"\nargs=["--mcp"]` |
| VS Code HTTP | `"context-mode": { "type": "http", "url": "http://localhost:3002/mcp" }` (via proxy) |
| Hook file | Agent-specific `hooks/context-mode.json` with pretooluse matcher |

Both hook AND MCP server should be configured when possible:
- Hook: intercepts tool calls at runtime (routing enforcement)  
- MCP server: exposes `ctx_execute`, `ctx_batch_execute`, `ctx_fetch_and_index` as callable tools

---

## Corrections from Review

### Antigravity IDE — NOT a VS Code Extension

Antigravity IDE is Google's **standalone coding agent IDE** — a separate product
with its own editor, workspace management, and Python SDK (`google-antigravity`).

**Key structural differences from previous analysis:**

| Point | Previous (Wrong) | Corrected |
|-------|-----------------|-----------|
| Product type | VS Code extension | Standalone IDE (Google) |
| IDE hooks.json format | Assumed `{}` output for ALL hooks | Only PostToolUse returns `{}`. PreToolUse returns `decision/reason/permissionOverrides` — same as CLI! |
| MCP format | Object mcpServers | Array mcpServers with `name` field + `serverUrl` key |
| PreInvocation | Basic injection | Supports `ephemeralMessage` (zero context cost) |
| Stop hook | Not implemented | Returns `{"decision": "continue|terminate", "reason": "..."}` |
| Google ADC | Not considered | Required for `authProviderType: google_credentials` MCP servers |

### CLI hooks.json format (from `antigravity.google/docs/hooks`)

```json
{ "hooks": { "PreToolUse": [...], "PostToolUse": [...], ... } }
```

### IDE hooks.json format (from `antigravity.google/docs/ide-hooks`)

```json
{
  "my-linter-hook": { "PostToolUse": [...] },
  "safety-gate": { "enabled": false, "PreToolUse": [...] },
  "reminder": { "PreInvocation": [...] }
}
```

---

## Universal Code Verification Policy

**All code in all plan files is PSEUDOCODE/REFERENCE ONLY.**
Every binary name, package name, file path, field name, and API call must be
independently verified against vendor documentation before implementation.

Key verification checklist per agent:
- `which <binary>` on clean target machine (not dev machine)
- `npm view <package>` for all npm packages  
- `npx --yes <package> --help` for all MCP server flags
- `<binary> --help` for all CLI flags used
- Vendor config schema docs for all JSON/YAML field names
- `architect-ai install --dry-run` before any real install
