# Specification: Phase 10 — MCP TUI Configurator

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/10-phase-mcp-tui-configurator.md`
> **Change:** phase-10-mcp-tui-configurator
> **Phase:** sdd-spec
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-spec (architect-ai)

## Requirements

### 1. Platform-Correct MCP Generation
- `GenerateConfig(platform, opts)` MUST produce correct JSON for: `vscode`, `antigravity`, `gemini`, `opencode`, `claude`.
- Unknown platforms MUST return error.

### 2. Transport Schema Purity
- Antigravity context7: `serverUrl` only. NO `command`, `args`, or `httpUrl` keys.
- Gemini CLI context7: `httpUrl` only. NO `command` or `args` keys.
- VSCode: explicit `type: stdio` for stdio servers. Root key `servers` (NOT `mcpServers`).
- OpenCode context7: `type: remote` + `url`. `context-mode`: `type: local` + command array.
- Claude Code context7: `npx` stdio. Root key: `mcp_servers`.

### 3. Credential Security
- `ODOO_PASSWORD` MUST NEVER appear as plaintext in any generated JSON.
- Antigravity/OpenCode/Claude/Gemini: use `${ODOO_PASSWORD}` referencing `.env.mcp`.
- VSCode: use `${input:odoo-password}` with `promptString` input definition.
- `WriteSecretsEnv()` MUST write to `.env.mcp` with 0600 permissions and auto-add to `.gitignore`.

### 4. Engram Binary Discovery
- `FindEngramBinary()` MUST search in order: `$ENGRAM_BIN` env var → `$PATH` → common paths → Homebrew/Linuxbrew Cellar (version-agnostic).
- MUST NOT hardcode any version number.
- Graceful fallback to `"engram"` (PATH lookup) if not found.

### 5. Gemini Auth Plugin Auto-Detection
- OpenCode config MUST include `opencode-gemini-auth@latest` in `plugin` array IF `gemini` binary is in `$PATH`.
- MUST NOT include plugin if `gemini` is not installed.

### 6. Atomic Config Writing
- `WriteConfig()` MUST write to `.tmp` file then `os.Rename()`.
- MUST create parent directories if needed.

### 7. VSCode Odoo Integration
- If `IsOdooProject=true`: add `odoo` server and `postgres` server (if PostgresURL provided).
- Add `inputs` section with `promptString` for odoo-password.

## Transport Rules Matrix

| Server | Claude | OpenCode | VSCode | Antigravity | Gemini |
|---|---|---|---|---|---|
| context7 | `npx` stdio | `remote` + `url` | `http` + `url` | `serverUrl` only | `httpUrl` only |
| engram | binary stdio | binary array `local` | binary `stdio` | binary stdio | binary stdio |
| sequential-thinking | `npx` stdio | `npx` local array | `npx` stdio | `npx` stdio | `npx` stdio |
| context-mode | `npx` stdio | `npx` local array | `npx` `stdio` explicit | `npx` stdio | `npx` stdio |
| odoo-mcp | `uvx` stdio | `uvx` local | `uvx` `stdio` | `uvx` stdio | `uvx` stdio |

## Scenarios

### Scenario 1: Antigravity context7 — Pure serverUrl
**Given** `GenerateConfig("antigravity", opts)` is called.
**Then** `context7` JSON has `serverUrl` key only.
**And** zero `command`, `args`, or `httpUrl` keys.

### Scenario 2: Gemini context7 — Pure httpUrl
**Given** `GenerateConfig("gemini", opts)` is called.
**Then** `context7` has `httpUrl`, `timeout`, `trust` keys.
**And** NO `command` or `args` keys.

### Scenario 3: VSCode Root Key
**Given** `GenerateConfig("vscode", opts)` is called.
**Then** root key is `"servers"` (NOT `"mcpServers"`).

### Scenario 4: Odoo Password Never Inline
**Given** `GenerateConfig("antigravity", {IsOdooProject: true})` is called.
**Then** `ODOO_PASSWORD` value is `"${ODOO_PASSWORD}"`, not a real password.

### Scenario 5: Gemini Auth Plugin Detection
**Given** `gemini` binary is in PATH.
**When** `GenerateConfig("opencode", {GeminiInstalled: true})` is called.
**Then** plugin array includes `"opencode-gemini-auth@latest"`.
**Given** `gemini` NOT in PATH.
**Then** plugin array does NOT include gemini-auth.

### Scenario 6: Engram Binary Discovery — Cellar
**Given** `ENGRAM_BIN` not set, `engram` not in PATH, binary at `/home/linuxbrew/.linuxbrew/Cellar/engram/1.15.9/bin/engram`.
**When** `FindEngramBinary()` is called.
**Then** returns the Cellar path without hardcoding version.

### Scenario 7: Atomic Config Write
**Given** generating config JSON.
**Then** writes to `.tmp` first, then atomic rename. No `.tmp` file remains after success.

### Scenario 8: OpenCode context-mode Present
**Given** `GenerateConfig("opencode", opts)` is called.
**Then** `context-mode` in `mcp` block with `type: local` and `npx -y @mksglu/context-mode` command array.

## Verification Criteria

| Test | Input | Expected | PASS if |
|---|---|---|---|
| Antigravity context7 | GenerateConfig("antigravity", {}) | context7 has exactly `serverUrl` | No command/args/httpUrl |
| Gemini context7 | GenerateConfig("gemini", {}) | context7 has `httpUrl` | No command key |
| VSCode root key | GenerateConfig("vscode", {}) | Root = "servers" | No "mcpServers" |
| Odoo password | GenerateConfig("antigravity", {odoo}) | `${ODOO_PASSWORD}` | No plaintext |
| Gemini plugin | GenerateConfig("opencode", {gemini}) | Plugin present | Matches gemini availability |
| Engram Cellar | Binary only in Cellar | Returns Cellar path | No hardcoded version |
| Atomic write | WriteConfig() | .tmp removed | No .tmp after success |

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
