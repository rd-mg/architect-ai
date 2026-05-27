# Architect-AI

> A multi-agent framework that turns any supported IDE or CLI (Claude Code, Cursor, Gemini CLI, Codex, Antigravity, Kiro, OpenCode, VSCode, and more) into a Spec-Driven Development (SDD) workspace. One orchestrator, many agents, shared persistent memory via Engram, and curated research routing (ripgrep → local → Context7 → internet-by-permission).

**Status**: V3.3.0 — Parallel multi-agent architecture with pipeline parallelization, skill registry concurrency, orchestrator routing fork, context compression, and state hardening

---

## Table of Contents

1. [What it does](#what-it-does)
2. [Architecture](#architecture)
3. [Supported agents](#supported-agents)
4. [Install](#install)
5. [Quick start](#quick-start)
6. [SDD commands](#sdd-commands)
7. [Research routing policy](#research-routing-policy)
8. [Mandatory skills](#mandatory-skills)
9. [Session metering](#session-metering)
10. [Uninstall & purge](#uninstall--purge)
11. [Repo layout](#repo-layout)
12. [Version history](#version-history)
13. [Contributing](#contributing)

---

## What it does

Architect-AI installs a thin coordination layer on top of whatever coding agent you already use. The orchestrator:

- **Coordinates SDD phases**: explore, propose, spec, design, tasks, apply, verify, archive.
- **Enforces SDD pipeline compliance**: validates all phases complete before concluding; halts apply if TDD specs are missing; prohibits test weakening during verify; commits and updates docs during archive.
- **Injects adaptive reasoning** and phase-specific cognitive posture before delegation.
- **Applies Caveman** as output compression across all interactions, including inline executions and tool outputs; it does not change reasoning depth.
- **Mandates tool execution** (Context-Mode): uses designated tools over generic model capabilities; never hallucinates tool outputs.
- **Supports dual orchestration modes**: Primary (Task tool delegation for Claude/Gemini/OpenCode) and Fallback (inline execution for Antigravity/Cursor/Copilot/Windsurf).
- Uses Engram as curated working memory with progressive disclosure.
- Uses OpenSpec files when team-readable audit trail is required.
- Uses ripgrep/local evidence before external documentation calls.
- Uses Context7 or official docs for current library/API documentation.
- Supports NotebookLM MCP as an optional tool when configured.
- Detects context-mode as an external companion runtime and routes around it safely.
- Manages installed files through ownership manifests for safe sync/uninstall/purge.
- Runs hooks and CLI subprocesses through bounded execution for runtime stability.

---

## Architecture

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                         USER (in editor / CLI)                              │
│              /sdd-new, /sdd-ff, "use sdd", "continue"                       │
└─────────────────────────────┬───────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                       SDD ORCHESTRATOR (per agent)                          │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │ Global System Directives (Caveman + Context-Mode)                  │    │
│  │ Intent Resolution → Session-Setup Triplet → Overlay Detection      │    │
│  │ Research-Routing Policy → Mandatory-Skills List → Model Assignment │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                              │                                              │
│              ┌───────────────┼───────────────┐                              │
│              ▼               ▼               ▼                              │
│        Phase Protocol   Cognitive Posture   Adaptive-Reasoning              │
│        (on-demand load) (6→8 postures)      (CLASSIFIER — MANDATORY)        │
│                                              │                              │
│                                              ▼                              │
│                                      Mode 1 / 2 / 3                         │
└─────────────────────────────┬───────────────────────────────────────────────┘
                              │  Sub-agent launch (Task / runSubagent / inline)
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│   SUB-AGENT (sdd-explore, sdd-propose, sdd-design, sdd-apply, sdd-verify…)  │
│                                                                             │
│   ┌───── Prompt Layers (stacked) ─────────────────────────────────────┐   │
│   │ 0. Global System Directives (Caveman + Context-Mode — MANDATORY)  │   │
│   │ 1. Cognitive Posture       (+++Socratic / +++Critical / +++Adv.…) │   │
│   │ 2. Adaptive Classifier     (score 4 dims → Mode 1/2/3 — REQUIRED) │   │
│   │ 3. Project Standards       (from .atl/skill-registry.md)          │   │
│   │ 4. Overlay Supplement      (if Odoo: sdd-supplements/{phase}.md)  │   │
│   │ 5. Research-Routing Policy (local → Context7 → NotebookLM → web)  │   │
│   │ 6. Available Tools         (from tool-availability probe)         │   │
│   │ 7. Phase Protocol          (sdd-phase-protocols/{phase}.md)       │   │
│   │ 8. Task                    (what to do)                           │   │
│   │ 9. Artifact-Store + Exec-Mode                                     │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   Hooks:  before_model  ← redact PII, enforce research routing              │
│           after_model   ← persist Context7/NotebookLM results to Engram     │
└────────────────┬─────────────────────────────┬──────────────────────────────┘
                 │                             │
                 ▼                             ▼
┌────────────────────────────┐   ┌──────────────────────────────────────────┐
│   ARTIFACT STORE           │   │   PERSISTENT MEMORY (Engram — MCP)       │
│   • engram (default)       │   │   mem_save / mem_search / mem_context    │
│   • openspec/changes/…     │   │   topic_key:  sdd/{change}/state         │
│   • hybrid                 │   │                context-pack/{project}/…  │
│   • none (inline)          │   │                skill-registry            │
└────────────────────────────┘   └──────────────────────────────────────────┘

                   ┌──────────────────────────────┐
                   │  SKILLS (global, always-on)  │
                   │  • adaptive-reasoning        │◄──── MUST be injected
                   │  • cognitive-mode (postures) │
                   │  • context-guardian          │
                   │  • ripgrep, bash-expert      │
                   │  • mcp-notebooklm-orchestr.  │
                   └──────────────────────────────┘

                   ┌──────────────────────────────┐
                   │  RULES (compact, injected)   │
                   │  • .atl/skill-registry.md    │◄──── single discovery surface
                   │    ├─ Project Standards       │
                   │    ├─ User Skills (triggers)  │
                   │    └─ Overlay-contributed     │
                   │       rules (NEW: emit here)  │
                   └──────────────────────────────┘

                   ┌──────────────────────────────┐
                   │  OVERLAYS (demand-loaded)    │
                   │  .atl/overlays/odoo-{v}/     │
                   │   ├─ manifest.json            │
                   │   ├─ skills/                  │
                   │   │   ├─ patterns-agnostic/  │
                   │   │   ├─ patterns-{v}/       │◄── version-gated by
                   │   │   └─ migration-{a}-{b}/  │    overlay.go at install
                   │   ├─ sdd-supplements/        │
                   │   │   ├─ explore-odoo.md     │◄── auto-injected per phase
                   │   │   ├─ propose-odoo.md     │
                   │   │   ├─ design-odoo.md      │
                   │   │   ├─ apply-odoo.md       │
                   │   │   ├─ verify-odoo.md      │
                   │   │   └─ domain-map.md (DDD) │
                   │   └─ rules/ (cudio-*)        │
                   └──────────────────────────────┘
```

---

## Supported agents

Nine agents receive orchestrator assets. All are SDD-capable (`sdd-orchestrator.md` + `general-orchestrator.md`). VSCode operates inline-only (no native sub-agent support) but follows the full SDD protocol:

| Agent | Runtime | Parallel sub-agents | Prompt cache visible |
|-------|---------|:-:|:-:|
| Claude Code | CLI / Desktop |  |  |
| Antigravity | IDE |  simulated |  |
| Codex | CLI |  |  |
| Cursor | IDE |  |  (provider-dependent) |
| Gemini CLI | CLI |  |  |
| Kilocode | IDE |  |  |
| Kiro | IDE |  |  |
| OpenCode | CLI |  |  (per profile) |
| Qwen | CLI |  |  |
| VSCode | IDE |  inline-only |  |
| Windsurf | IDE |  |  |

**Template** (not an agent):

| Asset | Purpose |
|-------|---------|
| Generic | Canonical orchestrator template — copied and agent-name-patched per agent at install time |

**Component** (not an agent):

| Component | Purpose |
|-----------|---------|
| GGA | Gentleman Guardian Angel — AI provider switcher, installed via `architect-ai --component gga` |

See `docs/antigravity-sdd-workaround.md` for notes on Antigravity's single-threaded sub-agent simulation.

---

## Install

```bash
# Prerequisites
go version          # 1.24+
git --version
rg --version        # ripgrep — mandatory for sub-agent searches

# Build from source
git clone https://github.com/rd-mg/architect-ai.git
cd architect-ai
go build -o architect-ai ./cmd/architect-ai

# Or via package manager (when released)
brew install architect-ai
```

Run the TUI and pick "Start installation":

```bash
architect-ai
```

The installer detects which agents you have locally and wires the orchestrator + skills + phase protocols into each.

---

## Quick start

```bash
cd your-project
architect-ai  # installs, if not already installed

# Open any supported agent and type either:
#   /sdd-new add-user-export
# or, as natural language (V3.1):
#   "use sdd to add user export"
```

On the **first SDD command of the session** the orchestrator will ask:

1. **Artifact store?** — `engram` / `openspec` / `hybrid` / `none`
2. **Execution mode?** — `interactive` (pause between phases) / `auto` (run through)
3. **Change name?** — short slug (e.g. `add-user-export`)

After that it runs the SDD cycle, asking for confirmation between phases if you chose interactive mode.

---

## SDD commands

Slash commands (auto-complete in every agent):

| Command | Effect |
|---------|--------|
| `/sdd-init` | Probe project context (usually auto-run) |
| `/sdd-onboard` | Guided first-time walkthrough |
| `/sdd-new <name>` | Start a new change — runs the full cycle |
| `/sdd-continue [name]` | Resume at the next dependency-ready phase |
| `/sdd-ff <name>` | Fast-forward: proposal → spec → design → tasks |
| `/sdd-explore <topic>` | Single-phase research (Socratic) |
| `/sdd-apply [name]` | Implement tasks in batches |
| `/sdd-verify [name]` | Adversarial validation against specs |
| `/sdd-archive [name]` | Close out a change; persists final report |

**Natural-language triggers** (V3.1) — the orchestrator also recognizes free-text intent:

- "use sdd" / "usa sdd" → `/sdd-new`
- "continue" / "continua" → `/sdd-continue`
- "fast forward" / "ff" → `/sdd-ff`
- "onboard me" / "guíame" → `/sdd-onboard`

On match, the orchestrator confirms the interpretation before acting.

### CLI Utility Commands

**1. Installation & Updates**
- `architect-ai install`: Configure AI coding agents on this machine.
- `architect-ai sync`: Sync agent configs and skills to current version.
- `architect-ai update`: Check for available updates for managed tools.
- `architect-ai upgrade`: Apply updates to managed tools (architect-ai, engram, gga).

**2. SDD Lifecycle (`sdd-init`, `sdd-status`)**
- `architect-ai sdd-init`: Initialize SDD context in the current project.
  ```bash
  architect-ai sdd-init --mode engram
  ```
- `architect-ai sdd-status`: View active SDD change state and progress.
- `architect-ai sdd-archive-preflight`: Check if a change is ready for archival.

**3. Project Maintenance**
- `architect-ai skill-registry`: Generate or refresh the project's `.atl/skill-registry.md`.
- `architect-ai overlay`: Manage specialist project overlays (e.g. Odoo).
- `architect-ai skills`: Manage community skills and project-specific learned patterns.

**4. System & Recovery**
- `architect-ai cleanup`: Deep clean of temporary files and logs.
- `architect-ai restore`: Restore a configuration backup from `~/.architect-ai/backups/`.
- `architect-ai uninstall`: Remove Architect AI managed files from this machine.

---

## Using Agents & Skills (Examples)

Architect-AI agents use "skills" (specialized instruction sets) that are automatically loaded when specific keywords or contexts are detected in your prompts.

### General Project Workflow

In a standard project, you trigger SDD phases and general skills:

**1. Starting a new feature**
> "use sdd to create a new authentication endpoint"
*(Triggers `/sdd-new` and loads the `sdd-propose` skill)*

**2. Writing Tests (Go Example)**
> "write tests for auth.go using teatest"
*(Detects the context and automatically loads the `go-testing` skill before writing code)*

**3. Code Review / Adversarial Mode**
> "judgment day for this PR"
*(Triggers the `judgment-day` skill for a rigorous adversarial review)*

### Odoo Project Workflow

When working in an Odoo repository (detected via `sdd-init`), specialized Odoo skills are available. The agent will automatically use Odoo best practices.

**1. Creating a new model**
> "create an Odoo model to manage vehicle reservations"
*(The agent recognizes the Odoo context and loads the `odoo-development-skill` overlay to ensure correct inheritance, security rules, and view definitions)*

**2. Debugging Odoo**
> "find why the tax calculation is failing on invoices"
*(The agent uses Odoo-specific debugging strategies, checking `account.move` overrides and server logs)*

### Analyst Skill Profile

The **Analyst** is a hybrid researcher-solver profile used for complex troubleshooting. It combines forensic evidence gathering with systemic problem-solving.

**1. Diagnosing a crash**
> "Analyze this stack trace and find the root cause in the local source"
*(Loads the `analyst` skill; uses `ripgrep` to trace paths and `Engram` to check for similar known bugs)*

**2. API Research & Integration**
> "How do I integrate the new Context7 telemetry in our backend?"
*(The Analyst queries documentation via `Context7` and proposes a low-blast-radius implementation)*

### Odoo 19 Spreadsheet Dashboard Architect (`available`)

Architect-AI includes a specialized skill for native Odoo 19 dashboard spreadsheet (`.osheet` / `.osps` JSON) work.

Use it for:

- Designing Odoo 19 dashboard spreadsheets.
- Creating or refactoring native o-spreadsheet JSON.
- Building dashboards with pivots, lists, scorecards, Odoo charts, global filters, Odoo formulas, and Odoo drill-down links.
- Validating `.osheet` files before importing or storing them in Odoo.
- Recreating sample Odoo dashboard JSON files for regression tests.

The skill is separate from generic XLSX editing. XLSX import/export remains handled by `odoo-minimax-xlsx-o-spreadsheets`, while this skill focuses on native Odoo dashboard semantics.

**Safety posture**: Uses Odoo pivots, lists, formulas, and native dashboard data sources by default. Does not vendor Dashboard Ninja code. Does not generate raw SQL dashboards. Respects Odoo model access and record rules.

**Developer commands:**
```bash
python3 osheet_profile.py  Sales.osheet.json --markdown
python3 osheet_validate.py Sales.osheet.json --strict
python3 osheet_recipe.py   export Sales.osheet.json Sales.recipe.json
python3 osheet_recipe.py   build  Sales.recipe.json Sales.recreated.osheet.json
python3 osheet_compare.py  Sales.osheet.json Sales.recreated.osheet.json --mode exact
python3 osheet_build.py    monthly-sales.json monthly-sales.osheet.json --pretty
```

> Status: `available` — skill, scripts, references, registry integration, and Go tests all complete.

---

## Research routing policy

External research follows a strict priority (ripgrep/local-first by default in V3.2):

```
1. Local code + docs   ← PRIMARY — ripgrep, find, cat, extract-text
2. Context7            ← SECONDARY — framework / library docs
3. NotebookLM          ← OPTIONAL — curated project knowledge (if configured)
4. Internet            ← ONLY on EXPLICIT user request
                         ("search the web", "look online", "busca en internet")
```

This is enforced by the orchestrator and by the `sdd-explore` phase protocol. Each sub-agent returns `research_sources_used: [...]` in its envelope so the orchestrator can audit routing compliance.

### Caveman output compression

Caveman is an output style, not a thinking mode. It shortens wording and removes filler where safe. It does not replace adaptive reasoning, SDD phase posture, code review, security checks, or deterministic verification.

Use normal explicit English for destructive actions, security warnings, commands that could delete data, and user confirmations.

### context-mode

context-mode MCP can be installed and configured outside Architect-AI. By default, Architect-AI detects it and adds routing awareness, not vendor it or overwrite its generated configuration.

---

## Mandatory skills

Four skills are marked `bridge: always` in their frontmatter and injected into **every** sub-agent prompt, regardless of task matcher:

- **`ripgrep`** — all code search must use `rg`, not `grep -r`
- **`bash-expert`** — strict-mode shell discipline
- **`architecture-guardrails`** — enforces system boundaries and state flow
- **`context-guardian`** — auto-invoked at > 50% window usage
- **`mcp-notebooklm-orchestrator`** — optional research source (query-only)

See each skill's `SKILL.md` in `internal/assets/skills/` for the compact rules.

---

## Session metering

At session end (exit, ctrl+c, or `/end`), the orchestrator prints a summary banner:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Session summary (claude) — 4m 32s
  Requests:         12
  Total tokens:     47,120
  From cache:       18,450 (39%)
  Est. savings:     ~$0.06
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Pricing estimates are approximate and use the table in `internal/metering/pricing.go`. Override with `SetPricing()` at startup for contract pricing.

Session stats are also persisted to Engram under `metering/{project}/{session-id}` so `sdd-archive` can include them in the final report.

---

## Uninstall & purge

Two levels of uninstall:

### Managed uninstall

```bash
architect-ai uninstall
```

Removes files that Architect-AI owns based on the managed manifest: agent config files, skill registry, orchestrator prompts, SDD phase protocols.

Keeps: User-owned configuration, Engram memories, `.atl/` artifacts, `~/.architect-ai/` backups, the binary itself.

### Deep purge

```bash
architect-ai uninstall --purge --purge-scope all --confirm PURGE
```

Removes managed config PLUS any subset of project/global data. You can perform a targeted purge by running `architect-ai` without arguments to access the TUI, which now supports selective agent purging and granular scope control. A **pre-purge snapshot** is captured in every case, allowing restore with `architect-ai restore <snapshot-path>`.

### Refactoring scope and generated directories

Root hidden directories such as `.agent/`, `.agents/`, `.atl/`, `.claude/`, `.gemini/`, `.cursor/`, `.vscode/`, `.github/`, `.opencode/`, `.engram/`, and `.context-mode/` are generated/runtime/config outputs. Do not treat them as source refactoring targets.

---

## Repo layout

```
architect-ai/
├── cmd/architect-ai/               # CLI entry point
├── docs/                           # User-facing documentation
├── internal/
│   ├── agentbuilder/               # Dynamic agent generation
│   ├── agents/                     # Per-agent adapters (Claude, Gemini, etc.)
│   ├── app/                        # Application entry and command routing
│   ├── assets/                     # Prompts, skills, personas, overlays
│   │   ├── skills/                 # Cross-cutting skills (ripgrep, bash-expert, etc.)
│   │   └── overlays/               # Specialist project overlays (Odoo, etc.)
│   ├── backup/                     # Config snapshot and restore logic
│   ├── catalog/                    # Registry of available agents and components
│   ├── cli/                        # CLI command implementations
│   ├── components/                 # Shared logic (GGA, SDD, hooks, etc.)
│   ├── installcmd/                 # Installation pipeline stages
│   ├── metering/                   # Session stats and token pricing
│   ├── model/                      # Shared domain models and types
│   ├── opencode/                   # OpenCode-specific multi-mode logic
│   ├── pipeline/                   # Sequential execution orchestrator
│   ├── planner/                    # Install/Sync dependency resolver
│   ├── process/                    # Bounded subprocess runner
│   ├── scope/                      # Context boundary management
│   ├── skills/                     # Skill resolution and loading logic
│   ├── state/                      # Local state persistence (state.json)
│   ├── system/                     # OS/Platform detection and validation
│   ├── tui/                        # Bubbletea-based interactive UI
│   ├── update/                     # Update check and self-update logic
│   └── verify/                     # Installation verification reports
├── openspec/                       # (optional) file-based artifacts
│   └── changes/
└── README.md
```

---

## Version history

| Version | Date | Highlights |
|---------|------|------------|
| **V3.3.0** | 2026-05 | Gentleman Angel suite: Nonce DAG for state sync, Empirical Feedback Loop (+++Empirical) in phase protocols, context-aware pipeline integration (context.Context hardening), differential context injection via skill-registry.md indexing. |
| **V3.2.1** | 2026-05 | Orchestrator directive improvements: Global System Directives (Caveman + Context-Mode), Primary vs Fallback orchestration modes, SDD Pipeline Enforcement (workflow validation, TDD prerequisite lock, testing protocol strictures, archival sequence with commit + doc updates). |
| **V3.2** | 2026-05 | Manifest-first purge, bounded subprocess runner, Caveman output-only clarification, Engram tool routing, context-mode external companion policy, dot-directory scope guard, optional NotebookLM policy. |
| V3.1 | 2026-04 | Artifact-store question asked explicitly; natural-language intent resolution; TUI deep purge; token-cache banner; ripgrep + bash-expert `bridge: always`; NotebookLM-first research routing |
| V3.0 | 2026-04 | V2 absorbed into caveman dual-mode; Odoo overlay restructured with version-gated bundles; 6 Odoo sub-agents absorbed into SDD supplements; judgment-day + autoreason-lite folded into adaptive-reasoning v1.0 |
| V2.x | 2025-Q4 | Multi-agent orchestrator; Engram integration |
| V1.x | 2025-Q3 | Initial SDD implementation |

See `plans/v3.1-remediation-plan.md` for the full V3.1 change list.

---

## Contributing

- Work only in `internal/assets/` and `internal/metering/` unless the plan explicitly instructs otherwise
- Phase protocols live in `internal/assets/{agent}/sdd-phase-protocols/` — one file per phase
- All orchestrators share the canonical Claude version with only the Delegation Syntax paragraph varying per agent; regenerate with `architect-ai sync --component sdd` when the canonical changes
- Any new skill that should apply everywhere must carry `bridge: always` in its frontmatter — the resolver respects it (V3.1)

Tests:

```bash
go test ./...                                    # all
go test ./internal/metering/...                  # session stats
go test ./internal/components/uninstall/...      # purge
go test ./internal/tui/...                       # screens
```

PR template: open against `main`, include a one-paragraph summary of intent and a `## Verification` section with commands and expected output. Green CI is required.

---

## License

Apache 2.0 (see `LICENSE`). Skills under `internal/assets/skills/` carry their own per-skill licenses in their frontmatter.

## Language

This project is English-only in voice. See [`docs/language-policy.md`](docs/language-policy.md) for the precise rule and the allowlist process for trigger phrases.
