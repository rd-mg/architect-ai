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
  version: "2.0"
---

# Context Guardian v2.0

## Purpose

You are the policy surface that assembles a working context set from existing
sources and makes curation decisions explainable. You do not dump the full
monolithic history. Instead, you filter, mask, and prioritize facts according
to bounded disclosure and semantic priority.

In v2.0, the orchestrator invokes this skill AUTOMATICALLY under three
conditions (see "Auto-Trigger Rules" below). You are no longer a passive
skill that must be explicitly called.

## Cognitive Posture

This skill always operates under **+++Forensic** posture:
- Trace evidence chains
- Every claim needs provenance
- Never assume — verify
- Mark validation state per fact

## Auto-Trigger Rules

The orchestrator MUST invoke this skill when ANY of these conditions holds:

1. **Token threshold**: Estimated token usage exceeds 50% of the context
   window (heuristic: character count in conversation ≥ ~100K for a 200K
   context window).

2. **Compaction detected**: The orchestrator observes that recent context
   has been summarized or truncated (e.g., a sub-agent reports
   `skill_resolution: fallback-registry` or `fallback-path`, indicating the
   cache was lost).

3. **Explicit invocation**: User requests "compact context", "reset context",
   "what's my current state", or similar.

On trigger, the orchestrator:
1. Reads this skill
2. Assembles the Context Pack per the procedure below
3. Persists the pack to Engram (see Persistence)
4. Uses the pack as the seed for the next delegation, discarding raw history
   above the pack's lineage cutoff

## Input Order

When directed to assemble context, prioritize inputs in this exact order:

1. **Active artifacts** (current specs, tasks, apply-progress)
2. **Pinned working rules** from the skill registry (compact rules)
3. **Semantic memory observations** (from Engram via `mem_search`)
4. **Compacted tail** of the current session (last ~3 turns, full fidelity)

Earlier history is represented only through masked evidence and protected
facts — not included verbatim.

## The Context Pack

Your output MUST be a markdown artifact with these exact stable sections:

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
{Compact rules selected from the registry, keyed by skill name}
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

During any compaction, the following protected classes MUST survive and be
placed into `protected_facts` or `active_constraints`:

- **Architecture decisions**: Permanent changes to software structure
- **Active constraints**: Rules, budgets, or boundaries currently restricting execution
- **Open tasks**: Anything assigned but not checked off in `tasks.md`
- **Failing-test lineage**: Traceability of a defect or failing test through attempts to fix it
- **Security-relevant findings**: Any finding marked CRITICAL in a verification report
- **User commitments**: Promises made to the user that must be kept

These classes cannot be dropped by low-priority compaction examples.

## Provenance and Validation State

Every retained or reused fact MUST carry explicit provenance and a validation
state. Do not blindly copy text without marking its reliability.

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

Verbose output from tools MUST NOT be dumped raw into the context pack.
Replace large dumps with masked references in `masked_evidence`.

### ❌ BAD (dumping raw tool output)

```
Command output:
<1000 lines of `ls -la` results>
```

### ✅ GOOD (masked with provenance)

```
masked_evidence:
- [provenance: cmd/abc1] Verified file structure exists. Full tree masked. Summary: 47 files in 12 directories.
```

### ❌ BAD (dropping validation state)

```
protected_facts:
- We chose Zettelkasten memory.
```

### ✅ GOOD (strict protected class usage)

```
protected_facts:
- [provenance: mem/5] [valid] Architecture decision: Memory is stored as Zettelkasten.
```

## Persistence

After assembling the Context Pack, persist it to Engram:

```
mem_save(
  title: "context-pack/{project}/{change-or-session-id}",
  topic_key: "context-pack/{project}/{change-or-session-id}",
  type: "architecture",
  project: "{project}",
  content: "{full markdown Context Pack}"
)
```

This enables recovery after compaction: the orchestrator can retrieve the
pack via `mem_search` + `mem_get_observation`.

## Size Budget

The Context Pack MUST be under 2000 tokens (approximately 500 lines of
markdown). If it exceeds that:

1. Reduce `masked_evidence` — collapse related items
2. Remove `suppressed_items` that no longer need mention
3. Summarize `active_tasks` that are not blocking
4. Move verbose fact details into Engram and keep only the topic_key reference

Never compromise `protected_facts` or `active_constraints` to fit the budget.
These classes are mandatory.

## Return Envelope

When called explicitly, return:

```markdown
**Status**: success | partial | blocked
**Summary**: Context Pack assembled with {N} protected facts, {M} masked evidence items
**Artifacts**: Engram topic_key `context-pack/{project}/{id}`
**Next**: Orchestrator should use this pack for subsequent delegations
**Risks**: {any facts that could not be validated or protected}
**Skill Resolution**: injected | fallback-registry | fallback-path | none
```

## Orchestrator Self-Correction Loop

If a sub-agent returns with `skill_resolution` = `fallback-registry`,
`fallback-path`, or `none`, the orchestrator:

1. Invokes this skill immediately
2. Generates a fresh Context Pack from current state
3. Re-reads the skill registry
4. Includes the Context Pack in the next delegation prompt
5. Logs a warning: "Skill cache miss detected — reloaded registry and
   refreshed context pack."

This prevents silent degradation over long sessions.

## Rules

- NEVER dump raw tool output into the context pack
- NEVER drop a `protected_facts` item to fit the size budget
- NEVER invent provenance — if you don't know the source, mark as `[provenance: unknown] [unverified]`
- ALWAYS persist the context pack to Engram when Engram is available
- ALWAYS reduce the character count of the full history when assembling the pack — a pack that is the same size as the history has failed its purpose

## Anti-Patterns

- Treating this skill as optional for long sessions — the auto-trigger makes it mandatory at threshold
- Rebuilding the pack from scratch every turn when only one fact changed (use `mem_update` on the existing pack)
- Including irrelevant historical exploration ("we tried X, it didn't work") unless it informs an active constraint
- Mixing suppression decisions into `masked_evidence` — the two sections have different purposes
