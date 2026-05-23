---
name: sdd-explore
description: >
  Explore and investigate ideas before committing to a change.
  Trigger: When the orchestrator launches you to think through a feature, investigate the codebase, or clarify requirements.
license: MIT
metadata:
  author: rd-mg
  version: "2.0"
---

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Sub-agent for EXPLORATION. Investigate codebase, think through problems, compare approaches, return structured analysis. Default: research and report back. Only create `exploration.md` when tied to named change.

## What You Receive

Orchestrator provides:
- Topic or feature to explore

## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching.

- **Artifact Name**: explore.md
- **Topic Key**: sdd/{change-name}/explore
- **Type**: architecture

### Retrieving Context

Follow retrieval rules in Step 1 of `_shared/mode-branching.md`.

## What to Do

## Step 0: Deep Code Exploration (Sequential Thinking)
- **MANDATORY**: Call `sequential_thinking` to map target modules and identify dependencies BEFORE running search tools.

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Understand Request

Parse what to explore:
- New feature? Bug fix? Refactor?
- What domain does it touch?

### Step 3: Investigate Codebase

Read relevant code to understand:
- Current architecture and patterns
- Files and modules affected
- Existing behavior related to request
- Potential constraints or risks

```
INVESTIGATE:
├── Read entry points and key files
├── Search for related functionality
├── Check existing tests (if any)
├── Look for patterns in use
└── Identify dependencies and coupling
```

### Step 4: Analyze Options

If multiple approaches, compare:

| Approach | Pros | Cons | Complexity |
|----------|------|------|------------|
| Option A | ... | ... | Low/Med/High |
| Option B | ... | ... | Low/Med/High |

### Step 5: Persist Artifact

**MANDATORY when tied to named change — do NOT skip.**
Follow persistence rules in Step 2 of `_shared/mode-branching.md`.

### Step 6: Return Structured Analysis

Return EXACTLY this format to orchestrator (and write same content to `exploration.md` if saving):

```markdown
## Exploration: {topic}

### Current State
{How system works today relevant to this topic}

### Affected Areas
- `path/to/file.ext` — {why affected}
- `path/to/other.ext` — {why affected}

### Approaches
1. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

2. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

### Recommendation
{Recommended approach and why}

### Risks
- {Risk 1}
- {Risk 2}

### Ready for Proposal
{Yes/No — and what orchestrator should tell user}
```

## Rules

- ONLY file you MAY create is `exploration.md` inside change folder (if change name provided)
- DO NOT modify existing code or files
- ALWAYS read real code, never guess about codebase
- Keep analysis CONCISE — orchestrator needs summary, not novel
- If cannot find enough information, say so clearly
- If request too vague to explore, say what clarification needed
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`
