---
name: context-guardian
trigger: "When the agent needs to assemble context or make curation decisions for session history."
description: Context assembly contract and output template
license: MIT
metadata:
  author: rd-mg
  version: "1.0"
---

# Context Guardian

## Purpose

You are the policy surface that assembles a working context set from existing sources and makes curation decisions explainable. You do not dump the full monolithic history. Instead, you filter, mask, and prioritize facts according to bounded disclosure and semantic priority.

## Input Order

When directed to assemble context, prioritize inputs in this exact order:
1. **Active artifacts** (e.g. current specs, tasks, apply-progress)
2. **Pinned working rules** from the skill registry
3. **Semantic memory observations** (from Engram)
4. **Compacted tail** of the current session

## The Context Pack 

Your final output MUST be an explicit markdown artifact structured with the following exact stable sections:

```markdown
goal: [The current bounded objective]
active_tasks: [Tasks not yet closed]
protected_facts: [Crucial history that survives compaction]
active_constraints: [Rules and architectural limits in force]
working_rules: [Compact rules selected from the registry]
masked_evidence: [Compressed references to verbose outputs]
suppressed_items: [What was intentionally left out]
lineage: [Pointers back to sources and compacted artifacts]
```

## Protected Classes

During any compaction, the following protected classes MUST survive and be placed into `protected_facts:` or `active_constraints:`, ensuring they cannot be dropped by low-priority compaction examples:
- **architecture decisions**: Permanent changes to the structure of the software.
- **active constraints**: Rules, budgets, or boundaries currently restricting execution.
- **open tasks**: Anything assigned but not physically checked off in `tasks.md`.
- **failing-test lineage**: The traceability of a defect or failing test through attempts to fix it.

## Provenance and Validation State

Every retained or reused fact MUST carry explicit provenance and a validation state. Do not blindly copy text without marking its reliability.

Every retained fact must expose one of the following states:
- `valid`: still consistent with current state.
- `stale`: source exists but no longer reflects current reality.
- `unverified`: reused provisionally and requires confirmation.

Example: `[provenance: mem/12] [stale] - Old authentication method was JWT...`

## Verbose Evidence Masking

Verbose output from tools MUST NOT be dumped raw into the prompt context. Instead, you must replace large dumps with masked references inside `masked_evidence:`.

### Examples (Failing / Golden)

**❌ Suppressed Example (BAD - Dumping raw tool output)**
```
Command output:
<1000 lines of ls -la results>
```

**✅ Masked Evidence Example (GOOD - Masked with provenance)**
```
masked_evidence:
- [provenance: cmd/abc1] Verified file structure exists. Full tree masked.
```

**❌ Retained Example (BAD - Dropping validation state)**
```
protected_facts:
- We chose to use Zettelkasten memory.
```

**✅ Retained Example (GOOD - Strict Protected Class Usage)**
```
protected_facts:
- [provenance: mem/5] [`valid`] architecture decisions: Memory is stored as Zettelkasten.
```
