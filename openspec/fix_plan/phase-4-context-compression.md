# Phase 4 — Context Compression & Sub-Agent Injection Optimization

> **Cognitive Mode**: +++Systemic +++Empirical +++Pragmatic  
> **CCLD Tag**: `[PHASE-4][CONTEXT][COMPRESSION][INJECTION]`  
> **Status**: BLOCKED until Phase 0 Task E + Phase 2 complete  
> **Estimated Duration**: 1–2 sessions  
> **Depends On**: `audit/context-saturation.md`, Phase 2 skill registry indexed format

---

## 4.1 Objective

Reduce per-sub-agent prompt token overhead by replacing full-file content injection with targeted compact rule injection. Eliminate redundant mandatory-skill blocks. Enforce context-guardian behavior mechanically rather than aspirationally.

**Target Outcome**: 40–70% reduction in mandatory-injection token overhead per sub-agent prompt. Context window budget redirected to task-specific content.

---

## 4.2 Root Cause (from Phase 0 Audit)

### 4.2.1 Full-Content Injection in Sub-Agent Template

Current General Orchestrator Sub-Agent Launch Template:
```
## Project Standards (auto-resolved)
{mandatory skills compact rules — ripgrep, bash-expert, notebooklm, context-guardian}
{task-matched skills compact rules}

## Research Procedure
1. FIRST: ...
2. If hit and age < 168h: ...
3. SECOND: Local ripgrep. Walk the repo. Persist key snippets.
...

## Available Tools
{verified tools from tool availability check}

## Shared Contract
{content from skills/_shared/general-phase-common.md}     ← FULL FILE CONTENT
```

**Problem**: `{content from skills/_shared/general-phase-common.md}` expands to the full `sdd-phase-common.md` content (~150–300 lines). This is injected into EVERY sub-agent, even when the sub-agent only needs 10% of it.

### 4.2.2 Redundant Research Procedure Block

The Research Procedure (5-step Engram-first → web fallback) is repeated verbatim in:
- `general-orchestrator.md` Sub-Agent Template
- `sdd-orchestrator.md` Sub-Agent Template  
- `research-routing.md` (shared fragment)

Each expansion is ~50–80 tokens. With 5–8 sub-agents per SDD workflow, this is 250–640 redundant research-routing tokens per session.

### 4.2.3 Mandatory Skills Always Injected Regardless of Task

"Mandatory Skills (ALWAYS injected)" includes:
- `ripgrep` — used by all. OK.
- `bash-expert` — used by most. OK.
- `context-guardian` — behavioral skill; full rules injected. Medium overhead.
- `mcp-notebooklm-orchestrator` — injected even when NotebookLM is unavailable.

**Problem**: `mcp-notebooklm-orchestrator` compact rules (30–50 tokens) injected into every sub-agent prompt even when NotebookLM probe returned `unavailable`.

### 4.2.4 No Context-Guardian Enforcement Mechanism

`context-guardian` is listed as a mandatory skill but there is no mechanical enforcement. The sub-agent is told to "detect context pressure" but no rule specifies what to do when pressure is detected (e.g., which content to drop first).

---

## 4.3 Refactoring Plan

### 4.3.1 Tiered Injection Protocol

Replace the single mandatory block with a three-tier injection system:

```markdown
## Sub-Agent Prompt Injection Protocol (Tiered)

### Tier 0 — Identity & Language (always inject, ~30 tokens)
{cognitive_posture_block}
Language Mandate: English only.
Caveman Output Compression: [3-line summary only].

### Tier 1 — Execution Fundamentals (always inject, ~80 tokens)
**ripgrep**: Use `rg` not `grep`. Pattern: `rg "query" --type go`. JSON: `rg --json`.
**bash-expert**: No interactive prompts. Quote all variables. Fail fast: `set -euo pipefail`.
**task-output**: Return envelope per Section D: { status, executive_summary, artifacts, risks }.

### Tier 2 — Tool-Conditional (inject only if tool available, ~20 tokens each)
IF engram=available:
  **engram**: mem_search before any research. topic_key: {assigned_key}. Save results.
IF notebooklm=available:
  **notebooklm**: Use only if Engram + ripgrep yield nothing. Mode 1/2 only.
IF context7=available:
  **context7**: Use for framework docs. resolve before searching web.

### Tier 3 — Task-Matched (inject only for matched workflow, ~40-100 tokens)
IF workflow=sdd-apply:
  **strict-tdd**: [compact rules from #skill-sdd-apply Quick Index]
IF workflow=research:
  **research-routing**: Engram→ripgrep→context7→notebooklm→web (escalate only on miss).
IF workflow=verify:
  **sdd-verify**: Tests first. Never modify tests to force pass.
(... per workflow type)
```

**Token savings estimate**:
- Current: ~350–500 tokens mandatory injection per sub-agent
- New: ~130–200 tokens (Tier 0+1) + ~40–80 conditional (Tier 2) + ~40–100 task-matched (Tier 3)
- Net: ~200–300 tokens saved per sub-agent × 6 sub-agents = 1,200–1,800 tokens per session

---

### 4.3.2 Compact Rules Extraction Procedure

The orchestrator extracts compact rules from the indexed skill registry (Phase 2 output):

```markdown
## Compact Rule Extraction (Orchestrator Internal)

Before building sub-agent prompt:
1. Load `.atl/skill-registry.md` Quick Index only (50–100 tokens).
2. For each skill in [mandatory + task-matched]:
   a. Find `### {skill-name} {#skill-{slug}}` anchor in registry.
   b. Extract text from anchor to next `###` heading.
   c. Trim to Compact Rules section only (strip Path, Trigger metadata).
