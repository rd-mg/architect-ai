# Phase 3 — Orchestrator Routing Fork & Parallel Tool Probes

> **Cognitive Mode**: +++Systemic +++Forensic +++Divergent  
> **CCLD Tag**: `[PHASE-3][ORCHESTRATOR][ROUTING-FORK][PARALLEL-PROBES]`  
> **Status**: BLOCKED until Phase 0 Tasks C + F complete  
> **Estimated Duration**: 2 sessions  
> **Depends On**: `audit/orchestrator-routing.md`, `audit/sdd-phase-graph.md`

---

## 3.1 Objective

Fix three delegation inefficiencies in the orchestrator prompt layer:

1. **Double Session-Setup**: When intent is SDD, the General Orchestrator runs its own Tool Availability Check, then routes to the SDD Orchestrator which runs its own. Fix: early-exit fork that skips General Orchestrator setup for SDD intents.
2. **Sequential Tool Probes**: Both orchestrators probe Engram, NotebookLM, Context7 in serial. Fix: parallel probe dispatch in the same response.
3. **No Static Dispatch Table**: Routing is determined by free-text scanning at runtime with no precomputed decision. Fix: add a deterministic routing gate before any setup work.

**Target Outcome**: Reduce SDD session cold-start overhead by ~40–60%. Eliminate redundant tool probes.

---

## 3.2 Root Cause (from Phase 0 Audit)

### 3.2.1 Double Session-Setup on SDD Path

**Current flow**:
```
User: "use sdd for add-user-export"
  → General Orchestrator receives message
  → General Orchestrator: scan Routing Table → match /sdd-new
  → General Orchestrator: confirm intent with user
  → General Orchestrator: [Tool Availability Check] → mem_search × 3
  → General Orchestrator: routes to SDD Orchestrator
    → SDD Orchestrator: [Session-Setup Triplet]
      → Step 1: SDD Init Guard → mem_search
      → Step 2: Artifact Store Resolution → mem_search × 2
      → Step 3: Execution Mode → ask user
```

**Problem**: The General Orchestrator's Tool Availability Check fires 3 `mem_search` calls that the SDD Orchestrator will repeat in its own Step 2. Net redundancy: 3–5 extra `mem_search` RPC calls per SDD session start.

### 3.2.2 Sequential Tool Probes

In both orchestrators, the Tool Availability Check is written as:
```
1. Engram: mem_search(query: "tool-test", ...)
2. NotebookLM: mem_search(query: "notebooklm/") + notebooklm_list_notebooks()
3. Context7: presence of context7_resolve tool
4. Other MCPs: per-tool status
```

**Problem**: Steps 1, 2, 3 are independent. No instruction to parallelize them. LLM executes tool calls one at a time unless explicitly instructed to batch.

### 3.2.3 Aspirational vs Enforced Parallelism

The `## Parallel Delegation (MANDATORY)` section says:
```
"Multiple file explorations → launch N explorers"
"Running tests + static analysis → launch both test runners"
```

But the enforcement mechanism is only the word "MANDATORY" — there's no static dispatch table that the orchestrator can verify against. If the orchestrator "forgets" to parallelize (context pressure, intent ambiguity), there's no fallback gate.

---

## 3.3 Refactoring Plan

### 3.3.1 Early-Exit Routing Gate (General Orchestrator)

Add a **Router Gate** as the **first** section of `general-orchestrator.md`, executed before any tool calls:

```markdown
## ROUTER GATE (Execute FIRST — before any tool calls, before session setup)

Read the user's message. In ONE decision step, classify it:

| Classification | Criteria | Action |
|---|---|---|
| `SDD_INTENT` | Message matches SDD Pattern Table below | STOP. Do not run Tool Availability Check. Transfer directly to SDD Orchestrator with full user message. |
| `NON_SDD` | All other intents | Continue with General Orchestrator setup below. |

### SDD Pattern Table (fast-path — no LLM needed, pure string match)
- Contains: "use sdd", "start sdd", "begin sdd", "apply spec-driven", "sdd-new", "sdd-continue", "sdd-ff", "sdd-explore", "sdd-init", "sdd-verify", "sdd-archive", "sdd-onboard"
- Regex equivalent: `/\b(sdd|spec-driven|sdd-new|sdd-ff|sdd-continue)\b/i`

### On SDD_INTENT
→ Emit: `[Router] SDD intent detected. Forwarding to SDD Orchestrator.`
→ DO NOT run Tool Availability Check.
→ DO NOT run Session-Setup Triplet (SDD Orchestrator owns this).
→ IMMEDIATELY transfer to SDD Orchestrator skill with the original user message.

### On NON_SDD
→ Continue reading this document from "## Intent Resolution & Task Router".
```

**Impact**: Eliminates 3+ redundant `mem_search` calls on every SDD session start. Orchestrator overhead for SDD: ~0 (direct forward).

