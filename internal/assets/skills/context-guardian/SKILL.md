---
name: context-guardian
description: >
  Context assembly contract and automated compaction hook. Builds a working
  context pack from active artifacts, pinned rules, semantic memory, and
  compacted tail. Triggers AUTOMATICALLY when token usage exceeds 50% of the
  context window, when compaction is detected, or when skill resolution falls
  back (indicating context loss). Persists context packs to Engram for
  recovery.
license: MIT
metadata:
  author: rd-mg
  version: "3.0"
---

# Context Guardian v3.0

## Purpose

Adaptive Reasoning gate: State Mode: {n} as first line of response per gate instructions.

Policy surface that assembles working context set from existing sources.
Makes curation decisions explainable. Filters, masks, and prioritizes facts
by bounded disclosure and semantic priority — does NOT dump full monolithic history.

In v3.0, orchestrator invokes this skill AUTOMATICALLY under six conditions (see below).
No longer a passive skill requiring explicit call.

## Cognitive Posture

Always operates under **+++Forensic**:
- Trace evidence chains
- Every claim needs provenance
- Never assume — verify
- Mark validation state per fact

## Auto-Trigger Rules (Orchestrator evaluates BEFORE each delegation)

Orchestrator MUST invoke when ANY condition holds:

1. `char_count(context_history) >= 100_000` → invoke
2. Sub-agent `skill_resolution: none` in last 2 turns → invoke
3. D4 >= 2 in current reasoning evaluation → invoke
4. 3+ file reads in same context window without compaction → invoke
5. User says "compact", "reset context", "what's my state" → invoke
6. `attempt_count >= 2` for current phase → invoke (context may be corrupted)

## Compaction Strategy

### Compaction Protocol (execute inline — do NOT delegate):

Goal: Durable cross-session memory. Zero information loss. LCM-compliant.

1. Extract `protected_facts`:
   - `active_change_name`, `artifact_store_mode`, `phases_completed[]`
   - `current_phase`, `attempt_count`, `caveman_firewall_active: true`
2. Summarize to: `{active_task, last_decision, critical_blockers[]}`
3. `mem_save(topic_key: "sdd/{change}/context-pack/{ts}", content: {pack})`
4. Emit: `[COMPACTION] Context reduced. Protected facts preserved.`
5. Continue with compressed context.

*If Branch A fails (Engram down), WARN and attempt Branch B.*

### Branch B: context-mode MCP Buffer (SECONDARY)

Goal: Transparent session buffering for large outputs. Session only.
Use `ctx_execute()` or `ctx_batch_execute()` for command output > 10KB
(e.g., large git logs, unfiltered ripgrep, large test suites).
Do NOT use for architectural decisions.

### Manual Summary Protocol (VSCode Copilot / Antigravity)

When no compress command available AND Engram checkpoint saved:
```
Emit LITE:
"Context limit approaching. I've saved a checkpoint to Engram.
Start a new chat and say: 'resume {change_name} from Engram'
I'll reload the checkpoint and continue from: {next_action}"
```

### Branch A+B Combined Protocol (recommended)

```
context-mode (Branch B) acts as session buffer throughout the session.
Branch A (Engram) fires when token_usage > 50% or D4 >= 2.

Timeline:
  [session start]    → Branch A: mem_context() to load prior state
  [during session]   → Branch B: ctx_execute() for large outputs
  [50% threshold]    → Branch A: mem_save() checkpoint + /compact
  [after compress]   → Branch A: mem_context() to reload
  [session end]      → Branch A: mem_session_summary()
```

### Platform Compress Command Matrix

| Platform | Native compress | Command | Branch A first? | Fallback |
|---|---|---|---|---|
| OpenCode | ✅ | `/compact` | YES | Branch B buffer |
| Claude Code | ✅ | `/compact` | YES | Branch B buffer |
| Gemini CLI | ✅ | `/compress` | YES | Branch B buffer |
| VSCode Copilot | ❌ | — | YES (Engram only) | Manual summary → new chat |
| Antigravity | ❌ | — | YES (Engram only) | Manual summary → new chat |

