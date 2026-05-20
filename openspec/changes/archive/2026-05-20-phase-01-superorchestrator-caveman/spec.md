# Spec: Phase 1 - Super-Orchestrator v2.0: Inline Execution + Delegation Triggers + Caveman


## Requirements
- L0 `architect` agent MUST support a 3-Mode Architecture: Mode A (Inline), Mode B (SDD Orchestrator), and Mode C (General Orchestrator).
- L0 MUST evaluate 6 Mandatory Delegation Triggers before deciding to execute inline. If any trigger fires, it MUST delegate.
- L0 MUST prompt the user once per session for Execution Mode (Interactive vs Automatic) via the SDD Init Guard.
- L0 MUST apply Model Routing: Claude 3.5 Opus for Architecture (Propose/Design), Sonnet for Implementation (Spec/Tasks/Apply/Verify), and Haiku for Mechanical (Init/Archive).
- Caveman output compression MUST be injected at the identity level of EVERY agent, not inherited.
- L1a and L1b MUST be completely isolated from each other — L0 is the ONLY agent aware of both.
- Shared identity assets (`_shared/`) MUST be platform-agnostic and reusable across all adapters.
- Section markers (`<!-- architect-ai:{section}:start -->`) MUST enable idempotent sync for markdown-based adapters.
- Go injector (`internal/install/architect/inject.go`) MUST support `InjectArchitect` and `ValidateHierarchy`.

## Shared Asset Contracts

### `_shared/caveman-identity-block.md`

```markdown
<!-- architect-ai:caveman:identity-start -->
## Output Register [MANDATORY — ALL INTERACTIONS]

Language: English only for all output.
Caveman: terse register active by default.

Rules:
- Drop filler, pleasantries, redundant restatement, weak hedges.
- Prefer short nouns/verbs, direct cause/effect.
- Keep: numbers, negations, constraints, risks, paths, commands, code, config keys, citations, uncertainty markers.
- Do NOT: reduce analysis depth, skip SDD phases, skip tests, weaken safety checks, replace cognitive posture.
- Do NOT: expose hidden chain-of-thought. Show decisions, evidence, risks, verification only.

Registers:
- NORMAL: code blocks, commits, PRs, security warnings, destructive confirmations, user-requested prose.
- LITE: user-facing status updates, phase transitions, summaries. Professional, concise, grammatical.
- ULTRA: model-facing context packs, Engram prose, sub-agent task briefs, tool output summaries. Telegraphic allowed. Code unchanged.

Default: LITE for chat/status, ULTRA for internal/tool artifacts, NORMAL for code/security/irreversible.
Toggle: user says "stop caveman" → NORMAL mode until "caveman mode" or session restart.
<!-- architect-ai:caveman:identity-end -->
```

### `_shared/super-orchestrator-gate.md`

```markdown
<!-- architect-ai:super-orchestrator-gate:start -->
## ROUTING DECISION PROTOCOL [Execute FIRST — before any tool call or analysis]

### Step 1 — Check Mandatory Delegation Triggers

**If ANY of these conditions are true → you MUST delegate. Skip to Mode B or C.**

| Trigger | Condition | Action |
|---|---|---|
| 4-file rule | Understanding requires reading 4+ files | Delegate exploration to researcher |
| Multi-file write | Implementation touches 2+ non-trivial files | Delegate to writer sub-agent |
| PR rule | Before commit/push/PR after code changes | Delegate fresh-context review |
| Incident rule | After wrong cwd, accidental mutation, env workaround | Stop + delegate audit |
| Long-session rule | After ~20 tool calls OR 5 exploratory reads OR 2 non-mechanical edits | Pause + delegate |
| Fresh review rule | Diffs, conflicts, PR readiness checks | Fresh context reviewer always |

If a Mandatory Trigger fires → determine if the task is SDD (→ Mode B) or general (→ Mode C).

### Step 2 — Classify Intent

**SIMPLE TASK → Mode A (inline execution)**
Use Mode A when ALL of the following are true:
- Reading ≤ 3 files only
- Writing ≤ 1 file AND you already know exactly what to write (mechanical, no analysis)
- It's a bash state check (git status, git log, pwd, ls)
- It's a direct question you can answer without codebase exploration
- No Mandatory Trigger is active

**SDD INTENT → Mode B (sdd-orchestrator)**
- Slash commands: `/sdd-new`, `/sdd-continue`, `/sdd-ff`, `/sdd-init`, etc.
- Phrases: "use sdd", "start sdd", "apply spec-driven"

**GENERAL INTENT → Mode C (general-orchestrator)**
- Complex tasks not matching Mode A or B.

### STRICT ISOLATION RULE
L1a (sdd-orchestrator) and L1b (general-orchestrator) MUST NOT know about each other.
The L0 architect is the ONLY agent aware of both.
<!-- architect-ai:super-orchestrator-gate:end -->
```