---

### 3.3.2 Parallel Tool Probe Dispatch

Replace the sequential Tool Availability Check with an explicit parallel dispatch instruction in **both** orchestrators:

```markdown
## Tool Availability Check (PARALLEL DISPATCH — all probes in ONE response)

Launch ALL of the following tool calls in the SAME response (parallel dispatch):

```tool-batch
[probe-1] mem_search(query: "tool-test", project: "{project}")
[probe-2] mem_search(query: "notebooklm/", project: "{project}")  
[probe-3] mem_search(query: "sdd-session/{project}/artifact-mode", project: "{project}")
[probe-4] (if context7_resolve is in tool list → mark available; otherwise → unavailable)
```

Wait for all results, then:
- probe-1 result: Engram = available if no error / unavailable if error
- probe-2 result: NotebookLM configured = available if hit
- probe-3 result: Artifact mode = use cached value if hit; otherwise ask user (SDD only)
- probe-4 result: Context7 = available if tool present

Record as: `tools = { engram: bool, notebooklm: bool, context7: bool }`
Cache to session memory (do not re-probe within same session).
```

**Impact**: 4 serial RPC calls → 1 parallel batch. Reduces cold-start latency.

---

### 3.3.3 SDD Parallel Dispatch Enforcement Table

Replace the aspirational "MANDATORY" language with a **static dispatch lookup table** in `sdd-orchestrator.md`:

```markdown
## Parallel Dispatch Table (STATIC — check before delegating)

Before delegating any phase, look up the phase in this table.
If `Parallelizable=YES`, you MUST emit ALL task tool calls in the same response.

| Phase | Parallelizable | Condition | Parallel Scope |
|---|---|---|---|
| sdd-explore | YES | Multiple topics or modules | One agent per topic/module |
| sdd-spec | YES | Multiple unrelated features | One agent per feature |
| sdd-verify | YES | Tests AND static analysis | test-runner + linter in parallel |
| sdd-apply | CONDITIONAL | Tasks modifying different files | Group by target file set |
| sdd-propose | NO | Single coherent proposal | — |
| sdd-design | NO | Single architecture doc | — |
| sdd-tasks | NO | Depends on design output | — |
| sdd-archive | NO | Sequential: merge → move → commit | — |

### Enforcement Mechanism
After deciding to delegate a phase:
1. Look up phase in table above.
2. If `Parallelizable=YES`: count work items (topics, features, test types).
3. If count > 1: MUST launch multiple task calls in same response. Verify by counting your tool calls — if count == 1 for a parallelizable phase, PAUSE and split.
4. If `Parallelizable=CONDITIONAL`: check if target files overlap. If no overlap → parallel. If overlap → sequential.

NEVER emit a single task call for a YES-parallelizable phase with multiple work items.
```

---

### 3.3.4 Session State Cache Protocol (Deduplication)

To prevent the General and SDD Orchestrators from independently re-probing tools within the same session, introduce a shared session state cache protocol:

```markdown
## Session State Cache (Both Orchestrators)

At session start, check:
```
mem_search(query: "session-state/{project}/tools", project: "{project}")
```

If hit AND age < 30min → USE cached tool availability. Skip all probes.
If miss OR stale → Run parallel probe batch (§ Tool Availability Check). 
After probe → save:
```
mem_save(
  title: "session-state/{project}/tools",
  topic_key: "session-state/{project}/tools",
  type: "session-cache",
  project: "{project}",
  content: JSON({ engram, notebooklm, context7, timestamp })
)
```

When General Orchestrator forwards to SDD Orchestrator → pass tool state in handoff context:
```
## Forwarded Session State
- Tools: {engram: true, notebooklm: false, context7: true}
- Artifact Mode: [if already resolved]
- Exec Mode: [if already resolved]
```

SDD Orchestrator MUST check for `## Forwarded Session State` before running its own probes.
```

---

### 3.3.5 Non-SDD Routing Precision Fix

The General Orchestrator's Routing Table currently uses vague phrase matching. Add specificity to reduce false positives:

```markdown
## Intent Resolution — Updated Routing Table

Priority order: later rows take precedence if multiple match.

| User phrase (match any) | Workflow | Target Agent | Confidence |
|---|---|---|---|
| "fix", "why", "crash", "error", "broken", "not working" | `/solve` | Solver | High |
| "debug", "trace", "step through", "breakpoint" | `/debug` | Solver | High |
| "ideas", "brainstorm", "options", "alternatives", "what if" | `/brainstorm` | Ideator | High |
| "research", "how does", "explain", "investigate", "understand" | `/investigate` | Researcher | High |
| "build", "prototype", "quick", "draft", "make" | `/prototype` | Generalist | Medium |
| (no match above) | (inline) | Generalist | Low |