On trigger, orchestrator:
1. Reads this skill
2. Assembles Context Pack per procedure below
3. Persists pack to Engram (see Persistence)
4. Seeds next delegation with pack, discarding raw history above lineage cutoff

## Input Priority Order

1. **Active artifacts** (current specs, tasks, apply-progress)
2. **Pinned working rules** from skill registry (compact rules)
3. **Semantic memory observations** (Engram via `mem_search`)
4. **Compacted tail** of current session (last ~3 turns, full fidelity)

Earlier history represented only through masked evidence and protected facts — not verbatim.

## The Context Pack

Output MUST be markdown artifact with these exact stable sections:

```markdown
# Context Pack — {change-name or session-id}
Generated: {ISO-8601 timestamp}
Token count (estimated): {number}

## goal
{The current bounded objective — one sentence}

## active_tasks
{Tasks not yet closed, with status}
- [ ] 1.1 Implement X — status: in-progress
- [x] 1.2 Write tests for Y — status: done

## protected_facts
{Crucial history that survives compaction — see "Protected Classes"}
- [provenance: mem/12] [valid] Architecture decision: auth uses JWT
- [provenance: specs/auth-2026] [valid] Constraint: must support SAML fallback

## active_constraints
{Rules and architectural limits in force}
- [provenance: CAUTION_POLICY.md] [valid] No breaking changes to public API
- [provenance: manifest.py] [valid] Odoo 18.0 target version

## working_rules
{Compact rules from registry, keyed by skill name}
### sdd-apply
- ALWAYS read specs before implementing
- NEVER implement tasks not assigned to you
### branch-pr
- Every PR MUST link an approved issue
- Exactly one type label

## masked_evidence
{Compressed references to verbose outputs — NOT the outputs themselves}
- [provenance: cmd/run-1234] Test run output (250 lines) — all pass
- [provenance: logs/deploy-5678] Deployment logs — no errors
- [provenance: mem/45] Discovery: ORM query optimization findings (full content in Engram)

## suppressed_items
{What was intentionally left out and why}
- Raw stack traces from failed runs (not relevant after fix)
- Intermediate tool call results (superseded)
- Exploratory prompts that led nowhere

## lineage
{Pointers back to sources and compacted artifacts}
- Session ID: {id}
- Engram topic keys referenced: sdd/auth-2026/proposal, sdd/auth-2026/design
- Compacted history cutoff: {timestamp before which history is suppressed}
- Previous pack: {topic_key of prior pack, if any}
```

## Protected Classes

During any compaction, these classes MUST survive and move into
`protected_facts` or `active_constraints`:

- **Architecture decisions**: Permanent changes to software structure
- **Active constraints**: Rules, budgets, boundaries currently restricting execution
- **Open tasks**: Anything assigned but not checked off in `tasks.md`
- **Failing-test lineage**: Traceability of defect/failing test through fix attempts
- **Security-relevant findings**: Any finding marked CRITICAL in verification report
- **User commitments**: Promises made to user that must be kept

**protected_facts MUST always include:**
- `caveman_firewall_active: true` (NORMAL register mandatory for code artifacts)
- Current change name, phase, attempt_count
- `artifact_store_mode`, `phases_completed[]`

Cannot be dropped by low-priority compaction.

## Provenance and Validation State

Every retained/reused fact MUST carry explicit provenance and validation state.
Do not blindly copy text without marking reliability.

Validation states:
- `valid` — still consistent with current state
- `stale` — source exists but no longer reflects current reality
- `unverified` — reused provisionally, requires confirmation

Format: `[provenance: {source-ref}] [{state}] {fact}`

Examples:
- `[provenance: mem/12] [valid] Architecture: auth uses JWT`
- `[provenance: openspec/auth/spec.md#L45] [stale] Old JWT expiry was 1h`
- `[provenance: cmd/run-1234] [unverified] Test appeared to pass`

