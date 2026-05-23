# Skill Resolver — Universal Protocol v3.0

Any agent that **delegates work to sub-agents** MUST follow this protocol to resolve and inject relevant skills. This applies to the ATL orchestrator, judgment-day, pr-review, and ANY future skill or workflow that launches sub-agents.

## Why This Exists

Sub-agents are born with NO context about what skills exist. Without skill injection, a judge reviewing a Next.js project won't know React 19 patterns, a fix agent won't follow project conventions, and a PR creator won't use the project's PR template.

## TIERED INJECTION PROTOCOL [MANDATORY]

The V3.0 architecture organizes all skills into 3 tiers to optimize context budget:

```
+-----------------------------------------------------------------+
| Tier 1: Foundation (Always Merged into _generated/foundation.md) |
|   - Injected into ALL sub-agent prompts via a single file read  |
+-----------------------------------------------------------------+
                               |
                               v
+-----------------------------------------------------------------+
| Tier 2: Context-Activated (Injected dynamically)               |
|   - File Diff Match (e.g. *.go -> go-testing)                   |
|   - Task Keyword Match (e.g. "PR" -> branch-pr)                  |
|   - Max 3 Tier 2 skills injected per task                       |
+-----------------------------------------------------------------+
                               |
                               v
+-----------------------------------------------------------------+
| Tier 3: On-Demand (Never injected automatically)                |
|   - Explicitly requested by name (e.g. researcher, solver)     |
+-----------------------------------------------------------------+
```

---

## When to Apply

Before EVERY sub-agent launch that involves **reading, writing, or reviewing code**. Skip only for purely mechanical delegations (e.g., "run this test command").

---

## The Protocol

### Step 1: Load Tier 1 (Foundation Block)

Every sub-agent prompt MUST start with the merged foundation standards block:
1. Read `.atl/_generated/foundation.md`.
2. Inject it at the very top of the system prompt under `## Project Foundation Standards`.
3. If the file is missing, the generator MUST be run first (`skill-registry --refresh`).

### Step 2: Match Tier 2 Skills (Context-Activated)

Match dynamically on TWO dimensions:

#### A. File Diff Match (what files will the sub-agent touch?)
- `.go` → go-testing
- `__manifest__.py` or Odoo directories → odoo-development-skill
- `go.mod` → go-testing

#### B. Task Keyword Match (what actions will the sub-agent perform?)
- "PR", "pull request", "git push" → branch-pr
- "commit", "apply", "sdd-apply" → work-unit-commits
- "issue", "bug report", "Jira" → issue-creation

**Max limit**: Inject a maximum of 3 Tier 2 skills to prevent context bloat. If more match, prioritize by the match type:
1. File Diff Match (highest priority)
2. Task Keyword Match

### Step 3: Tier 3 (On-Demand) Routing

Tier 3 skills (e.g. `researcher`, `solver`, `ideator`, `generalist`, `skill-creator`, `mcp-notebooklm-orchestrator`) are NEVER injected automatically. They are only invoked when the orchestrator explicitly routes a task to them.

---

## Token Budget

- **Tier 1 (Foundation)**: ~400-600 tokens (merged block of 6 core skills).
- **Tier 2 (Context-Activated)**: Max 3 skills (~150 tokens each) → ~450 tokens.
- **Total Skill Overhead**: ~1000 tokens. This is highly optimized and fits comfortably in the context window.

---

## Feedback Loop

Sub-agents MUST report their skill resolution status in their return envelope:
- `injected` — received `## Project Standards (auto-resolved)` from the orchestrator (ideal path)
- `fallback-registry` — no standards received, self-loaded from skill registry
- `none` — no skills loaded at all

**Orchestrator self-correction rule**: if a sub-agent reports anything other than `injected`, the orchestrator MUST:
1. Re-read the skill registry immediately (it may have been lost to compaction)
2. Ensure ALL subsequent delegations include `## Project Standards (auto-resolved)`
3. Log a warning to the user: "Skill cache miss detected — reloaded registry for future delegations."