3. Inject extracted compact rules (never the full SKILL.md path content).

This replaces `{content from skills/_shared/context-mode-routing-policy.md}` patterns.
```

---

### 4.3.3 Context-Guardian Enforcement: Drop Priority List

Add a mechanical enforcement list to the sub-agent template:

```markdown
### Context-Guardian Enforcement

If token budget is under pressure (Mode 2 or Mode 3 detected):
Drop content in this order (earlier = drop first):

| Priority | Content | Action |
|---|---|---|
| 1 | Research procedure steps | Drop steps 4-5 (NotebookLM, Web); keep 1-3 |
| 2 | Task-matched compact rules | Drop Tier 3 injection |
| 3 | Context7 tool rules | Drop if context7 result already in Engram |
| 4 | Example outputs / templates | Drop all examples |
| 5 | Detailed risk descriptions | Keep risk IDs only |
| NEVER DROP | Task description, file paths, error messages, code snippets | Critical for correctness |

Context-guardian MUST emit: `[CTX] Mode {1|2|3}. Dropped: {list of dropped items}.` at start of response when any content is dropped.
```

---

### 4.3.4 Remove Redundant Research Procedure from Sub-Agent Template

The Research Procedure is already in `research-routing.md`. The sub-agent template should reference it via Tier 3 injection, not embed it:

```markdown
## Research Procedure (REMOVE from Sub-Agent Template)
```

Replace with:
```markdown
## Research
Follow research-routing protocol (injected via Tier 3 when workflow=research/investigate).
For all other workflows: Engram-first, then ripgrep. No web unless workflow=research.
```

**Token savings**: ~60 tokens per non-research sub-agent × 4 non-research sub-agents per session = 240 tokens/session.

---

### 4.3.5 SDD Phase Protocol Progressive Loading — Strict Enforcement

The SDD Orchestrator says "Load protocol JUST BEFORE delegating that phase. Do NOT preload all protocols at session start." This is correct but not enforced. Add explicit enforcement:

```markdown
## Protocol Loading Guard

FORBIDDEN: Loading more than ONE phase protocol per orchestrator response.
FORBIDDEN: Loading sdd-apply.md before sdd-tasks is complete.
FORBIDDEN: Retaining a loaded protocol in orchestrator context after delegation is complete.

Enforcement:
- Load protocol → inject into sub-agent → delegate → DISCARD from orchestrator context.
- Next phase starts fresh: load only that phase's protocol.
- If context pressure detected: drop previously-loaded (now-used) protocols first.
```

---

## 4.4 Files to Create / Modify

| File | Action | Notes |
|---|---|---|
| `.agent/skills/_shared/general-orchestrator.md` | MODIFY | Replace Sub-Agent Template with Tiered Injection Protocol |
| `.agent/skills/_shared/sdd-orchestrator.md` | MODIFY | Replace Sub-Agent Template; add Protocol Loading Guard |
| `.agent/skills/_shared/sdd-phase-common.md` | MODIFY | Add Compact Rules extraction note; tag sections for Tier 3 |
| `.agent/skills/_shared/skill-resolver.md` | MODIFY | Document tiered injection; extraction procedure |
| `internal/assets/*/general-orchestrator.md` | MODIFY | Sync tiered template to all agent variants |
| `internal/assets/*/sdd-orchestrator.md` | MODIFY | Sync tiered template to all agent variants |

---

## 4.5 Token Budget Targets

| Injection Type | Current Tokens | Target Tokens | Reduction |
|---|---|---|---|
| Mandatory skills block | ~350–500 | ~130–200 | ~55% |
| Research procedure | ~60–80 | ~15 (reference only) | ~80% |
| Tool availability section | ~40–60 | ~20 (Tier 2 conditional) | ~60% |
| Shared contract (full file) | ~150–300 | ~50 (targeted section) | ~75% |
| **Total per sub-agent** | **~600–940** | **~215–285** | **~65%** |

For a full SDD workflow (6–8 sub-agents): saves ~2,400–5,000 tokens per session.

---

## 4.6 Acceptance Criteria

- [ ] Tiered Injection Protocol replaces monolithic mandatory block in all orchestrators
- [ ] Compact rules injected via Quick Index extraction (not `{content from ...}` expansion)
- [ ] `mcp-notebooklm-orchestrator` rules NOT injected when NotebookLM unavailable
- [ ] Context-Guardian drop priority list present; emits `[CTX] Mode X` when triggered
- [ ] Research Procedure removed from sub-agent template; replaced with 2-line reference
- [ ] Protocol Loading Guard present in `sdd-orchestrator.md`
- [ ] No regression in sub-agent quality — verify against SDD workflow scenario tests from Phase 3

---

## 4.7 Sub-Agent Delegation

```
[PHASE-4 ORCHESTRATOR]
    │
    ├── [4A] md-writer-agent     → general-orchestrator.md Tiered Injection Protocol
    ├── [4B] md-writer-agent     → sdd-orchestrator.md Tiered Template + Protocol Guard
    ├── [4C] md-writer-agent     → sdd-phase-common.md Compact Rules section tags
    ├── [4D] md-sync-agent       → sync to all internal/assets variants (depends 4A+4B)
    └── [4E] verify-agent        → token count verification before/after (depends 4D)
```

4A, 4B, 4C launch in parallel.  
4D launches after 4A + 4B.  
4E launches after 4D.
