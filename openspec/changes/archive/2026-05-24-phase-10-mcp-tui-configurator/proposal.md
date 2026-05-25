# Proposal: Phase 10 — MCP TUI Configurator: Detection, Generation & Management

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/10-phase-mcp-tui-configurator.md`
> **Change:** phase-10-mcp-tui-configurator
> **Phase:** sdd-propose
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-propose (architect-ai)

## Intent

Unify and secure the generation of Model Context Protocol (MCP) server configurations across all 5 supported platforms (VSCode, Antigravity, Gemini CLI, OpenCode, Claude Code) via a Go-based generator with dynamic Engram discovery, credential isolation, and atomic writes.

## Problem Context

1. **Hybrid Schema Collisions**: Antigravity and Gemini CLI crash when `command` and `httpUrl`/`serverUrl` are mixed for the same MCP server (e.g., `context7`).
2. **Credential Leaks**: Passwords (e.g., `ODOO_PASSWORD`) are written in plain text directly into configuration JSONs.
3. **Hardcoded Engram Paths**: Configurations point to static paths like `/opt/homebrew/bin/engram` or version-specific Cellar paths, which break across machines or updates.
4. **Missing MCP Nodes**: `context-mode` is entirely missing from some platforms, and `gemini-auth` plugin absent in OpenCode when `gemini` is installed.

## Proposed Approach

1. **MCP Generator Component** (`internal/components/mcp/generator.go`): Strategy pattern branching JSON generation by platform. VSCode gets `servers`, OpenCode gets `mcp`, others get `mcpServers`.
2. **Transport Purity**: `serverUrl` only for Antigravity, `httpUrl` only for Gemini, `type: stdio` explicit for VSCode. NEVER mix `command`/`args` with `*Url` keys.
3. **Secret Isolation** (`internal/components/mcp/secrets.go`): Passwords in `.env.mcp` (gitignored, 0600). Configs reference via `${ODOO_PASSWORD}` or `${input:odoo-password}`.
4. **Dynamic Engram Discovery** (`internal/components/mcp/engram_path.go`): `FindEngramBinary()` scans `$ENGRAM_BIN` → `$PATH` → common paths → Cellar (version-agnostic).
5. **Atomic Config Writing**: `.tmp` file + `os.Rename()` to prevent corruption.

## Transport Rules by Platform

| Server | Claude Code | OpenCode | VSCode | Antigravity | Gemini CLI |
|---|---|---|---|---|---|
| context7 | `npx` stdio | `remote` + `url` | `http` + `url` | `serverUrl` only | `httpUrl` only |
| engram | binary stdio | binary array `local` | binary `stdio` | binary stdio | binary stdio |
| sequential-thinking | `npx` stdio | `npx` local array | `npx` stdio | `npx` stdio | `npx` stdio |
| context-mode | `npx` stdio | `npx` local array | `npx` `stdio` explicit | `npx` stdio | `npx` stdio |
| odoo-mcp | `uvx` stdio | `uvx` local | `uvx` `stdio` | `uvx` stdio | `uvx` stdio |

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
