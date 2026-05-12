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

Follow `_shared/mode-branching.md` for artifact-store branching (metadata only). This skill uses a fixed hybrid-style persistence for infrastructure availability.

- **Artifact Name**: .atl/skill-registry.md
- **Topic Key**: skill-registry
- **Type**: config

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

You generate or update the **skill registry** — a dynamic catalog of available skills with **compact rules** (pre-digested, 5-15 line summaries).

## Lazy Load Protocol

The project uses a lazy-load architecture:
1. Skills are discovered via `glob` dynamically at runtime.
2. The kernel's intent router handles top-level routing.
3. This skill-registry generates compact rules for injection into sub-agent prompts.
4. No static index is maintained in the kernel.

## When to Run

- After installing or removing skills
- After setting up a new project
- When the user explicitly asks to update the registry
- As part of `sdd-init` (it calls this same logic)

## What to Do

### Step 1: Scan Skills

1. **Auto-Installed Skills (skills-lock.json)**:
   - Check for `skills-lock.json` in the project root.
   - If found, parse it to get the list of auto-installed skill names.
   - For each auto-installed skill, check if it is already present in the filesystem (might be cached).
   - If cached, include it in the scan.
   - If not cached, note it as "available but not downloaded".
2. **Filesystem Skills (glob)**:
   - Glob `internal/assets/skills/**/SKILL.md` for project-level skills.
   - Glob `.agent/skills/**/SKILL.md` for workspace-level skills.
3. **SKIP** `_shared`, `_archived`, and `skill-registry`.
4. **Deduplicate**: Auto-installed skills > Project-level > Workspace-level.
5. For each skill found, read the **full SKILL.md** (if >200 lines, focus on frontmatter and Critical Patterns/Rules) to extract:
   - `name` field
   - `description` field
   - **Compact rules** (see Step 1b)
6. Do NOT build a static trigger table.

### Step 1b: Generate Compact Rules

For each skill found, generate a **compact rules block** (5-15 lines max) containing ONLY:
- Actionable rules and constraints ("do X", "never Y", "prefer Z over W")
- Key patterns with one-line examples where critical
- Breaking changes or gotchas that would cause bugs if missed

**DO NOT include**: purpose/motivation, when-to-use, full code examples, installation steps.

Format per skill:
```markdown
### {skill-name}
- Rule 1
- Rule 2
- ...
```

**The compact rules are the MOST IMPORTANT output of this skill.**

### Step 2: Scan Project Conventions

1. Check the project root for convention files: `agents.md`, `AGENTS.md`, `CLAUDE.md` (project-level only), `.cursorrules`, `GEMINI.md`, `copilot-instructions.md`.
2. **If an index file is found**: READ its contents and extract all referenced file paths. Include referenced paths in the registry table.
3. For non-index files, record the file directly.

### Step 3: Write the Registry

Build the registry markdown:

```markdown
# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts.

## Lazy Load Protocol
The kernel handles top-level routing. Skills are loaded on demand. 

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

### Step 4: Persist the Registry

**This step is MANDATORY — do NOT skip it.**
Follow the persistence rules defined in Step 2 of `_shared/mode-branching.md`. Note: This skill defaults to **hybrid** behavior.

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
The orchestrator reads this registry once per session. To discover skills dynamically, use glob.
```

## Rules

- ALWAYS write `.atl/skill-registry.md` regardless of any SDD persistence mode
- ALWAYS save to engram if the `mem_save` tool is available
- Check `skills-lock.json` BEFORE scanning filesystem.
- SKIP `_shared`, `_archived`, and `skill-registry` directories when scanning.
- Compact rules MUST be 5-15 lines per skill.
- If an auto-installed skill is not cached, note that it needs to be downloaded.
- Add `.atl/` to the project's `.gitignore` if not listed.
