---
name: sdd-init
description: >
  Initialize Spec-Driven Development context in any project. Detects stack, conventions, testing capabilities, and bootstraps the active persistence backend.
  Trigger: When user wants to initialize SDD in a project, or says "sdd init", "iniciar sdd", "openspec init".
license: MIT
metadata:
  author: rd-mg
  version: "3.0"
---

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Sub-agent for initializing SDD context in a project. Detect project stack, conventions, testing capabilities, bootstrap active persistence backend.

EXECUTOR for this phase, not orchestrator. Do initialization work yourself. Do NOT launch sub-agents, do NOT call `delegate` or `task`, do NOT hand execution back unless hitting real blocker that must be reported upstream.

## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching.

### Artifact 1: Project Context
- **Artifact Name**: config.yaml (openspec/hybrid)
- **Topic Key**: sdd-init/{project-name}
- **Type**: architecture

### Artifact 2: Testing Capabilities
- **Artifact Name**: config.yaml section (openspec/hybrid)
- **Topic Key**: sdd/{project-name}/testing-capabilities
- **Type**: config

## What to Do

### Step 0: Pre-flight Validation (MANDATORY)

Validate environment fit for SDD before any analysis. Failure gates all further work.

```markdown
VALIDATE:
├── Required Tools: Check `rg`, `git`, project's build tool (npm, go, etc.) available.
├── Writable Paths: Verify project root and `openspec/` (if mode includes openspec) writable.
└── Critical MCPs: Verify `engram` and `context7` listed in `## Available Tools`.
```

If any critical validation fails:
1. STOP immediately.
2. Return status `pre-flight-fail`.
3. List missing/failing items in `risks` section of envelope.
4. Set `executive_summary` to "Pre-flight validation failed. [Reason]".

### Step 1: Detect Project Context

Read project to understand:
- Tech stack (check package.json, go.mod, pyproject.toml, etc.)
- Existing conventions (linters, test frameworks, CI)
- Architecture patterns in use

### Step 1b: Surface Mapping (MANDATORY)

Identify project's public interface and entry points.

```markdown
SCAN FOR:
├── Entry Points: CLI commands (`main.go`, `cli.ts`), server routes (`routes/`), or hooks.
├── Public APIs: Exported functions, classes, REST/GraphQL endpoints.
├── Core Interfaces: Key data structures or domain interfaces defining the system.
└── Key Dependencies: External services or critical internal modules.
```

Return brief "Surface Map" in detailed report or executive summary.

### Step 2: Detect Testing Capabilities

Scan project for ALL testing infrastructure.

```
Detect testing capabilities:
├── Test Runner
│   ├── package.json → devDependencies: vitest, jest, mocha, ava
│   ├── package.json → scripts.test (what command it runs)
│   ├── pyproject.toml / pytest.ini / setup.cfg → pytest
│   ├── go.mod → go test (built-in)
│   ├── Cargo.toml → cargo test (built-in)
│   ├── Makefile → make test
│   └── Result: {framework name, command} or NOT FOUND
│
├── Test Layers
│   ├── Unit: test runner exists → AVAILABLE
│   ├── Integration:
│   │   ├── JS/TS: @testing-library/* in dependencies
│   │   ├── Python: pytest + httpx/requests-mock/factory-boy
│   │   ├── Go: net/http/httptest (built-in)
│   │   ├── .NET: xUnit/NUnit + WebApplicationFactory
│   │   └── Result: AVAILABLE or NOT INSTALLED
│   ├── E2E:
│   │   ├── playwright, cypress, selenium in dependencies
│   │   ├── Python: playwright, selenium
│   │   ├── Go: chromedp
│   │   └── Result: AVAILABLE or NOT INSTALLED
│   └── Each layer → record tool name
│
├── Coverage Tool
│   ├── JS/TS: vitest --coverage, jest --coverage, c8, istanbul/nyc
│   ├── Python: coverage.py, pytest-cov
│   ├── Go: go test -cover (built-in)
│   ├── .NET: coverlet
│   └── Result: {command} or NOT AVAILABLE
│
└── Quality Tools
    ├── Linter: eslint, pylint, ruff, golangci-lint, clippy
    ├── Type checker: tsc --noEmit, mypy, pyright, go vet
    ├── Formatter: prettier, black, gofmt, rustfmt
    └── Each: {command} or NOT AVAILABLE
```

### Step 3: Resolve STRICT TDD MODE

Determine whether Strict TDD Mode enabled. Priority chain — first match wins:

```
1. Read from system prompt / agent config (highest priority):
   ├── Search for "strict-tdd-mode" marker in agent's system prompt file
   │   (e.g., CLAUDE.md, GEMINI.md, .cursorrules, etc.)
   ├── If found and "enabled" → strict_tdd: true
   ├── If found and "disabled" → strict_tdd: false
   └── Preference set by user in architect-ai TUI

2. If no marker found, check openspec config:
   ├── Read openspec/config.yaml → strict_tdd field
   └── If found → use that value

3. If nothing found AND test runner detected in Step 2:
   ├── Default: strict_tdd: true (enable if project CAN do TDD)
   └── Ensures TDD active even without architect-ai TUI setup

4. If no test runner detected:
   ├── strict_tdd: false (cannot enable without test runner)
   └── Include NOTE: "Strict TDD Mode unavailable — no test runner detected"
```

**Do NOT ask user interactively.** Preference resolved from existing config. To change, user runs `architect-ai sync` with TUI or sets `strict_tdd` in `openspec/config.yaml`.

### Step 4: Initialize Persistence Backend

If mode resolves to `openspec`, create:

```
openspec/
├── config.yaml              ← Project-specific SDD config
├── specs/                   ← Source of truth (empty initially)
└── changes/                 ← Active changes
    └── archive/             ← Completed changes
```

### Step 5: Generate Config (openspec mode)

Based on detected context, create config when in `openspec` mode:

```yaml
# openspec/config.yaml
schema: spec-driven

context: |
  Tech stack: {detected stack}
  Architecture: {detected patterns}
  Testing: {detected test framework}
  Style: {detected linting/formatting}

strict_tdd: {true/false}

rules:
  proposal:
    - Include rollback plan for risky changes
    - Identify affected modules/packages
  specs:
    - Use Given/When/Then format for scenarios
    - Use RFC 2119 keywords (MUST, SHALL, SHOULD, MAY)
  design:
    - Include sequence diagrams for complex flows
    - Document architecture decisions with rationale
  tasks:
    - Group tasks by phase (infrastructure, implementation, testing)
    - Use hierarchical numbering (1.1, 1.2, etc.)
    - Keep tasks small enough to complete in one session
  apply:
    - Follow existing code patterns and conventions
    - Load relevant coding skills for project stack
  verify:
    - Run tests if test infrastructure exists
    - Compare implementation against every spec scenario
  archive:
    - Warn before merging destructive deltas (large removals)
```

### Step 6: Persist Testing Capabilities

**MANDATORY — do NOT skip.**

Persist detected testing capabilities as separate Engram observation (or section in config.yaml for openspec). Cache prevents re-detection on every `sdd-apply` and `sdd-verify` run.

If mode `engram` or `hybrid`:
Follow persistence rules in Step 2 of `_shared/mode-branching.md` using **Artifact 2** metadata.

**Testing Capabilities format**:

```markdown
## Testing Capabilities

**Strict TDD Mode**: {enabled/disabled}
**Detected**: {date}

### Test Runner
- Command: `{command}`
- Framework: {name}

### Test Layers
| Layer | Available | Tool |
|-------|-----------|------|
| Unit |  /  | {tool or —} |
| Integration |  /  | {tool or —} |
| E2E |  /  | {tool or —} |

### Coverage
- Available:  / 
- Command: `{command or —}`

### Quality Tools
| Tool | Available | Command |
|------|-----------|---------|
| Linter |  /  | {command or —} |
| Type checker |  /  | {command or —} |
| Formatter |  /  | {command or —} |
```

If mode `openspec` or `hybrid`, also write as section in `openspec/config.yaml` under `testing:`.

### Step 7: Build Skill Registry

Follow same logic as `skill-registry` skill (`skills/skill-registry/SKILL.md`):

1. Scan user skills: glob `*/SKILL.md` across ALL known skill directories. **User-level**: `~/.claude/skills/`, `~/.config/opencode/skills/`, `~/.gemini/skills/`, `~/.cursor/skills/`, `~/.copilot/skills/`, parent of this skill file. **Project-level**: `.claude/skills/`, `.gemini/skills/`, `.agent/skills/`, `skills/`. Skip `sdd-*`, `_shared`, `skill-registry`. Deduplicate by name (project-level wins). Read frontmatter triggers.
2. Scan project conventions: check `agents.md`, `AGENTS.md`, `CLAUDE.md` (project-level), `.cursorrules`, `GEMINI.md`, `copilot-instructions.md` in project root. If index file found (e.g., `agents.md`), READ it and extract all referenced file paths — include both index and referenced files in registry.
3. **ALWAYS write `.atl/skill-registry.md`** in project root (create `.atl/` if needed). Mode-independent — infrastructure, not SDD artifact.
4. If engram available, **ALSO save to engram**: `mem_save(title: "skill-registry", topic_key: "skill-registry", type: "config", project: "{project}", content: "{registry markdown}")`

See `skills/skill-registry/SKILL.md` for full registry format and scanning details.

### Step 8: Persist Project Context

**MANDATORY — do NOT skip.**
Follow persistence rules in Step 2 of `_shared/mode-branching.md` using **Artifact 1** metadata.

### Step 9: Return Summary

Return structured summary adapted to resolved mode:

#### If mode `engram`:

Persist project context following `skills/_shared/engram-convention.md` with title and topic_key `sdd-init/{project-name}`.

```
## SDD Initialized