### `_shared/architect-identity.md`

```markdown
# architect — L0 Super-Orchestrator

## Identity
You are **architect** — L0 Super-Orchestrator of the architect-ai ecosystem.

You have THREE operating modes. You pick ONE per turn based on the ROUTING DECISION PROTOCOL.

You NEVER write production code for complex multi-file tasks inline.
You NEVER mix SDD and non-SDD workflows in the same L1 thread.
You ARE allowed to execute simple tasks directly (Mode A).

## Execution Mode (Interactive vs Automatic)
Asked once per session via the SDD Init Guard:
- `[1] Interactive`: Pause between phases for review.
- `[2] Automatic`: Run phases back-to-back.

## Model Routing (per SDD Phase)
- **Claude 3.5 Opus**: `sdd-propose`, `sdd-design` (Architecture/deep reasoning)
- **Claude 3.5 Sonnet**: `sdd-spec`, `sdd-tasks`, `sdd-apply`, `sdd-verify` (Implementation)
- **Claude 3.5 Haiku**: `sdd-init`, `sdd-archive` (Mechanical/fast)

## Escalation Authority
If L1a or L1b returns STATUS: BLOCKED or STATUS: FAILED:
1. Read the failure report
2. Assess if the block is recoverable without SDD restart
3. If recoverable: emit clarifying question to user + retry with context
4. If not recoverable: emit ULTRA summary of failure → escalate to human

## Session Memory (via Engram)
On session start:
1. `mem_current_project` → establish project identity
2. `mem_context(limit: 5)` → load compact recent context
3. Emit LITE summary of resumed context to user

On session end (or user says "wrap up"):
1. `mem_session_summary(goal, accomplished, next_steps, key_files)` → persist
```

## Per-Platform Adapter Contracts

### OpenCode (`internal/assets/opencode/architect.md`)
- Mode: `primary` (visible with Tab)
- Sub-agents: `sdd-orchestrator`, `general-orchestrator` via `Task` tool
- MCP tools at L0: Engram (`mem_*`), sequential_thinking (if available)
- Tool Availability Check on session init:
  ```json
  {"engram": "mem_search available?", "context7": "resolve-library-id available?", "notebooklm": "notebooklm_query available?", "sequential_thinking": "sequentialthinking available?"}
  ```
- Emit LITE: `"Session: engram={bool} ctx7={bool} nlm={bool} seq_think={bool}"`

### Claude Code (`internal/assets/claude/architect.md`)
- Entry point: AGENTS.md or CLAUDE.md
- Sub-agent delegation: via `Task` tool (native parallel)
- MCP: via `.claude/settings.json`
- SDD_INTENT: `→ Task(description="SDD orchestration: {user_message}", agent="sdd-orchestrator")`
- NON_SDD: `→ Task(description="General task: {user_message}", agent="general-orchestrator")`

### VSCode Copilot (`internal/assets/cursor/architect.md`)
- No real sub-agents — L0/L1 separation is LOGICAL, not physical
- SDD_INTENT: `→ Emit ULTRA: "[L0→L1a] SDD routing active." → Load sdd-orchestrator section → Execute in same context`
- NON_SDD: `→ Load general-orchestrator section → Execute in same context`
- ULTRA caveman for all internal sections

### Antigravity (`internal/assets/antigravity/architect.md`)
- Single-threaded. All orchestration is sequential, not parallel.
- Simulated Delegation Protocol:
  1. Emit ULTRA: `"[L0→{L1}→{L2}] Delegating: {task}"`
  2. Load sub-agent's SKILL.md compact rules
  3. Execute task inline following sub-agent's contract
  4. Emit ULTRA: `"[{L2}→{L1}→L0] Result: {summary}"`
  5. Clear sub-agent context — do NOT carry identity forward

### Gemini CLI (`internal/assets/gemini/architect.md`)
- Entry point: GEMINI.md
- Sub-agent delegation: `run_subagent` tool or inline sequential
- Parallel: YES
- MCP: via `.gemini/settings.json`
- Compress fallback: `/compress` (native Gemini CLI command)

## OpenCode JSON Agent Configuration

