---
name: skill-registry
description: >
  Create or update the skill registry for the current project. Scans user skills and project conventions, writes .atl/skill-registry.md, and saves to engram if available.
  Trigger: When user says "update skills", "skill registry", "actualizar skills", "update registry", or after installing/removing skills.
license: MIT
metadata:
  author: rd-mg
  version: "2.0"
---

## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching (metadata only). Fixed hybrid-style persistence for infrastructure availability.

- **Artifact Name**: .atl/skill-registry.md
- **Topic Key**: skill-registry
- **Type**: config

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Generate or update **skill registry** — dynamic catalog of available skills with **compact rules** (pre-digested, 5-15 line summaries).

## Lazy Load Protocol

Project uses lazy-load architecture:
1. Skills discovered via `glob` dynamically at runtime.
2. Kernel's intent router handles top-level routing.
3. This registry generates compact rules for injection into sub-agent prompts.
4. No static index maintained in kernel.

## When to Run

- After installing or removing skills
- After setting up new project
- When user explicitly asks to update registry
- As part of `sdd-init` (calls same logic)

## Steps

### Step 1: Scan Skills

1. **Auto-Installed Skills (skills-lock.json)**:
   - Check for `skills-lock.json` in project root.
   - Parse to get list of auto-installed skill names.
   - For each auto-installed skill, check if present in filesystem (maybe cached).
   - If cached, include in scan.
   - If not cached, note as "available but not downloaded".
2. **Filesystem Skills (glob)**:
   - Glob `internal/assets/skills/**/SKILL.md` for project-level skills.
   - Glob `.agent/skills/**/SKILL.md` for workspace-level skills.
3. **SKIP** `_shared`, `_archived`, and `skill-registry`.
4. **Deduplicate**: Auto-installed > Project-level > Workspace-level.
5. For each skill, read **full SKILL.md** (>200 lines → focus on frontmatter and Critical Patterns/Rules) to extract:
   - `name` field
   - `description` field
   - **Compact rules** (see Step 1b)
6. Do NOT build static trigger table.

### Step 1b: Generate Compact Rules

For each skill, generate **compact rules block** (5-15 lines max) containing ONLY:
- Actionable rules and constraints ("do X", "never Y", "prefer Z over W")
- Key patterns with one-line examples where critical
- Breaking changes or gotchas causing bugs if missed

**DO NOT include**: purpose/motivation, when-to-use, full code examples, installation steps.

Format per skill:
```markdown
### {skill-name}
- Rule 1
- Rule 2
```

**Compact rules are MOST IMPORTANT output.**

### Step 2: Scan Project Conventions

1. Check project root for convention files: `agents.md`, `AGENTS.md`, `CLAUDE.md` (project-level only), `.cursorrules`, `GEMINI.md`, `copilot-instructions.md`.
2. **If index file found**: READ contents, extract all referenced file paths. Include in registry table.
3. For non-index files, record file directly.

### Step 3: Write Registry

```markdown
# Skill Registry

**Delegator use only.** Agents launching sub-agents read this registry to resolve compact rules, then inject them directly into sub-agent prompts.

## Lazy Load Protocol
Kernel handles top-level routing. Skills loaded on demand.

## Skills Index
Use `glob internal/assets/skills/**/SKILL.md` or `glob .agent/skills/**/SKILL.md` to discover skills dynamically.

## Auto-Installed Skills
| Skill | Source | Status | Path |
|-------|--------|--------|------|
| {name} | {source} | cached / not cached | {path} |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### {skill-name-1}
- Rule 1
- ...

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| {file} | {path} | |
```

### Step 4: Persist Registry

**MANDATORY — do NOT skip.** Follow persistence rules in Step 2 of `_shared/mode-branching.md`. Defaults to **hybrid** behavior.

### Step 5: Return Summary

```markdown
## Skill Registry Updated

**Project**: {project name}
**Location**: .atl/skill-registry.md
**Engram**: {saved / not available}

### Skills Processed
| Skill | Compact Rules Count |
|-------|---------------------|
| {name} | {count} |

### Project Conventions Found
| File | Path |
|------|------|
| {file} | {path} |

### Next Steps
Orchestrator reads this registry once per session. To discover skills dynamically, use glob.
```

## Rules

- ALWAYS write `.atl/skill-registry.md` regardless of SDD persistence mode
- ALWAYS save to engram if `mem_save` tool available
- Check `skills-lock.json` BEFORE scanning filesystem
- SKIP `_shared`, `_archived`, `skill-registry` directories when scanning
- Compact rules: 5-15 lines per skill
- If auto-installed skill not cached, note needs download
- Add `.atl/` to project's `.gitignore` if not listed
