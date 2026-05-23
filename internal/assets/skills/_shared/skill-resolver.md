# Skill Resolver — Universal Protocol

Any agent that **delegates work to sub-agents** MUST follow this protocol to resolve and inject relevant skills. Applies to ATL orchestrator, judgment-day, pr-review, and ANY future skill/workflow launching sub-agents.

## Why

Sub-agents are born with NO context about skills. Without injection, a judge reviewing Next.js won't know React 19 patterns, fix agent won't follow conventions, PR creator won't use project's PR template.

## When

Before EVERY sub-agent launch involving **reading, writing, or reviewing code**. Skip only for purely mechanical delegations (e.g., "run this test command").

## Protocol

### Step 1: Obtain Skill Registry (once per session)

Registry contains **Compact Rules** section with pre-digested rules per skill (5-15 lines each). Inject this — NOT full SKILL.md paths.

Resolution order:
1. Cached from earlier session? → use cache
2. `mem_search(query: "skill-registry", project: "{project}")` → `mem_get_observation(id)` for full content
3. Fallback: read `.atl/skill-registry.md` from project root if exists
4. No registry? → proceed without skills (warn user: "No skill registry found — sub-agents work without project-specific standards. Run `skill-registry` to fix.")

### Step 2: Match Relevant Skills

**Mandatory Skills (ALWAYS injected)**
Regardless of task matcher, always inject:
- `ripgrep` — pattern search (replaces grep)
- `bash-expert` — safe shell scripting
- `mcp-notebooklm-orchestrator` — optional research source
- `context-guardian` — context pressure detection

Match additional skills on TWO dimensions:

**A. Code Context** — what files will sub-agent touch or review?

Compact rules provided by resolver are **immutable** for session. Override via specific skill file at project root (`./SKILL.md`) or project skills directory (`.architect/skills/`).

Map file patterns to skills from registry (always defer to registry's Trigger field):
- `.tsx`, `.jsx` → react skills
- `.ts` → typescript skills
- `app/**`, `pages/**` → nextjs/angular/framework skills
- `.py` → python/django skills
- `.go` → go skills
- `*.test.*`, `*.spec.*` → testing skills
- Style files → tailwind/css skills

Use `Trigger` field in registry's User Skills table to match.

**B. Task Context** — what ACTIONS will sub-agent perform?

| Sub-agent action | Match skills with triggers mentioning... |
|-----------------|------------------------------------------|
| Create PR | "PR", "pull request" |
| Write/review code | Specific framework/language |
| Create Jira tickets | "Jira", "epic", "task" |
| Write Notion docs | "Notion", "RFC", "PRD" |
| Write comments | "comment" |
| Run tests | "test", "vitest", "pytest", "playwright" |

### Step 3: Inject into Sub-Agent Prompt

From registry's **Compact Rules** section, copy matching skill blocks directly into sub-agent's prompt.

```
## Project Standards (auto-resolved)

{paste compact rules blocks for each matching skill}
```

Goes BEFORE sub-agent's task-specific instructions. Standards loaded before work begins.

**Key**: inject COMPACT RULES text, not paths. Sub-agent should NOT read SKILL.md files — rules arrive pre-digested in prompt.

### Step 4: Include Project Conventions

If registry has **Project Conventions** section and sub-agent works on project code:

```
## Project Conventions
Read these files for project-specific patterns:
- {path1} — {notes}
- {path2} — {notes}
```

Conventions are short references (paths + notes), cheap to pass. Sub-agent reads only if relevant.

## Adaptive Routing Contract Integration

When `adaptive-reasoning` used before delegation, resolver SHOULD consume routing record directly.

Required fields: `owner`, `scope`, `ambiguity`, `dependency_shape`, `risk`, `verification_burden`, `cost_sensitivity`, `route`, `reason`.

Resolver behavior by `route`:
- `native-owner`: delegate to owner skill only when extra delegation still needed
- `deterministic-validators`: prioritize machine-checkable validation flow; do not substitute judge route
- `judgment-day`: launch adversarial review for defect discovery
- `autoreason-lite`: only for bounded proposal/spec/design comparison with incumbent + competitor
- `native-sdd-first`: route to SDD owner phase before narrower overlays

Routing record missing/incomplete: fallback to normal resolver matching, log warning.

## Token Budget

Compact rules add **50-150 tokens per skill** to sub-agent prompt. Typical delegation matching 3-4 skills = ~400-600 tokens — negligible vs. code sub-agent reads.

More than **5 skill blocks** match → keep only 5 most relevant (prioritize code context over task context).

## Compaction Safety

Protocol is compaction-safe:
- Registry lives in engram/filesystem, not orchestrator memory
- Each delegation re-reads registry if needed (Step 1 handles cache miss)
- Compact rules copied into each sub-agent prompt at launch — even if orchestrator forgets, sub-agents already have rules

## Feedback Loop

Sub-agents MUST report skill resolution status in return envelope:
- `injected` — received `## Project Standards (auto-resolved)` from orchestrator (ideal)
- `fallback-registry` — no standards received, self-loaded from skill registry
- `fallback-path` — loaded via `SKILL: Load` path
- `none` — no skills loaded

**Orchestrator self-correction**: if sub-agent reports anything other than `injected`, orchestrator MUST:
1. Re-read skill registry immediately (may be lost to compaction)
2. Ensure ALL subsequent delegations include `## Project Standards (auto-resolved)`
3. Log warning: "Skill cache miss detected — reloaded registry for future delegations."

Prevents silent degradation where orchestrator forgets skills after compaction.

## Integration Points

- **ATL Orchestrator**: follows protocol for ALL delegations (SDD and non-SDD)
- **judgment-day**: follows protocol before launching Judge A, Judge B, Fix Agent
- **pr-review**: has internal skill loading — should migrate to this protocol
- **Any future skill delegating**: MUST reference this protocol

## Context Assembly Integration

When invoking `context-guardian` to build context pack, map resolved Compact Rules directly into `working_rules:` section of context pack.