## Verbose Evidence Masking

MUST NOT dump raw verbose tool output into context pack.
Replace large dumps with masked references in `masked_evidence`.

### BAD (dumping raw tool output)

```
Command output:
<1000 lines of `ls -la` results>
```

### GOOD (masked with provenance)

```
masked_evidence:
- [provenance: cmd/abc1] Verified file structure exists. Full tree masked. Summary: 47 files, 12 directories.
```

### BAD (dropping validation state)

```
protected_facts:
- We chose Zettelkasten memory.
```

### GOOD (strict protected class usage)

```
protected_facts:
- [provenance: mem/5] [valid] Architecture decision: Memory stored as Zettelkasten.
```

## Persistence

After assembling Context Pack, persist to Engram:

```
mem_save(
  title: "context-pack/{project}/{change-or-session-id}",
  topic_key: "context-pack/{project}/{change-or-session-id}",
  type: "architecture",
  project: "{project}",
  content: "{full markdown Context Pack}"
)
```

Enables recovery after compaction: orchestrator retrieves via `mem_search` + `mem_get_observation`.

## Size Budget

Context Pack MUST be under 2000 tokens (~500 lines markdown). If exceeding:

1. Reduce `masked_evidence` — collapse related items
2. Remove `suppressed_items` no longer needing mention
3. Summarize `active_tasks` that are not blocking
4. Move verbose fact details into Engram, keep only topic_key reference

NEVER compromise `protected_facts` or `active_constraints` to fit budget. These are mandatory.

## Return Envelope

When called explicitly:

```markdown
**Status**: success | partial | blocked
**Summary**: Context Pack assembled with {N} protected facts, {M} masked evidence items
**Artifacts**: Engram topic_key `context-pack/{project}/{id}`
**Next**: Orchestrator should use this pack for subsequent delegations
**Risks**: {any facts that could not be validated or protected}
**Skill Resolution**: injected | fallback-registry | fallback-path | none
```

## Orchestrator Self-Correction Loop

If sub-agent returns `skill_resolution` = `fallback-registry`,
`fallback-path`, or `none`:

1. Invoke this skill immediately
2. Generate fresh Context Pack from current state
3. Re-read skill registry
4. Include Context Pack in next delegation prompt
5. Log warning: "Skill cache miss detected — reloaded registry, refreshed context pack."

Prevents silent degradation over long sessions.

## Rules

- NEVER dump raw tool output into context pack
- NEVER drop `protected_facts` to fit size budget
- NEVER invent provenance — mark unknown sources as `[provenance: unknown] [unverified]`
- ALWAYS persist context pack to Engram when available
- ALWAYS reduce character count of full history — pack same size as history failed its purpose

## Hook Detection

```bash
# .atl/hooks/on_context_pressure.sh (generated by architect-ai install)
#!/usr/bin/env bash
# Called by context-guardian when threshold reached

set -euo pipefail

PLATFORM="${ARCHITECT_PLATFORM:-unknown}"
ENGRAM_BIN="${ENGRAM_BIN:-engram}"

# Step 1: Try Engram checkpoint
"${ENGRAM_BIN}" checkpoint save --project "${ARCHITECT_PROJECT}" 2>/dev/null \
  && echo "CHECKPOINT: saved to Engram" \
  || echo "WARN: Engram checkpoint failed"

# Step 2: Platform compress
case "${PLATFORM}" in
  opencode|claude) echo "/compact" ;;
  gemini)          echo "/compress" ;;
  *)               echo "MANUAL_SUMMARY_REQUIRED" ;;
esac
```

## Anti-Patterns

- Treating this skill as optional for long sessions — auto-trigger makes it mandatory at threshold
- Rebuilding pack from scratch every turn when only one fact changed (use `mem_update` on existing pack)
- Including irrelevant historical exploration ("we tried X, didn't work") unless it informs active constraint
- Mixing suppression decisions into `masked_evidence` — different sections, different purposes
