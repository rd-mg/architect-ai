# Proposal: Phase 8 — IDE/CLI Full Adapter Matrix v2

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/08-phase-ide-cli-full-matrix.md`
> **Priority:** 🟢 MEDIA-ALTA
> **Sources:** "Building Effective AI Coding Agents for the Terminal" (arxiv 2603.05344) · "Per Agent Adapter Harness" (gentle-ai) · "AgentGuard Runtime Verification"
> **Constraint:** Go = solo instalador. Enforcement = JSON configs + MD prompts + hooks.

## Intent

Ensure ALL changes from Phases 1-7 are reflected across all 5 IDE/CLI adapters. Each adapter has unique constraints (sub-agents, MCP, compression). This phase is the integration and parity guarantee.

## v2 Changes from Prior Draft

Adversarial audit detected:
1. **delegation_read en L2 (OpenCode)** — todos los L2 tenían delegation_read, destruyendo el aislamiento clean-room. Removido de todos los L2.
2. **VSCode degraded mode no documentado** — cuando MCP no disponible, el agente no tenía instrucciones claras.
3. **Antigravity sin sequential thinking inline** — el fallback inline faltaba en el adapter.
4. **Bash deny-list en OpenCode** — patrones de inyección de comandos no cubiertos.
5. **CLAUDE.md marker collision** — sin hash/checksum del contenido, sync sucesivos podían dejar marcadores huérfanos.

## Scope

### In Scope
- Full configuration for all 5 platforms: OpenCode, Claude Code, VSCode Copilot, Antigravity, Gemini CLI.
- Capability matrix documentation: real sub-agents, parallel execution, MCP, compress command.
- Platform-specific agent config files (opencode.json, CLAUDE.md, copilot-instructions.md, agent.md, GEMINI.md).
- MCP server configuration per platform (engram, context7, sequential-thinking, context-mode).
- Go implementation: `internal/install/adapter/injector.go` with SHA256 content hash, platform detection, and section injection.
- Validation: `ValidateInstallation()` for health checks against 6 required .atl/ files.

### Out of Scope
- New features beyond what Phases 1-7 define.
- Cross-platform test automation (future phase).

## Capability Matrix — Final State

| Feature | OpenCode | Claude Code | VSCode Copilot | Antigravity | Gemini CLI |
|---|---|---|---|---|---|
| L0 architect (super-orchestrator) | ✅ mode:primary | ✅ CLAUDE.md L0 section | ✅ Logical L0 | ✅ Simulated L0 | ✅ GEMINI.md L0 |
| L1 real sub-agents | ✅ JSON agents | ✅ Task tool | ❌ Logical only | ❌ Simulated | ✅ run_subagent |
| L2 parallel execution | ✅ | ✅ | ❌ | ❌ | ✅ |
| delegation_read on L1 ONLY | ✅ (fixed v2) | ✅ Task tool isolates | N/A | N/A | ✅ |
| delegation_read on L2 | ❌ REMOVED | ❌ Task tool clean | N/A | N/A | ❌ |
| MCP servers | ✅ opencode.json | ✅ .claude/settings.json | ⚠️ Extension API only | ❌ No MCP | ✅ .gemini/settings.json |
| Sequential thinking MCP | ✅ | ✅ | ❌ → inline fallback | ❌ → inline fallback | ✅ |
| Caveman mandatory | ✅ | ✅ | ✅ (in instructions) | ✅ (in agent.md) | ✅ |
| Native compress | ✅ /compact | ✅ /compact | ❌ → manual summary | ❌ → manual summary | ✅ /compress |
| GGA pre-commit | ✅ bash hook | ✅ bash hook | ✅ bash hook | ✅ bash hook | ✅ bash hook |
| Odoo L3 agents | ✅ | ✅ | ⚠️ Inline simulate | ⚠️ Inline simulate | ✅ |
| git branch isolation (sdd-apply) | ✅ | ✅ | ✅ | ✅ | ✅ |
| sdd-state.yaml Phase DAG | ✅ | ✅ | ✅ | ✅ | ✅ |
| Model routing per phase | ✅ per-agent model field | ✅ in Task delegation | ⚠️ Quality hint only | ⚠️ Quality hint only | ✅ |

## Impact
- Cross-platform parity: all 5 adapters have equivalent agent hierarchies.
- Platform limitations (no MCP, no sub-agents) are documented and have explicit fallbacks.
- Single `Supported` map in Go serves as source of truth for platform capabilities.
- SHA256 hash in markers prevents duplicate injection on successive syncs.

## Source
`/home/rdmachadog/Documents/fix_achitect-ai/08-phase-ide-cli-full-matrix.md`


## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