### Disambiguation Rule
If match confidence == Medium AND message also matches another pattern → ask:
`"Did you mean to [/solve] or [/prototype]?"`
Never guess. Never default to Generalist silently when a high-confidence match exists.
```

---

## 3.4 Files to Create / Modify

### Prompt Layer

| File | Action | Notes |
|---|---|---|
| `.agent/skills/_shared/general-orchestrator.md` | MODIFY | Add Router Gate section at top; parallel tool probes; session cache protocol |
| `.agent/skills/_shared/sdd-orchestrator.md` | MODIFY | Add Parallel Dispatch Table; session cache deduplication; accept forwarded state |
| `internal/assets/claude/general-orchestrator.md` | MODIFY | Mirror same changes for Claude target |
| `internal/assets/gemini/general-orchestrator.md` | MODIFY | Mirror for Gemini CLI |
| `internal/assets/cursor/general-orchestrator.md` | MODIFY | Mirror for Cursor |
| `internal/assets/codex/general-orchestrator.md` | MODIFY | Mirror for Codex |
| `internal/assets/gga/general-orchestrator.md` | MODIFY | Mirror for GGA |
| `internal/assets/kiro/general-orchestrator.md` | MODIFY | Mirror for Kiro |
| `internal/assets/generic/general-orchestrator.md` | MODIFY | Mirror for Generic |
| `internal/assets/antigravity/general-orchestrator.md` | MODIFY | Mirror for Antigravity |
| *(All sdd-orchestrator.md variants above)* | MODIFY | Same SDD changes |

### Go Layer (Sync Asset Write)

| File | Action | Notes |
|---|---|---|
| `internal/cli/sync.go` (or equivalent) | MODIFY | `RunSync` must propagate new orchestrator files to all agent targets |

---

## 3.5 Sync Strategy

The `internal/assets/` directory contains per-agent variants of every orchestrator file. The refactoring must update ALL variants consistently. Use the `RunSync` pipeline:

1. Modify canonical `.agent/skills/_shared/` files.
2. Run `architect-ai sync` — verify sync propagates to all agent target directories.
3. Verify per-agent variants are updated via:
   ```
   rg "ROUTER GATE" internal/assets/*/general-orchestrator.md
   ```

---

## 3.6 Testing Requirements

Since orchestrator files are markdown prompts, testing is behavioral (cannot be unit-tested in Go). Define acceptance tests as reproducible scenarios:

| Scenario | Input | Expected Behavior | Expected Anti-Pattern (must NOT occur) |
|---|---|---|---|
| SDD cold start | "use sdd for add-user-export" | Router Gate fires; direct forward to SDD Orchestrator; no General tool probes | General Orchestrator runs Tool Check before forwarding |
| SDD continue | "continue the add-user-export change" | SDD Orchestrator checks forwarded session state; no reprobing | SDD Orchestrator runs fresh probe if state is < 30min old |
| Multi-module explore | "explore the payment module and the stock module" | Two sdd-explore agents launched in same response | Single sdd-explore agent launched sequentially |
| Verify phase | "/sdd-verify" | test-runner agent + linter agent launched in same response | Sequential: test runs, then linter |
| Non-SDD debug | "why is my server crashing" | Routes to Solver with +++Forensic posture | Routes to Generalist |

---

## 3.7 Acceptance Criteria

- [ ] Router Gate section is first in `general-orchestrator.md` — before any tool calls
- [ ] On SDD intent: General Orchestrator emits 0 tool calls; forwards immediately
- [ ] Tool Availability Check dispatches all probes in ONE response (parallel)
- [ ] SDD Orchestrator checks for forwarded session state before reprobing
- [ ] Parallel Dispatch Table present in `sdd-orchestrator.md` with all 8 phases listed
- [ ] sdd-explore with 2 topics → 2 task calls in same response (verified in scenario test)
- [ ] All `internal/assets/*/general-orchestrator.md` variants contain Router Gate section
- [ ] `architect-ai sync` propagates changes to all agent targets

---

## 3.8 Sub-Agent Delegation

```
[PHASE-3 ORCHESTRATOR]
    │
    ├── [3A] md-writer-agent     → general-orchestrator.md Router Gate + parallel probes + session cache
    ├── [3B] md-writer-agent     → sdd-orchestrator.md Parallel Dispatch Table + session dedup
    ├── [3C] md-sync-agent       → sync Router Gate to all internal/assets/*/general-orchestrator.md (depends 3A)
    ├── [3D] md-sync-agent       → sync SDD changes to all internal/assets/*/sdd-orchestrator.md (depends 3B)
    └── [3E] verify-agent        → run scenario tests for Router Gate behavior (depends 3C + 3D)
```

3A, 3B launch in parallel.  
3C launches after 3A; 3D launches after 3B (parallel).  
3E launches after 3C + 3D.