**Project**: {project name}
**Stack**: {detected stack}
**Persistence**: engram
**Strict TDD Mode**: {enabled  / disabled  / unavailable (no test runner)}

### Testing Capabilities
| Capability | Status |
|------------|--------|
| Test Runner | {tool}  /  Not found |
| Unit Tests |  /  |
| Integration Tests | {tool}  /  Not installed |
| E2E Tests | {tool}  /  Not installed |
| Coverage |  /  |
| Linter | {tool}  /  |
| Type Checker | {tool}  /  |

### Context Saved
Project context persisted to Engram.
- **Engram ID**: #{observation-id}
- **Topic key**: sdd-init/{project-name}
- **Capabilities ID**: #{capabilities-observation-id}
- **Capabilities key**: sdd/{project-name}/testing-capabilities

No project files created.

### Engram Mode Limitations
Engram mode ideal for **solo developers** doing fast iteration. Be aware:
- **No iteration history**: re-running phase (e.g., `sdd-spec`) overwrites previous version. Only latest artifact retained.
- **Not shareable**: engram is local database — team members cannot see SDD artifacts.
- **Partial audit trail**: archive phase saves summary report, not full artifact folder.

For **team projects** or work needing full audit trail, consider switching to `openspec` (file-based, git-friendly) or `hybrid` (files + engram recovery).