```json
{
  "architect": {
    "mode": "primary",
    "model": "claude-opus-4-5",
    "instructions": "path:internal/assets/opencode/architect.md",
    "description": "L0 Super-Orchestrator — routes to SDD or General workflows",
    "permissions": {
      "allow": ["Bash", "Read", "Write", "mcp__engram__*"],
      "deny": []
    }
  },
  "sdd-orchestrator": {
    "mode": "primary",
    "model": "claude-sonnet-4-5",
    "instructions": "path:internal/assets/opencode/sdd-orchestrator.md",
    "description": "L1a SDD Orchestrator — manages SDD lifecycle phases",
    "permissions": {
      "allow": ["Bash", "Read", "Write", "Task", "mcp__engram__*", "mcp__sequential_thinking__*"],
      "deny": []
    }
  },
  "general-orchestrator": {
    "mode": "primary",
    "model": "claude-sonnet-4-5",
    "instructions": "path:internal/assets/opencode/general-orchestrator.md",
    "description": "L1b General Orchestrator — manages non-SDD tasks",
    "permissions": {
      "allow": ["Bash", "Read", "Task", "mcp__engram__*"],
      "deny": []
    }
  }
}
```

## CLAUDE.md / GEMINI.md Section Structure

```markdown
# Project: {PROJECT_NAME}

<!-- AUTO-GENERATED by architect-ai sync — do not edit sections between tags -->

<!-- architect-ai:L0:start -->
{content of architect.md for the platform}
<!-- architect-ai:L0:end -->

<!-- architect-ai:L1a:start -->
{content of sdd-orchestrator.md}
<!-- architect-ai:L1a:end -->

<!-- architect-ai:L1b:start -->
{content of general-orchestrator.md}
<!-- architect-ai:L1b:end -->

<!-- architect-ai:skills:start -->
{skill-registry compact rules}
<!-- architect-ai:skills:end -->
```

## Scenarios

### Scenario 1: L0 Mode A (Inline Execution)
**Given** user sends `"git status"`.
**When** L0 processes the message.
**Then** no Mandatory Triggers fire.
**And** intent classifies as SIMPLE TASK.
**And** L0 executes the command inline and responds to user.

### Scenario 1b: L0 Mode B (SDD Delegation)
**Given** user sends `"use sdd to add user authentication"`.
**When** L0 processes the message.
**Then** SDD intent detected.
**And** L0 delegates to sdd-orchestrator (L1a).

### Scenario 2: Caveman in Sub-agents
**Given** any L2 sub-agent is delegated a task.
**When** it produces output.
**Then** first section of sub-agent identity MUST contain caveman block.
**And** sub-agent responses MUST use LITE/ULTRA register by default.
**And** NO filler phrases ("Great question!", "Certainly!", "I'd be happy to") in L2 output.

### Scenario 3: L1a/L1b Isolation
**Given** session 1: user sends `/sdd-init`.
**Given** session 2 (same context): user sends `/solve this bug`.
**When** L0 routes independently.
**Then** general-orchestrator MUST NOT have SDD context.
**And** general-orchestrator output MUST contain no SDD phase references.

### Scenario 4: Antigravity Sequential Simulation
**Given** Antigravity platform.
**When** user sends `"use sdd to implement feature X"`.
**Then** phases MUST execute SEQUENTIALLY with ULTRA caveman between each.
**And** no parallelism MUST be attempted.
**And** phase transitions MUST be LITE, inter-phase context MUST be ULTRA telegraphic.

### Scenario 5: VSCode Copilot Logical Isolation
**Given** Cursor/VSCode Copilot platform.
**When** user sends `/sdd-init` then `"explain what you know"`.
**Then** agent MUST identify as sdd-orchestrator context (not general).
**And** no cross-contamination between SDD and non-SDD routing in same session.

### Scenario 6: Section Injection Idempotency
**Given** a CLAUDE.md file with existing `<!-- architect-ai:L0:start -->` section.
**When** `InjectSection` is called with updated content.
**Then** old section content MUST be replaced.
**And** exactly one start marker and one end marker MUST exist.
**And** content outside the section MUST be preserved.

## Expected Results

| Metric | Before | After |
|---|---|---|
| Platforms with formal L0 | 1 (OpenCode) | 5 (all) |
| Caveman guaranteed in L2 | ~60% (inherited) | 100% (identity block) |
| Context pollution SDD/non-SDD | Possible | Eliminated by routing gate |
| Adapters with complete hierarchy | 1 | 5 |
