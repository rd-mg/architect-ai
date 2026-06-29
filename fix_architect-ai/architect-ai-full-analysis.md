# architect-ai — Full Codebase Analysis & Error Report

**Date:** 2026-06-27  
**Batches Analysed:** 7 (all files)  
**Scope:** MCP config, hooks, guardrails, overlays, orchestrators, skills, personas, postures

> **Severity legend:** 🔴 Critical (breaks functionality) · 🟠 High (significant gap) · 🟡 Medium · 🟢 Low

---

## Table of Contents

1. [MCP Configuration Analysis — Per Agent](#1-mcp-configuration-analysis)
2. [MCP Proxy Conflict Analysis](#2-mcp-proxy-conflict-analysis)
3. [Missing & Active MCP Coverage Matrix](#3-mcp-coverage-matrix)
4. [Antigravity CLI — Install Detection Gap](#4-antigravity-cli-install-gap)
5. [CodeGraph in Research & Explore Agents](#5-codegraph-in-research-agents)
6. [Skill Registry (.atl) Usage in Orchestrators](#6-skill-registry-atl-usage)
7. [Odoo Overlay Alignment to L Architecture](#7-odoo-overlay-alignment)
8. [Judge Agents — Adversarial Verification Instantiation](#8-judge-agents)
9. [SDD Archive — README.md and __manifest__.py](#9-sdd-archive-gaps)
10. [SDD Spec/Design/Tasks Completeness Before Apply](#10-sdd-spec-design-tasks-completeness)
11. [General Orchestrator Sub-Agent Cognitive Postures](#11-general-orchestrator-postures)
12. [Routing Matrix vs All 11 Cognitive Postures](#12-routing-matrix-coverage)
13. [Personas — Guardrails & MCP Completeness](#13-personas-guardrails-mcp)
14. [Spanish Regionalism Audit](#14-spanish-regionalism)
15. [Additional Critical Bugs Found](#15-additional-critical-bugs)
16. [Consolidated Fix Priority Queue](#16-fix-priority-queue)

---

## 1. MCP Configuration Analysis

### 1.1 Two Parallel MCP Code Paths — Architecture Conflict

There are **two separate code paths** that write MCP configuration:

**Path A — `internal/components/mcp/generator.go` (`GenerateConfig`)**  
Used for **fresh installs**. Writes a complete config from scratch for each platform.

**Path B — `internal/components/mcp/inject.go` + `overlay.go`**  
Used for **incremental updates** (adding individual MCP servers to existing configs).

These two paths are NOT synchronized:

| Feature | Path A (generator) | Path B (inject/overlay) |
|---------|-------------------|------------------------|
| engram | ✅ | ❌ No `InjectEngram()` function |
| context7 | ✅ | ✅ `Inject()` calls `OverlayFor(ServerContext7)` |
| sequential-thinking | ✅ | ✅ `InjectSequentialThinking()` |
| context-mode | ✅ | ❌ No `InjectContextMode()` function |
| codegraph | ✅ | ❌ No `InjectCodeGraph()` function |
| notebooklm-mcp | ❌ (absent from generator!) | ✅ `InjectNotebookLM()` |
| odoo MCP | ✅ (VSCode + Antigravity only) | ❌ No `InjectOdoo()` function |

🔴 **Critical Bug**: `InjectCodeGraph()` function does not exist in `inject.go`. CodeGraph can only be applied via `generator.go` (fresh install). Any update/sync operation that uses the inject path will NOT add codegraph.

🔴 **Critical Bug**: `notebooklm-mcp` is absent from ALL generators (`generateVSCode`, `generateAntigravity`, `generateGemini`, `generateOpenCode`, `generateClaude`). NotebookLM can only be added via `InjectNotebookLM()`, which requires a separate install step. Fresh installs do NOT configure NotebookLM.

🟠 **High**: No `InjectContextMode()` function means context-mode cannot be selectively added to an existing config. Only fresh generates include it.

🟠 **High**: No `InjectEngram()` function means Engram cannot be selectively added either.

### 1.2 Per-Agent MCP Analysis

#### Claude Code (`generateClaude`)

```json
{
  "mcp_servers": {
    "engram":             { "command": "...", "type": "stdio" },
    "context7":           { "command": "npx", "args": ["-y", "@upstash/context7-mcp@latest"] },
    "sequential_thinking": { "command": "npx", ... },
    "context_mode":       { "command": "npx", ... },
    "codegraph":          { "command": "npx", ... }
  }
}
```

🔴 **Critical Bug**: Key names use **underscores** (`sequential_thinking`, `context_mode`) not hyphens. Every other agent uses `sequential-thinking`, `context-mode`. This is inconsistent with:
- `overlay.go` which returns `"sequential-thinking"` (hyphenated)
- The tool references in phase protocols which use the hyphenated names
- The MCP server name that Claude Code actually registers and exposes to sub-agents

**Fix**: Change `"sequential_thinking"` → `"sequential-thinking"` and `"context_mode"` → `"context-mode"` in `generateClaude()`.

🟠 **High**: Missing `notebooklm-mcp` in `generateClaude`.

🟡 **Medium**: Context7 package uses `@upstash/context7-mcp@latest` (with version tag). All other agents use unversioned `@upstash/context7-mcp` or the remote URL. Inconsistency may cause version drift.

Also, Claude's `overlay.go` context7 entry returns:
```json
{ "command": "npx", "args": ["-y", "@upstash/context7-mcp"] }
```
This is **stdio** format (separate MCP file for Claude). But `generator.go` Claude has the same command in `mcp_servers` object — both paths configure context7 differently causing potential duplication.

#### VS Code (`generateVSCode`)

```json
{
  "servers": {
    "context-mode": { "type": "stdio", "command": "npx", "args": ["-y", "@mksglu/context-mode"] },
    "context7":     { "type": "http",  "url": "https://mcp.context7.com/mcp" },
    "engram":       { "type": "stdio", "command": "...", "args": ["mcp", "--tools=agent"] },
    "sequential-thinking": { "type": "stdio", "command": "npx", ... },
    "codegraph":    { "type": "stdio", "command": "npx", ... },
    "odoo":         { ... }  // if Odoo project
  }
}
```

✅ All 5 core MCPs present. Odoo MCP conditional on `IsOdooProject`.  
🟠 **High**: Missing `notebooklm-mcp`.  
🟡 **Medium**: VS Code uses `"servers"` key (correct for VS Code format), but `overlay.go` sequentialThinkingOverlay for VS Code/Cursor also uses `"servers"` key — merge would work but could cause duplication if both paths apply.

#### Antigravity IDE (`generateAntigravity`)

```json
{
  "mcpServers": {
    "context7":           { "serverUrl": "https://mcp.context7.com/mcp" },
    "context-mode":       { "command": "npx", "args": ["-y", "@mksglu/context-mode"] },
    "engram":             { "command": "...", "args": ["mcp", "--tools=agent"] },
    "sequential-thinking": { "command": "npx", ..., "timeout": 30000, "trust": true },
    "codegraph":          { "command": "npx", ..., "timeout": 30000, "trust": true },
    "odoo":               { ... }  // if Odoo project
  }
}
```

✅ All 5 core MCPs present. Odoo MCP conditional.  
🟠 **High**: Missing `notebooklm-mcp`.  
🟡 **Medium**: Uses `@mksglu/context-mode` (npm package) — should also support fallback to global `context-mode` binary.

#### Antigravity CLI — Asset MCP Config (`internal/assets/antigravity-cli/mcp_config.json`)

```json
{
  "mcpServers": {
    "engram":              { "command": "${ENGRAM_BIN}", "args": ["mcp", "--tools=agent"] },
    "context7":            { "serverUrl": "https://mcp.context7.com/mcp" },
    "sequential-thinking": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"] },
    "context-mode":        { "command": "context-mode", "args": ["--mcp"] },
    "codegraph":           { "command": "npx", "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"] }
  }
}
```

✅ All 5 core MCPs defined in the asset file.  
🔴 **Critical Bug**: This file CANNOT be installed — no Go adapter or model.AgentID for `antigravity-cli`. See §4.  
🟠 **High**: Missing `notebooklm-mcp`.  
🟡 **Medium**: `${ENGRAM_BIN}` placeholder — not resolved at write time since there's no Go code to do so.

#### Gemini CLI (`generateGemini`)

```json
{
  "mcpServers": {
    "context7":           { "httpUrl": "https://mcp.context7.com/mcp", "timeout": 30000, "trust": false },
    "context-mode":       { "command": "npx", "args": ["-y", "@mksglu/context-mode"], "timeout": 15000 },
    "engram":             { "command": "...", "args": ["mcp", "--tools=agent"] },
    "sequential-thinking": { "command": "npx", ..., "timeout": 30000, "trust": true },
    "codegraph":          { "command": "npx", ..., "timeout": 30000, "trust": true }
  }
}
```

✅ All 5 core MCPs present.  
🟠 **High**: Missing `notebooklm-mcp`.  
🟡 **Medium**: Gemini adapter uses `MCPStrategy: StrategyMergeIntoSettings` — writes to `~/.gemini/settings.json`. But `generateGemini` returns a FULL config object (not just mcpServers). The merge behavior of `filemerge.MergeJSONObjects` must handle top-level keys like `general`, `ide`, `model`, `security`, `ui` without clobbering user settings.

#### OpenCode (`generateOpenCode`)

```json
{
  "mcp": {
    "context7":           { "enabled": true, "type": "remote", "url": "https://mcp.context7.com/mcp" },
    "context-mode":       { "type": "local", "command": ["npx", "-y", "@mksglu/context-mode"], "enabled": true },
    "engram":             { "type": "local", "command": ["...", "mcp", "--tools=agent"] },
    "sequential-thinking": { "type": "local", "command": [...], "enabled": true },
    "codegraph":          { "type": "local", "command": [...], "enabled": true }
  }
}
```

✅ All 5 core MCPs present.  
🟠 **High**: Missing `notebooklm-mcp`.  
🟡 **Medium**: `notebooklm-mcp` in `overlay.go` for OpenCode is set to `"enabled": false` — it's registered but disabled. This is deliberate (user must configure their notebook) but not documented clearly.

#### Cursor, Kiro, Windsurf, QwenCode — via `overlay.go` only

These agents use `overlay.go` for individual server injection:
- `context7Overlay` ✅ (uses `"serverUrl"` format for Windsurf, QwenCode, KiroIDE, Antigravity)
- `sequentialThinkingOverlay` ✅ (all agents covered)
- `codegraphOverlay` ✅ (all agents covered)
- `notebookLMOverlay` ✅ (all agents covered)

🟠 **High**: These agents only receive MCPs via INCREMENTAL overlay injection. Fresh generate path (`generator.go`) does NOT exist for these agents. All MCPs depend on `Inject()`, `InjectSequentialThinking()`, `InjectNotebookLM()` being called separately and correctly sequenced in the install pipeline.

🟡 **Medium**: No validation that ALL overlay injections succeed before marking install complete.

### 1.3 Hooks Configuration

#### Antigravity IDE hooks — Format is correct

The IDE uses named-group format (corrected in fix plans). No Claude/VS Code hook config found in `internal/components/hooks/hooks.go`.

🟡 **Medium**: `internal/components/hooks/hooks.go` implements session metering hook only. The context-mode pre-tool-use hook (`.claude/hooks/`, `.cursor/hooks/`) is NOT managed by the hooks component — it would need to be written by `inject.go` or a new hooks installer.

🔴 **Critical Bug**: No hook file writer exists for Claude, Cursor, or Gemini. The context-mode routing policy is injected as text into prompts, but the **actual hook file** that intercepts tool calls (e.g., `.claude/hooks/context-mode.json`) is NOT created by any Go code. This means context-mode routing is advisory (text in prompt) but NOT enforced at the tool call level.

---

## 2. MCP Proxy Conflict Analysis

**No actual proxy code exists.** The VS Code `generator.go` configures stdio MCPs with `"type": "stdio"` which is the correct VS Code mcp.json format. VS Code DOES support stdio MCPs natively — no HTTP proxy needed.

🟡 **Potential confusion**: The fix plan `07-vscode.md` documents an HTTP proxy for VS Code stdio MCPs but this was written before confirming VS Code supports stdio in `mcp.json`. The actual implementation in `generator.go` correctly uses `"type": "stdio"` for local servers and `"type": "http"` for remote servers. **No conflict or proxy needed.**

**However**: The `overlay.go` `sequentialThinkingOverlay` for VS Code/Cursor uses `"servers"` key with `"type": "stdio"`. The `generator.go` also uses `"servers"` key. If both are applied to `.vscode/mcp.json`, `filemerge.MergeJSONObjects` will deep-merge them correctly, but the `servers.sequential-thinking` entry would be written twice (once by generate, once by overlay injection). This creates duplicate registration but no functional conflict.

---

## 3. MCP Coverage Matrix

| MCP Server | Claude | VSCode | Antigravity-IDE | Antigravity-CLI | Gemini | OpenCode | Cursor | Kiro | Windsurf | Qwen |
|-----------|--------|--------|-----------------|-----------------|--------|----------|--------|------|----------|------|
| engram | ✅ | ✅ | ✅ | ✅ asset | ✅ | ✅ | via overlay | via overlay | via overlay | via overlay |
| context7 | ✅ | ✅ | ✅ | ✅ asset | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| sequential-thinking | ⚠️ wrong key | ✅ | ✅ | ✅ asset | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| context-mode | ⚠️ wrong key | ✅ | ✅ | ✅ asset | ✅ | ✅ | ❌ no overlay | ❌ | ❌ | ❌ |
| codegraph | ✅ | ✅ | ✅ | ✅ asset | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| notebooklm-mcp | ❌ | ❌ | ❌ | ❌ | ❌ | ⚠️ disabled | ✅ | ✅ | ✅ | ✅ |
| odoo MCP | ❌ | ✅ Odoo only | ✅ Odoo only | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

**Key findings:**
- `notebooklm-mcp` missing from ALL generators (Claude, VSCode, Antigravity, Gemini, OpenCode)
- `context-mode` missing from Cursor, Kiro, Windsurf, Qwen overlay paths
- `odoo MCP` only available for VSCode and Antigravity IDE — missing from Gemini, OpenCode, Cursor
- Claude has wrong key names for sequential-thinking and context-mode
- Antigravity CLI has correct asset config but NO installer

### Expected MCP Active Status per User Requirements

Per requirements, ALL agents should have active:
- engram ✅ (mostly present)
- context-mode ❌ (partial)
- codegraph ✅ (mostly present)
- context7 ✅ (present everywhere)
- sequential-thinking ✅ (mostly present, wrong key for Claude)
- notebooklm-mcp-cli ❌ **missing from generators — must be added**
- mcp-server-odoo ❌ **only for VSCode + Antigravity per requirement — correct but Antigravity-CLI missing**

---

## 4. Antigravity CLI — Install Detection Gap

### 4.1 Current State

**Assets exist** in `internal/assets/antigravity-cli/`:
- `hooks.json` ✅ (CLI named-group format)
- `mcp_config.json` ✅ (all 5 core MCPs)
- `plugin.json` ✅
- `settings.json` ✅ (permissions)
- `sidecars/archive-cleaner.json` ✅

**What is MISSING:**

🔴 **Critical**: No Go adapter at `internal/agents/antigravity-cli/adapter.go`
🔴 **Critical**: No `model.AgentAntigravityCLI` constant in `internal/model/`
🔴 **Critical**: Not registered in `overlay.go` agent switches
🔴 **Critical**: Not in `generator.go`
🔴 **Critical**: Not in catalog/registry
🔴 **Critical**: Not in `app.go` install pipeline
🔴 **Critical**: `install` command does NOT detect or install for Antigravity CLI (`agy`)
🔴 **Critical**: `antigravity-cli` asset dir is NOT in the `go:embed` directive

### 4.2 Detection Logic Needed

The `agy` binary is the CLI tool. Detection should:
```go
// internal/agents/antigravity-cli/adapter.go (NEEDS CREATION)
func (a *Adapter) Detect(ctx context.Context, homeDir string) (bool, string, string, bool, error) {
    // 1. Check for agy binary on PATH
    binaryPath, err := exec.LookPath("agy")
    installed := err == nil
    
    // 2. Check plugin dir  
    pluginDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins")
    stat := os.Stat(pluginDir)
    
    return installed, binaryPath, pluginDir, installed, nil
}
```

### 4.3 Plugin Installer Needed

The CLI uses a plugin system. Installation writes to:
`~/.gemini/antigravity-cli/plugins/architect-ai/`

The installer must:
1. Copy `mcp_config.json`, `hooks.json`, `plugin.json`, `settings.json` to plugin dir
2. Run `agy plugin install ~/.gemini/antigravity-cli/plugins/architect-ai` (or equivalent)
3. Copy skills to `~/.gemini/antigravity-cli/skills/` (global skills)
4. Install sidecar if schedule support detected

### 4.4 embed Directive Fix

**File:** `internal/assets/assets.go` (or wherever the `go:embed` is)

```go
// Current (line 19057 in batch 5):
//go:embed all:claude all:opencode all:generic all:skills all:gga all:gemini all:codex all:antigravity all:windsurf all:cursor all:qwen all:kiro all:overlays all:_shared all:vscode all:workflows source-map.json

// Needed:
//go:embed all:claude all:opencode all:generic all:skills all:gga all:gemini all:codex all:antigravity all:antigravity-cli all:windsurf all:cursor all:qwen all:kiro all:overlays all:_shared all:vscode all:workflows source-map.json
```

---

## 5. CodeGraph in Research Agents

### 5.1 sdd-explore.md — CodeGraph Missing

Looking at Claude `sdd-phase-protocols/sdd-explore.md`:

**Step 0**: Calls `sequential_thinking` ✅  
**Section B (Code Investigation)**: 5-Step Skim Protocol using **ripgrep only**

**MISSING**: No CodeGraph integration. The tools `codegraph_context`, `codegraph_trace`, `codegraph_callers`, `codegraph_impact` are NOT mentioned anywhere in any sdd-explore.md (Claude, generic, Gemini, etc.).

🔴 **Critical**: CodeGraph MCP is installed but never used. The eintegrate check E-07 expects `LspFindReferences` in skill-registry (satisfied by codegraph) but the actual phase protocols don't call codegraph tools.

### 5.2 Research Routing Policy — CodeGraph Not Included

The research routing policy (5 steps) in ALL orchestrators:
```
STEP 1 — Engram
STEP 2 — Local ripgrep
STEP 3 — Context7
STEP 4 — NotebookLM
STEP 5 — Web search
```

🔴 **Critical**: CodeGraph not in the research routing policy. CodeGraph should be between Step 2 (ripgrep) and Step 3 (Context7):

```
STEP 1 — Engram (session memory)
STEP 2a — Local ripgrep (text search)
STEP 2b — CodeGraph (semantic: call chains, impact, callers) ← MISSING
STEP 3 — Context7 (external docs)
STEP 4 — NotebookLM
STEP 5 — Web search
```

### 5.3 Fix for sdd-explore.md

Add after Step 0 (sequential_thinking) and before Section B:

```markdown
## Step 0b: Semantic Code Graph (if codegraph_context available)

Priority: run BEFORE ripgrep when codegraph is available.

1. `codegraph_context(query: "{change_topic}", maxNodes: 25, format: "markdown")`
   → Semantic context pack: related functions, types, files, call chains
2. `codegraph_trace(entry: "{identified_entrypoint}")`
   → Full call chain from trigger to leaf
3. `codegraph_impact(nodeId: "{primary_node}", depth: 3)`
   → Blast radius: what breaks if this changes
4. `codegraph_callers(nodeId: "{primary_node}")`
   → All inbound call sites (LspFindReferences equivalent)

Fallback: if codegraph unavailable, continue with Section B ripgrep.
```

### 5.4 Fix for Research Routing Policy

Update all `context-mode-routing-policy.md` and `research-routing-policy.md` instances:

```markdown
## RESEARCH-ROUTING POLICY

STEP 1 — Engram: mem_search with specific topic_key → USE if found
STEP 2a — Local ripgrep: rg in project files → USE if pattern found
STEP 2b — CodeGraph (if available): codegraph_context/callers/trace → USE for semantic relationships
STEP 3 — Context7: framework/library docs → USE for external APIs
STEP 4 — NotebookLM: synthesis, migration guides (Mode 1/2 only)
STEP 5 — Web search: last resort (not available in Mode 3)
```

---

## 6. Skill Registry (.atl) Usage in Orchestrators

### 6.1 Skill Registry Read — Implementation Analysis

Both General Orchestrator and SDD Orchestrator have:

```markdown
## Skill Resolution
1. mem_search("skill-registry", project) → mem_get_observation(id)
2. Fallback: read .atl/skill-registry.md
3. Cache Compact Rules section
```

✅ The mechanism is correct and present in Claude orchestrators.

🟡 **Issue**: `{{ include "_shared/caveman-identity-block.md" }}` and similar includes are Go template syntax, but architect-ai embeds files and writes them directly. These `{{ include }}` directives are NOT rendered by any Go template engine — they are written as literal text. Sub-agents receiving the prompt would see `{{ include ... }}` literally.

**Fix**: Either:
1. Replace `{{ include "..." }}` with actual file content at write time (Go `text/template`)
2. Or document that these are handled by the agent runtime (not Go code)

🔴 **Critical**: No Go code processes `{{ include "..." }}` directives before writing agent config files. If these appear in the written files, they are non-functional placeholders.

### 6.2 .atl/skill-registry.md Creation

The skill registry is built by `internal/app/skills_cmd.go`. It writes to `.atl/skill-registry.md`. The keyword index described in the blueprint is **NOT implemented** — the registry is a flat markdown file without a machine-readable keyword index.

🟠 **High**: Without a keyword index, the Dynamic Context Assembler described in the V4 blueprint cannot function. The orchestrators' "Skill Resolver" mechanism (described in `_shared/skill-resolver.md`) cannot filter skills by keyword match.

---

## 7. Odoo Overlay Alignment to L Architecture

### 7.1 L-Layer Alignment

The Odoo overlay L3 agents in `odoo-overlay-routing.md`:
```
L0 = thinking-agent (thin proxy router)
L1a = general-orchestrator
L1b = sdd-orchestrator
L2 = sdd-phase sub-agents (sdd-explore, sdd-apply, etc.)
L3 = Odoo-specific sub-agents (odoo-expert, odoo-code-reviewer, etc.)
```

✅ Cross-calling patterns in overlay-routing.md correctly show L2 calling L3.

### 7.2 Overlay Detection Gap

🔴 **Critical**: The SDD Orchestrator checks for Odoo overlay:
```
Look for .atl/overlays/odoo-*/manifest.json
```

But NO Go code creates `.atl/overlays/odoo-*/manifest.json`. The overlay files are in `internal/assets/overlays/odoo-development-skill/` but the installer must write them to `.atl/overlays/`. 

**Verify**: Does the overlay installer write a `manifest.json`? Looking at batch 6 components... No `manifest.json` creation found in any component code. If no `manifest.json` exists, the Odoo overlay is NEVER activated by the sdd-orchestrator.

### 7.3 Odoo MCP — Server Availability

The `mcp-server-odoo` is configured only for VSCode and Antigravity IDE. Per requirements, odoo MCP should be VSCode and Antigravity only.

🟡 **Medium**: Antigravity IDE gets odoo MCP ✅. Antigravity CLI does NOT (but CLI doesn't have installer). VSCode gets odoo MCP ✅.

🟡 **Medium**: The `mcp-server-odoo` uses `uvx mcp-server-odoo` command. `uvx` requires `uv` (Python package manager). No detection or warning if `uv` is not installed.

### 7.4 archive-odoo.md — README + manifest.py

Checking `sdd-supplements/archive-odoo.md` (batch 2) for README and __manifest__.py requirements:
- Must verify this file explicitly requires updating `__manifest__.py` and `README.md`
- The base `sdd-archive.md` does NOT mention these files (see §9 below)

---

## 8. Judge Agents — Adversarial Verification Instantiation

### 8.1 Current Implementation

`sdd-verify.md` implements adversarial verification as:
```markdown
7. Apply adversarial-review (from adaptive-reasoning Mode 2):
   - Pass A: happy-path correctness
   - Pass B: failure-mode lens
```

This is a **SINGLE sub-agent** doing two internal passes. It is NOT two separate judge instances.

### 8.2 What Was Required

The `judgment-day` skill (now archived) was:
> "Parallel adversarial review — two independent judges review the same target"

Two SEPARATE Task tool calls to two SEPARATE judge sub-agents, each with independent context windows, independently reviewing the same artifact. The results are then synthesized.

### 8.3 The Gap

🔴 **Critical** (per user requirement): "Los agentes jueces y la verificacion adversaria son instanciados se forma separada" — Judge agents must be instantiated separately.

Current `sdd-verify.md` does NOT spawn two separate judge agents. It does a two-pass review within a single sub-agent context.

### 8.4 Fix Required

The SDD Orchestrator should, at sdd-verify delegation, spawn:
```
[PARALLEL — same response]
Task(agent="judge-primary",   model="sonnet", task="Review {change} from correctness angle. ++Adversarial")
Task(agent="judge-secondary", model="sonnet", task="Review {change} from failure-mode angle. ++Forensic")
```

Then synthesize both judge verdicts into a final APPROVED/NEEDS CHANGES verdict.

**New protocol section in sdd-verify.md:**

```markdown
## Judge Instantiation (MANDATORY — spawn BOTH in same orchestrator response)

Judge Primary:   +++Adversarial — find defects in happy path
Judge Secondary: +++Forensic   — trace failure modes and edge cases

Synthesis: APPROVED only if BOTH judges independently reach ≥ CONDITIONALLY APPROVED
Conflict: If judges disagree → escalate to user
```

---

## 9. SDD Archive — README.md and __manifest__.py Gaps

### 9.1 Generic sdd-archive.md — Missing File Updates

The generic `sdd-archive.md` procedure:
```
1. Read verify-report → check APPROVED
2. Generate archive summary
3. Persistence (Learned Patterns)
4. Move to archive/ folder (OpenSpec mode)
5. Update DAG state to "archived"
```

🔴 **Critical** (per user requirement): NO mention of updating:
- `README.md` — project/module documentation
- `__manifest__.py` — Odoo module manifest (version, depends, etc.)
- `CHANGELOG.md` — change history

### 9.2 Odoo supplement archive-odoo.md

The Odoo supplement should explicitly require these updates. Need to verify:
- `__manifest__.py`: `version` field must be bumped
- `README.md`: Installation and changelog sections must be updated

### 9.3 Fix for sdd-archive.md

Add to the `## Procedure` section in ALL agent variants of `sdd-archive.md`:

```markdown
## Step 3d: Project File Updates (MANDATORY — run BEFORE moving to archive/)

### Always:
- [ ] Check if `CHANGELOG.md` exists → add entry for this change:
      ```
      ## [{version}] — {date}
      ### Added / Changed / Fixed / Removed
      - {description from archive summary}
      ```
- [ ] Check if `README.md` exists → update:
      - Installation steps (if new dependency added)
      - Usage section (if API or behavior changed)
      - Known issues (if any remain open)

### If Odoo project (IS_ODOO = true):
- [ ] Update `__manifest__.py`:
      - Bump `version` (use SemVer within Odoo version prefix: `17.0.X.Y.Z`)
      - Update `depends` if new module dependencies were introduced
      - Update `data`, `assets` keys if new XML/CSS files added
      - Update `summary` or `description` if module purpose changed
- [ ] Verify `__manifest__.py` is valid Python (syntax check)

### If library/package:
- [ ] Update `package.json` / `pyproject.toml` / `go.mod` version field
```

---

## 10. SDD Spec/Design/Tasks Completeness Before Apply

### 10.1 Current Apply Gate

`sdd-apply.md` (all agents):
```
**HALT execution.** Forbidden from writing/modifying implementation code until:
- TDD specifications fully defined, documented, and approved in change's spec.
- tasks.md explicitly authorizes implementation phase.
- If TDD specs missing or incomplete, delegate back to sdd-spec or sdd-design.
```

This is the ONLY gate. It does NOT explicitly check:

🔴 **Critical** (per user requirement): No verification that spec/design/tasks have NO stubs, gaps, or TODO items.

### 10.2 Missing Completeness Checks

Required pre-apply validation that is ABSENT:

```markdown
## Pre-Apply Completeness Gate (MANDATORY — halt if ANY fails)

### Spec Completeness
- [ ] Each capability has ALL mandatory sections (Purpose, Preconditions, Behavior, 
      Postconditions, Error Handling, Invariants, Test Hooks)
- [ ] NO section contains "TODO", "TBD", "PLACEHOLDER", "N/A without justification"
- [ ] FMEA table present for all external I/O (severity × detection matrix)
- [ ] Sad-path BDD scenarios exist for all FMEA severity ≥ 3
- [ ] All success criteria are MEASURABLE (not "the system should work")

### Design Completeness
- [ ] Architecture diagram present (ASCII or Mermaid)
- [ ] ALL module boundaries documented
- [ ] Data flow & interface contracts defined
- [ ] ALL mandatory sections present (no "To be designed" stubs)
- [ ] ADR table has at least 1 entry
- [ ] YAGNI Gate table completed
- [ ] No open questions (if open → block apply, resolve first)

### Tasks Completeness
- [ ] Every spec capability maps to ≥ 1 task
- [ ] Every task has explicit acceptance criterion (not "it works")
- [ ] Execution Graph (Mermaid) present if ≥ 5 tasks
- [ ] No task describes "implement the feature" as a whole (must be atomic)
- [ ] Risk classification present for ALL tasks
- [ ] Risk-reason mandatory for all HIGH risk tasks

IF ANY check fails → STOP. Return status: blocked.
Reason: "Pre-apply completeness gate failed: {list of missing items}"
Route to appropriate phase: sdd-spec, sdd-design, or sdd-tasks.
```

### 10.3 Cross-Phase Reference Validation

Also missing: verification that tasks reference ACTUAL spec capabilities by name.

```markdown
## Cross-Phase Reference Check (MANDATORY before apply)

For each task in tasks.md:
  mem_search("sdd/{change-name}/spec") → mem_get_observation(id)
  Verify: task's acceptance criterion can be traced to a spec capability
  IF no matching spec capability → BLOCK: "Task N.N references unknown capability"
```

---

## 11. General Orchestrator Sub-Agent Cognitive Postures

### 11.1 Inconsistency Between Claude and Generic Routing Tables

**Claude `general-orchestrator.md`:**
```
"fix this", "debug", "research", "investigate" → /analyze → Analyst → +++Forensic, +++Systemic, +++Critical
"brainstorm", "ideate" → /brainstorm → Ideator → +++Divergent, +++Lateral, +++Diamond
"prototype" → /prototype → Generalist → +++Pragmatic
Other → Generalist → Auto-detected (D1-D4)
```

**Generic `general-orchestrator.md`:**
```
"fix this", "solve" → /solve → Solver → +++Forensic, +++Systemic
"debug", "trace" → /debug → Solver → +++Forensic, +++Adversarial
"research", "investigate" → /investigate → Researcher → +++Socratic, +++Empirical
"brainstorm", "ideate" → /brainstorm → Ideator → +++Divergent, +++Lateral, +++Diamond
"prototype" → /prototype → Generalist → +++Pragmatic
```

🟠 **High**: Two different agent taxonomies (Analyst vs Solver/Researcher). All agents should use the same routing table. The **generic version is more precise** — split debug/research/solve correctly.

### 11.2 Max Postures Violation

🔴 **Critical**: The `docs/cognitive-modes.md` states: **"Max Postures per prompt: 2"**

But routing tables assign:
- Analyst/Solver: `+++Forensic, +++Systemic, +++Critical` → **3 postures** ❌
- Ideator: `+++Divergent, +++Lateral, +++Diamond` → **3 postures** ❌

**Fix**: Reduce to max 2 postures per agent:

| Workflow | Agent | Fixed Postures |
|----------|-------|---------------|
| /solve | Solver | +++Forensic + +++Systemic |
| /debug | Solver | +++Forensic + +++Adversarial |
| /investigate | Researcher | +++Socratic + +++Empirical |
| /brainstorm | Ideator | +++Divergent + +++Lateral |
| /prototype | Generalist | +++Pragmatic |
| /analyze (complex) | Analyst | +++Critical + +++Systemic |

Note: `+++Diamond` is removed (it's Divergent+Lateral combined — splitting them achieves the same without requiring Diamond as a third option).

### 11.3 "Auto-detected (D1-D4)" for General Tasks — No Posture Mapping

🟡 **Medium**: Claude general-orchestrator says "Other general tasks → Generalist → Auto-detected (D1-D4)" but provides no explicit mapping of D1-D4 values to postures for Generalist tasks.

The Adaptive Reasoning Gate maps D1-D4 to modes but NOT to specific postures for the Generalist. The orchestrator must compute mode → posture before injecting into the Generalist prompt.

**Required mapping:**
```
Mode 1 (simple): Generalist → +++Pragmatic
Mode 2 (tactical): Generalist → +++Critical
Mode 2-ERR: Generalist → +++Forensic
Mode 3 (diagnostic): Generalist → +++Adversarial
Mode 3-CTX: Generalist → +++Pragmatic (compressed output)
```

---

## 12. Routing Matrix Coverage of All 11 Cognitive Postures

### 12.1 Current Routing Matrix

```
| Condition              | Mode          | Posture                      |
|------------------------|---------------|------------------------------|
| D1+D2 ≤ 2, D3+D4 ≤ 2  | Mode 1        | +++Pragmatic                 |
| D1+D2 ≥ 3 OR D3 ≥ 1   | Mode 2        | +++Critical                  |
| D3 ≥ 2 OR D4 ≥ 3      | Mode 3        | +++Adversarial + +++Systemic |
| D4 ≥ 3 (Saturated)    | Mode 3-CTX    | +++Pragmatic                 |
| D3 = 1 (Initial Error) | Mode 2-ERR    | +++Forensic                  |
```

**Postures covered:** +++Pragmatic, +++Critical, +++Adversarial, +++Systemic, +++Forensic = **5 of 11**

**NOT covered by routing matrix:**
- +++Socratic (6)
- +++Economic (7)
- +++Empirical (8)
- +++Divergent (9)
- +++Lateral (10)
- +++Diamond (11)

### 12.2 Architecture of the Gap

These 6 postures are **task-triggered** (via workflow intent) not **complexity-triggered** (via D1-D4). The routing matrix governs adaptation to complexity/error conditions; the task router governs workflow selection. This is architecturally sound.

🟠 **High issue**: The routing matrix and task router are documented in DIFFERENT sections of the orchestrator and sub-agent templates. There is no unified posture-assignment specification that clarifies when each source applies. Sub-agents may be confused about which posture source takes precedence.

### 12.3 Fix — Unified Posture Assignment Specification

Add to `_shared/adaptive-reasoning-gate-v2.md`:

```markdown
## Posture Assignment — Two Sources (explicit precedence)

### Source 1: Task Router (workflow-triggered postures)
Applied FIRST. Determined by user intent (brainstorm, debug, research, etc.):

| Workflow | Postures |
|----------|---------|
| /investigate | +++Socratic + +++Empirical |
| /brainstorm | +++Divergent + +++Lateral |
| /debug | +++Forensic + +++Adversarial |
| /solve | +++Forensic + +++Systemic |
| /prototype | +++Pragmatic |
| SDD: sdd-explore | +++Socratic |
| SDD: sdd-propose | +++Critical |
| SDD: sdd-spec | +++Systemic |
| SDD: sdd-design | +++Critical + +++Systemic |
| SDD: sdd-tasks | +++Pragmatic + +++Economic |
| SDD: sdd-apply | +++Pragmatic |
| SDD: sdd-verify | +++Adversarial |

### Source 2: Routing Matrix (complexity/error-triggered override)
Applied SECOND. Overrides Source 1 posture ONLY for:
- Mode 3: ALWAYS override → +++Adversarial + +++Systemic (error condition)
- Mode 2-ERR: ALWAYS override → +++Forensic (bug mode)
- Mode 3-CTX: ALWAYS override → +++Pragmatic (saturated context)

For Mode 1 and Mode 2: Source 1 posture is retained.

### Max Postures: 2 (INVARIANT)
If Source 1 assigns 2 postures and Source 2 triggers → Source 2 REPLACES (not adds).
```

---

## 13. Personas — Guardrails & MCP Completeness

### 13.1 Current Persona-Architect Content

`internal/assets/claude/persona-architect.md`:
- ✅ Core Identity (experience, style, values, stance)
- ✅ Communication Principles (5 rules)
- ✅ Technical Approach
- ✅ Collaboration with Sub-Agents
- ✅ Caveman block (via `{{ include "_shared/caveman-identity-block.md" }}`)
- ✅ Rules (6 rules)
- ✅ Tools section

**MISSING:**
- ❌ Architecture Guardrails (5 constitutional rules)
- ❌ MCP tool availability listing with usage guidelines
- ❌ Sandbox Security rules (L0/L1 delegation constraints)
- ❌ Context-Mode routing policy (tool blocking/redirecting)

### 13.2 Generic Persona Files

`internal/assets/generic/persona-architect.md` and `persona-neutral.md` exist. They need the same additions.

### 13.3 Fix — Persona Must Include

```markdown
## Architecture Constitution (MANDATORY — govern all behavior)

1. **Source of Truth**: State lives in ONE place. No replication without sync.
2. **Thin Adapters**: Business logic in domain/core. Integrations are thin wrappers.
3. **Explicit Boundaries**: No hidden cross-system coupling in helpers/utilities.
4. **Mental Model First**: Fit new features into logical model BEFORE designing implementation.
5. **Sandbox Security**: L2 agents CANNOT perform destructive mutations without L0/L1 
   authorization. Stop, report RISK, defer to human if escalation required.

## Available MCP Tools

The following MCP servers are active (verify with tool probe at session start):

| Server | Tool examples | Use when |
|--------|--------------|----------|
| engram | mem_search, mem_save, mem_get_observation | Session memory, SDD artifacts |
| context7 | resolve-library-id, get-library-docs | External library/framework docs |
| sequential-thinking | sequential_thinking | Complex analysis, architectural exploration |
| context-mode | ctx_execute, ctx_batch_execute, ctx_fetch_and_index | Protecting context window |
| codegraph | codegraph_context, codegraph_trace, codegraph_callers, codegraph_impact | Semantic code exploration |
| notebooklm-mcp | notebooklm_* | Research synthesis (Mode 1/2 only) |

## Context-Mode Routing (MANDATORY)
{content of generic/context-mode-routing-policy.md}
```

### 13.4 Persona-Neutral — Missing Guardrails Too

`persona-neutral.md` likely has same gaps. Same additions apply.

---

## 14. Spanish Regionalism Audit

### 14.1 Files Affected

**`internal/assets/_shared/architect-identity.md` line ~10:**
```
Regex: \b(use sdd|start sdd|sdd mode|spec-driven|iniciar sdd|haceme un sdd)\b
```
🔴 `"haceme un sdd"` — River Plate regionalism (haceme = hazme in Rioplatense Spanish)

**`internal/assets/claude/sdd-orchestrator.md` Intent Resolution table:**
```
| (ES: "usa sdd", "vamos con sdd") |
| (ES: "sigue", "continua") |
| (ES: "guíame") |
| (ES: "investiga X") |
| (ES: "valida") |
| (ES: "cierra el cambio") |
| (ES: "rápido", "ff hasta tasks") |
```

**`internal/assets/gemini/sdd-orchestrator.md` (and all other agent variants):**
Same ES pattern columns in routing tables.

**`internal/assets/claude/general-orchestrator.md`:**
```
| "fix this", "why is X crashing", "solve", "debug", "research", "investigate" (EN + ES) |
```
Though the ES phrases are shown as column name, not literally defined.

**`internal/assets/generic/general-orchestrator.md`:**
No ES phrases in routing table (clean). ✅

### 14.2 Fix — Remove All Spanish Routing Patterns

Remove all `(ES: ...)` annotations and Spanish keyword variants from ALL routing tables in ALL agent orchestrators. The language policy doc (`docs/language-policy.md`) should clarify agents respond in the user's language but ROUTE based on English keywords only.

**Files to update:**
- `internal/assets/_shared/architect-identity.md` — remove `iniciar sdd`, `haceme un sdd`
- `internal/assets/claude/sdd-orchestrator.md` — remove all ES column entries
- `internal/assets/gemini/sdd-orchestrator.md`
- `internal/assets/cursor/sdd-orchestrator.md`
- `internal/assets/generic/sdd-orchestrator.md`
- `internal/assets/opencode/sdd-orchestrator.md`
- `internal/assets/kiro/sdd-orchestrator.md`
- `internal/assets/gga/sdd-orchestrator.md`
- All `general-orchestrator.md` files with routing table
- `internal/assets/claude/general-orchestrator.md` — routing table "EN + ES" column

**Keep**: Non-regional Spanish is acceptable in documentation (e.g., `Terra Incógnita` as a technical term).

---

## 15. Additional Critical Bugs Found

### 15.1 sdd-orchestrator.md — L1a/L1b Label Confusion

Claude `general-orchestrator.md` frontmatter:
```yaml
description: >
  L1a General Orchestrator.
```

Generic `general-orchestrator.md` header:
```
# L1b General Orchestrator Core (Generic)
```

🔴 **Critical**: Generic calls the General Orchestrator "L1b" but the correct designation is L1a (General = L1a, SDD = L1b). This inverts the L1a/L1b labeling in the generic agent.

All generic/opencode/cursor/gemini general-orchestrator.md files must be checked and corrected to use L1a consistently.

### 15.2 Adaptive Reasoning Gate Duplicate Inclusion

`internal/assets/claude/sdd-orchestrator.md` contains:
- `<!-- adaptive-reasoning-gate:START -->` block embedded INLINE
- Reference to `_shared/adaptive-reasoning-gate-v2.md` (which sub-agents also include)

🟠 **High**: The full D1-D4 matrix appears TWICE in sub-agent contexts (once from the orchestrator template, once from the v2 gate injected per sub-agent). This wastes ~200 tokens per sub-agent dispatch.

**Fix**: Remove the `<!-- adaptive-reasoning-gate:START/END -->` blocks from sdd-orchestrator.md and general-orchestrator.md. Let each sub-agent receive only the v2 gate file.

### 15.3 Generic sdd-archive.md — Duplicate Step Numbering

Looking at `internal/assets/generic/sdd-phase-protocols/sdd-archive.md`:
```
3b. Eval Gate Check: ...
4. Persistence (Learned Patterns)...
4. If OpenSpec mode: move change directory...   ← DUPLICATE step 4!
5. Update DAG state to "archived"
```

🟠 **High**: Step 4 is duplicated. The second step 4 should be step 5, and "Update DAG state" should be step 6. This exists in both Claude AND generic versions.

### 15.4 context7 Package Name Inconsistency

`generateClaude` uses `@upstash/context7-mcp@latest` (versioned)  
`overlay.go` Claude entry returns `@upstash/context7-mcp` (unversioned)  
`generateGemini` uses `httpUrl: https://mcp.context7.com/mcp` (remote)  
`generateOpenCode` uses `url: https://mcp.context7.com/mcp` + `type: remote`  

All non-Claude agents use the remote URL. Claude uses the npm package. Both work but inconsistency makes maintenance harder.

### 15.5 Gemini Generator vs Adapter Strategy Conflict

`adapter.go` for Gemini: `MCPStrategy: StrategyMergeIntoSettings` (merges into `~/.gemini/settings.json`)

`generateGemini` in `generator.go`: returns a full config map with `general`, `ide`, `mcpServers`, `model`, `security`, `ui` keys.

🟠 **High**: When MCP is injected via `StrategyMergeIntoSettings`, it calls `mergeJSONFile(settingsPath, overlay)` which deep-merges. But the overlay only contains the `mcpServers` section. When `generateGemini` is used for fresh installs, it writes a COMPLETE settings.json. If the user already has a settings.json with other keys, the fresh generate would OVERWRITE (not merge) their settings.

**Fix**: `generateGemini` should only generate the `mcpServers` section for the MCP injection path, not the full settings object.

### 15.6 sdd-verify.md — WCAG Check Out of Scope

`sdd-verify.md` (Claude and generic):
```
- [ ] **WCAG Compliance Check**: Verify aria-labels, contrast ratios, and keyboard accessibility.
```

This check is hardcoded in the general-purpose sdd-verify template but only applies to frontend/UI changes. For backend-only changes, this wastes agent effort and creates unnecessary failures.

**Fix**: Make WCAG check conditional:
```markdown
- [ ] **WCAG Check** (ONLY if change includes UI components):
      Verify aria-labels, contrast ratios, keyboard navigation.
      Skip if: backend-only, API-only, CLI-only, or data migration change.
```

### 15.7 sdd-init Missing Overlay Detection Trigger

`sdd-init.md` detects:
```
6. Active overlays: check for .atl/overlays/ directory
```

But doesn't set `IS_ODOO = true` or write `manifest.json` for the overlay. The overlay detection in `sdd-orchestrator.md` relies on `manifest.json` existing. There's a chicken-and-egg: sdd-init doesn't create `manifest.json`, sdd-orchestrator can't detect the overlay.

**Fix**: `sdd-init.md` must:
```markdown
### Step 6b: Overlay Registration (if Odoo detected)

IF Odoo project detected (pyproject.toml has odoo OR addons/ with __manifest__.py):
  Create: .atl/overlays/odoo-{version}/manifest.json
  Content: { "overlay": "odoo-development-skill", "version": "{detected_version}", "active": true }
  Set: IS_ODOO = true in session state
  Inject to session: sdd-orchestrator will use this for all subsequent phases
```

---

## 16. Consolidated Fix Priority Queue

### 🔴 Sprint 1 — Critical (Week 1-2)

| ID | Fix | Files |
|----|-----|-------|
| FIX-01 | Add `InjectCodeGraph()` to inject.go | `internal/components/mcp/inject.go` |
| FIX-02 | Fix Claude MCP key names (underscores→hyphens) | `generator.go:generateClaude()` |
| FIX-03 | Add notebooklm-mcp to ALL generators | `generator.go` (all generate* funcs) |
| FIX-04 | Create Antigravity CLI Go adapter + model ID | New `internal/agents/antigravity-cli/` |
| FIX-05 | Add antigravity-cli to go:embed + catalog | `assets.go`, `catalog/registry.go` |
| FIX-06 | Add CodeGraph steps to sdd-explore.md (all agents) | All `sdd-explore.md` files |
| FIX-07 | Add CodeGraph to research routing policy | All orchestrator research routing sections |
| FIX-08 | Add separate judge agent instantiation to sdd-verify | All `sdd-verify.md` files |
| FIX-09 | Add pre-apply completeness gate (spec/design/tasks) | All `sdd-apply.md` files |
| FIX-10 | Fix duplicate step 4 in sdd-archive.md | Claude + generic `sdd-archive.md` |
| FIX-11 | Fix L1b label in generic general-orchestrator | `generic/general-orchestrator.md` |
| FIX-12 | Add manifest.json creation to sdd-init.md | All `sdd-init.md` files |

### 🟠 Sprint 2 — High (Week 3-4)

| ID | Fix | Files |
|----|-----|-------|
| FIX-13 | Add README.md + __manifest__.py to sdd-archive.md | All `sdd-archive.md` files |
| FIX-14 | Add context-mode overlay for Cursor, Kiro, Windsurf, Qwen | `overlay.go` |
| FIX-15 | Fix max postures violation (3→2) in routing tables | All `general-orchestrator.md` |
| FIX-16 | Unify agent routing tables Claude↔Generic | Claude and generic `general-orchestrator.md` |
| FIX-17 | Remove Spanish regionalisms from all routing patterns | All orchestrators + architect-identity.md |
| FIX-18 | Add guardrails + MCP listing to all persona files | All `persona-architect.md` + `persona-neutral.md` |
| FIX-19 | Add Unified Posture Assignment Specification to gate-v2 | `_shared/adaptive-reasoning-gate-v2.md` |
| FIX-20 | Remove duplicate adaptive-reasoning blocks from orchestrators | All `sdd-orchestrator.md` + `general-orchestrator.md` |
| FIX-21 | Fix Gemini generator to only write mcpServers section | `generator.go:generateGemini()` |
| FIX-22 | Make WCAG check conditional in sdd-verify.md | All `sdd-verify.md` files |

### 🟡 Sprint 3 — Medium (Week 5-6)

| ID | Fix | Files |
|----|-----|-------|
| FIX-23 | Add `InjectContextMode()` to inject.go | `internal/components/mcp/inject.go` |
| FIX-24 | Add `InjectEngram()` to inject.go | `internal/components/mcp/inject.go` |
| FIX-25 | Add context-mode hook writer (`.claude/hooks/context-mode.json`) | New component |
| FIX-26 | Process `{{ include "..." }}` template directives | Go template processing at write time |
| FIX-27 | Add keyword index to skill-registry.md generator | `internal/app/skills_cmd.go` |
| FIX-28 | Add uv/uvx detection for Odoo MCP install | `internal/system/deps.go` |
| FIX-29 | Normalize context7 package name across all generators | `generator.go` |
| FIX-30 | Add D1-D4→posture mapping for Generalist in all orchestrators | All orchestrators |

---

## Summary Statistics

| Severity | Count | Percentage |
|----------|-------|-----------|
| 🔴 Critical | 12 | 40% |
| 🟠 High | 9 | 30% |
| 🟡 Medium | 9 | 30% |

**Most impactful fixes by area:**
- **MCP**: FIX-01, FIX-02, FIX-03 (generator/inject consistency)
- **Antigravity CLI**: FIX-04, FIX-05 (complete new adapter needed)
- **CodeGraph**: FIX-06, FIX-07 (tool configured but never used)
- **Verification**: FIX-08 (judge agents), FIX-09 (completeness gate)
- **Archive**: FIX-10, FIX-13 (file updates, numbering)
- **Cognitive postures**: FIX-15, FIX-16, FIX-19 (consistency + coverage)
- **Personas**: FIX-17, FIX-18 (remove regional Spanish, add guardrails)
