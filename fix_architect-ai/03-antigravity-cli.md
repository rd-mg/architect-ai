# Fix & Improvement Plan — `antigravity-cli` Agent

**Agent ID:** `antigravity-cli` (new, split from `antigravity` in V4)  
**Runtime:** Antigravity CLI — `agy` TUI binary  
**Product string:** `antigravity-cli`  
**Config root:** `~/.gemini/antigravity-cli/`  
**Assets:** `internal/assets/antigravity-cli/` (NEW — to be created)  
**Go Adapter:** None (config written to `~/.gemini/antigravity-cli/` paths)  
**Priority:** 🟠 High — Currently completely absent from architect-ai; merged wrongly into IDE config

---

## 1. What Antigravity CLI Is

Antigravity CLI is the `agy` TUI binary — a terminal application that wraps the Gemini model with:
- A **rich TUI** (prompt panel, artifact picker, status line)
- A **plugin system** at `~/.gemini/antigravity-cli/plugins/<name>/`
- **Fine-grained permissions** engine (`allow/deny/ask` lists in `settings.json`)
- **Hooks** via `hooks.json` with `decision/reason/permissionOverrides` output (DIFFERENT from IDE)
- **Session management**: `/fork`, `/branch`, `--continue`, `agy <session-uuid>`
- **Skills** in `.agents/skills/` (project) or `~/.gemini/antigravity-cli/skills/` (global)
- **Sidecars** for background scheduled tasks
- **Status line** customization via `statusline.sh`
- **Context window**: 1,048,576 tokens (Gemini model, confirmed via status JSON `context_window_size`)
- Binary: `agy`; install via `curl | bash`

The CLI uses the **same hook `decision` output format** as the documented CLI hooks — NOT the IDE's `{}` format.

---

## 2. Current State in architect-ai

architect-ai has **zero CLI-specific assets**. The `internal/assets/antigravity/` directory was written for the IDE. The CLI has:
- No dedicated `sdd-orchestrator.md`
- No dedicated `thinking-agent.md`  
- No plugin structure awareness
- No skills installation to `~/.gemini/antigravity-cli/skills/`
- No `settings.json` permissions configuration
- No sidecar config

This is a **complete gap**.

---

## 3. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| AC-01 | Zero CLI-specific assets exist | 🔴 Critical |
| AC-02 | No plugin installer for `~/.gemini/antigravity-cli/plugins/` | 🔴 Critical |
| AC-03 | No `settings.json` permissions config written at install | 🟠 High |
| AC-04 | sequential-thinking absent from CLI plugin MCP config | 🟠 High |
| AC-05 | context-mode absent from CLI plugin MCP config | 🟠 High |
| AC-06 | codegraph absent from CLI plugin MCP config | 🟠 High |
| AC-07 | No skill installation to `~/.gemini/antigravity-cli/skills/` | 🟡 Medium |
| AC-08 | No sidecar config for archive cleanup (L5 fix) | 🟡 Medium |
| AC-09 | No status line script template | Low |
| AC-10 | No hooks.json for CLI (uses CLI decision format) | 🟠 High |
| AC-11 | `/fork` session management not documented for SDD parallel exploration | 🟡 Medium |

---

## 4. New Files to Create

### `internal/assets/antigravity-cli/sdd-orchestrator.md`

Adapts the generic SDD orchestrator with CLI-specific features:
- References `agy` binary affordances (TUI, `/artifact` picker, `/fork` for parallel exploration)
- Uses CLI hook decision format  
- References `~/.gemini/antigravity-cli/` paths for skill resolution
- Notes 1M token window (context saturation not a concern; still use `ctx_batch_execute` for transcript cleanliness)

### `internal/assets/antigravity-cli/thinking-agent.md`

- D1–D4 classifier (identical logic to other agents)
- Adds: use `/fork` for parallel hypothesis exploration in MODE 2 tasks
- Adds: use `agy --continue` to resume after context compaction

### `internal/assets/antigravity-cli/plugin.json`

architect-ai as an Antigravity CLI plugin:
```json
{
  "name": "architect-ai"
}
```

### `internal/assets/antigravity-cli/mcp_config.json`

```json
{
  "mcpServers": {
    "engram": {
      "command": "${ENGRAM_BIN}",
      "args": ["mcp", "--tools=agent"]
    },
    "context7": {
      "serverUrl": "https://mcp.context7.com/mcp"
    },
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"],
      "env": {}
    },
    "context-mode": {
      "command": "context-mode",
      "args": ["--mcp"],
      "env": {}
    },
    "codegraph": {
      "command": "npx",
      "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"],
      "env": {}
    },
    "notebooklm-mcp": {
      "command": "notebooklm-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### `internal/assets/antigravity-cli/hooks.json`

CLI format (uses `decision/reason/permissionOverrides`):
```json
{
  "safety-gate": {
    "PreToolUse": [{
      "matcher": "run_command",
      "hooks": [{
        "type": "command",
        "command": "architect-ai hook cli-pretooluse",
        "timeout": 5
      }]
    }]
  },
  "context-routing": {
    "PreInvocation": [{
      "hooks": [{
        "type": "command",
        "command": "architect-ai hook cli-preinvocation",
        "timeout": 2
      }]
    }]
  },
  "sdd-stop": {
    "Stop": [{
      "hooks": [{
        "type": "command",
        "command": "architect-ai hook cli-stop",
        "timeout": 3
      }]
    }]
  }
}
```

**`architect-ai hook cli-pretooluse`** reads stdin JSON, checks `toolCall.name` against `.atl/tool_policy.yaml`, outputs:
```json
{
  "decision": "allow|ask|deny",
  "reason": "string",
  "permissionOverrides": []
}
```

### `internal/assets/antigravity-cli/settings.json`

```json
{
  "permissions": {
    "allow": [
      "command(git)",
      "command(go test)",
      "command(npx)",
      "mcp(*)"
    ],
    "deny": [
      "command(rm -rf /)"
    ],
    "ask": [
      "command(npm install)",
      "write_file(*)"
    ]
  },
  "enableTerminalSandbox": false
}
```

### `internal/assets/antigravity-cli/sidecars/archive-cleaner.json`

```json
{
  "description": "Daily cleanup of openspec/changes/archive/ entries older than configured TTL.",
  "builtin": "schedule",
  "args": ["0 2 * * *", "architect-ai", "archive", "--prune", "--max-age", "30d"]
}
```

---

## 5. Fix Plan

### Fix AC-02 — Plugin Installer

**File:** `internal/agents/antigravity_cli/adapter.go` (new)

```go
package antigravitycli