### Next Steps
Ready for /sdd-explore <topic> or /sdd-new <change-name>.
```

#### If mode `openspec`:
```
## SDD Initialized

**Project**: {project name}
**Stack**: {detected stack}
**Persistence**: openspec
**Strict TDD Mode**: {enabled  / disabled  / unavailable (no test runner)}

### Testing Capabilities
{same table as above}

### Structure Created
- openspec/config.yaml ← Project config with detected context + testing capabilities
- openspec/specs/      ← Ready for specifications
- openspec/changes/    ← Ready for change proposals

### Next Steps
Ready for /sdd-explore <topic> or /sdd-new <change-name>.
```

#### If mode `none`:
```
## SDD Initialized

**Project**: {project name}
**Stack**: {detected stack}
**Persistence**: none (ephemeral)
**Strict TDD Mode**: {enabled  / disabled  / unavailable (no test runner)}

### Testing Capabilities
{same table as above}

### Context Detected
{summary of detected stack and conventions}

### Recommendation
Enable `engram` or `openspec` for artifact persistence across sessions. Without persistence, all SDD artifacts lost when conversation ends.

### Next Steps
Ready for /sdd-explore <topic> or /sdd-new <change-name>.
```

## Rules

- NEVER create placeholder spec files — specs created via sdd-spec during change
- ALWAYS detect real tech stack, don't guess
- NEVER behave like orchestrator from this phase — execute directly, return results
- If project already has `openspec/` directory, report what exists and ask orchestrator if should be updated
- Keep config.yaml context CONCISE — no more than 10 lines
- ALWAYS perform Surface Mapping scan, return in summary
- ALWAYS detect testing capabilities — not optional
- ALWAYS persist testing capabilities as separate observation/section — downstream phases depend on it
- If Strict TDD Mode requested but no test runner exists, set strict_tdd: false and explain why
- Return structured envelope with: `status`, `executive_summary`, `detailed_report` (optional), `artifacts`, `next_recommended`, `risks`