const pluginDir = "~/.gemini/antigravity-cli/plugins/architect-ai"

func Install(dryRun bool) error {
    dir := expandHome(pluginDir)
    files := map[string]string{
        "plugin.json":    antigravityCLIPluginJSON,
        "mcp_config.json": antigravityCLIMCPJSON,
        "hooks.json":     antigravityCLIHooksJSON,
        "skills/":        "",  // directory
    }
    for path, content := range files {
        writeTo(filepath.Join(dir, path), content, dryRun)
    }
    // Write settings.json to ~/.gemini/antigravity-cli/settings.json
    writeSettings(dryRun)
    // Copy skills
    installSkills(dryRun)
    return nil
}
```

### Fix AC-03 — settings.json Permissions

Written to `~/.gemini/antigravity-cli/settings.json` at install time. Merges with existing settings (sparse persistence — only changed keys written).

### Fix AC-07 — Skill Installation

```bash
agy plugin install ~/.gemini/antigravity-cli/plugins/architect-ai
```

Or: copy SKILL.md files to `~/.gemini/antigravity-cli/skills/architect-sdd/SKILL.md`.

### Fix AC-08 — Sidecar for L5

```go
func InstallSidecar(dryRun bool) error {
    sidecarDir := expandHome("~/.gemini/config/sidecars/architect-archive/")
    writeTo(filepath.Join(sidecarDir, "sidecar.json"), sidecarJSON, dryRun)
    // Enable in settings
    enableSidecar("architect-archive", dryRun)
    return nil
}
```

### Fix AC-11 — Fork for Parallel Exploration

Add to `sdd-explore.md` (antigravity-cli version):

```markdown
### CLI Parallel Exploration (MODE 2 tasks)
When D2 ≥ 3 (high complexity, multiple approaches):
1. `/fork` → creates a new session for each hypothesis
2. Run sdd-explore independently in each fork
3. `/branch` back to original session
4. `mem_search("sdd/{change_name}/explore")` to collect all fork results
5. Synthesize in original session

Note: architect-ai Stop hook writes fork results to Engram automatically.
```

---

## 6. sequential-thinking Detection & Configuration (CLI)

```go
// internal/agents/antigravity_cli/adapter.go
func DetectMCPs() MCPAvailability {
    return MCPAvailability{
        SequentialThinking: checkNPX("@modelcontextprotocol/server-sequential-thinking"),
        ContextMode:        checkBinary("context-mode"),
        CodeGraph:          checkNPX("@colbymchenry/codegraph") || checkBinary("codegraph"),
        NotebookLM:         checkBinary("notebooklm-mcp"),
        Context7:           true, // remote — always available if internet
        Engram:             checkBinary("engram"),
    }
}
```

CLI-specific note: `sequential-thinking` runs via `npx -y` — requires npm/npx. If absent, the CLI agent will still invoke `sequential_thinking` in its prompt but the tool call will fail silently. Detection must warn at install time.

---

## 7. context-mode Detection & Configuration (CLI)

CLI uses context-mode as an **MCP server** in `mcp_config.json` (same as IDE).  
Additionally, a CLI hook routes shell tool calls through `context-mode`:

```json
{
  "context-mode-hook": {
    "PreToolUse": [{
      "matcher": "run_command",
      "hooks": [{
        "command": "context-mode hook agy pretooluse",
        "timeout": 5
      }]
    }]
  }
}
```

**Detection**: `context-mode --version` at install. If absent: install via `npm install -g context-mode`.

**CLI-specific advantage**: The `/mcp` TUI command opens the interactive MCP Manager Overlay, letting users verify context-mode is connected. Add this to the install success message.

---

## 8. Improvement Roadmap

| Week | Task | Files |
|------|------|-------|
| 1 | AC-01/02: Create antigravity-cli asset dir + plugin installer | New files |
| 1 | AC-10: hooks.json (CLI format) | `antigravity-cli/hooks.json` |
| 1 | AC-03: settings.json permissions | `antigravity-cli/settings.json` |
| 2 | AC-04/05/06: MCP config (seq-think, ctx-mode, codegraph) | `mcp_config.json` |
| 2 | AC-07: Skill installation to global skills dir | `adapter.go` |
| 3 | AC-08: Archive sidecar | `sidecars/archive-cleaner.json` |
| 3 | AC-11: Fork parallel exploration in sdd-explore | `sdd-explore.md` |
| 4 | AC-09: Status line script template | `statusline.sh.tpl` |
