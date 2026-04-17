# Adaptive Reasoning v1.0 — Absorption of judgment-day + autoreason-lite

**Scope**: How `adaptive-reasoning` in V3 absorbed the two prior reasoning skills as inline modes, and how to migrate.

---

## TL;DR

V2 had three separate reasoning skills:

```
adaptive-reasoning      → classifier + router
  ├── judgment-day      → adversarial two-pass review
  └── autoreason-lite   → bounded A/B/AB synthesis
```

V3 collapses these into one skill with inline modes:

```
adaptive-reasoning v1.0 (single skill)
  ├── Mode 1: direct-exec        (fast path, no second pass)
  ├── Mode 2: adversarial-review (was judgment-day)
  └── Mode 3: bounded-synthesis  (was autoreason-lite)
```

No more delegation to sub-skills. No more scattered SKILL.md files. The classifier and the three reasoning modes live in one file: `internal/assets/skills/adaptive-reasoning/SKILL.md`.

---

## Why absorb

V2 had three problems:

1. **Token waste**. Loading all three SKILL.md files cost ~4K tokens. Only one mode applies per task, so two were always wasted.
2. **Ambiguous delegation**. The classifier had to decide *which* sub-skill to call, then the sub-skill had to re-read the task context. Context re-loading doubled the prompt cost.
3. **Drift between files**. `judgment-day` and `autoreason-lite` drifted out of sync with `adaptive-reasoning`'s assumptions about its own output contract.

V3 design:

- **One classifier** at the top of `adaptive-reasoning/SKILL.md`
- **Three inline reasoning modes** below, under headers `## Mode 1`, `## Mode 2`, `## Mode 3`
- **No cross-skill calls** — the classifier outputs a mode number and the sub-agent follows the corresponding block

---

## The four observable dimensions

The classifier inspects the task across four dimensions, each 1-4:

| # | Dimension | 1 — low | 4 — high |
|---|-----------|---------|----------|
| 1 | Reversibility | Easy to undo | One-way door |
| 2 | Blast radius | Single file | Multi-system |
| 3 | Ambiguity | Spec is precise | Spec is vague |
| 4 | Defect cost | Typo-level | Security / data loss |

Sum → mode:
- **4–6**: Mode 1 (direct-exec)
- **7–11**: Mode 3 (bounded-synthesis — gives you options)
- **12–16**: Mode 2 (adversarial-review — two-pass defect hunt)

These thresholds are tunable; see `SKILL.md` for the current values.

---

## Mode 1 — direct-exec

**When**: low stakes, clear spec, small reversible change.

**Procedure**: just do it. One pass. No self-review.

**Example**: "add a log statement to the retry handler."

---

## Mode 3 — bounded-synthesis (was autoreason-lite)

**When**: medium stakes, ambiguous spec, small blast radius.

**Procedure**: produce **A**, produce **B**, produce the synthesis **AB** that captures the best of both; hand all three to the user; user picks.

**Budget**: 3 candidates max. More is analysis paralysis.

**Example**: "refactor the error handling — should we use exceptions or error-return?" → A = exceptions, B = error-return, AB = error-return with a central logger that happens to look like exceptions at the boundary.

---

## Mode 2 — adversarial-review (was judgment-day)

**When**: high stakes, large blast radius, low tolerance for defects.

**Procedure**: two passes.
1. **Pass 1**: produce the implementation.
2. **Pass 2**: adopt +++Adversarial posture, try to break your own pass 1. Enumerate attacks. Find at least 3 defects. If you can't find 3, you're not trying.
3. **Pass 3** (implied): fix the defects found in pass 2. If the fix is bigger than the original, escalate to the orchestrator — spec or design is wrong.

**Example**: "implement session expiry for the auth service."

---

## Migration from V2

### Code changes

- **`adaptive-reasoning/SKILL.md`** (REPLACE) — V3 version absorbs the three modes.
- **`judgment-day/`** (ARCHIVE) — move the directory to `_archived/judgment-day/`. Keep a stub `SKILL.md` that redirects:
  ```
  This skill was absorbed into adaptive-reasoning v1.0 as Mode 2.
  See `_archived/README.md` for rollback procedure.
  ```
- **`autoreason-lite/`** (ARCHIVE) — same pattern, now Mode 3.

### Orchestrator changes

No changes required. The orchestrator always delegates reasoning to `adaptive-reasoning`; it never knew about the sub-skills.

### Project references

Search the repo for `judgment-day` and `autoreason-lite` references in prompts and docs:
```bash
rg "judgment-day|autoreason-lite" internal/assets/
```

Replace each with `adaptive-reasoning (Mode 2)` or `adaptive-reasoning (Mode 3)` as appropriate.

---

## Rollback

If you need to restore V2:

1. Copy `_archived/judgment-day/` and `_archived/autoreason-lite/` back to `internal/assets/skills/`.
2. Replace `adaptive-reasoning/SKILL.md` with the V2 version (in git history at tag `v2.9`).
3. Regenerate skill registry: `architect-ai skill-registry`.

No data migration is required — the two V2 skills did not persist anything unique.

---

## Why this is V1 of adaptive-reasoning

Because the V2 `adaptive-reasoning` was a classifier + router — a dispatch table with no reasoning of its own. V3's `adaptive-reasoning` is the first version that actually reasons inline. The version numbering restarts.

---

## See also

- `internal/assets/skills/adaptive-reasoning/SKILL.md` — the skill itself
- `internal/assets/skills/_archived/README.md` — archival + rollback reference
- `cognitive-modes.md` — complementary discipline layer (posture ≠ reasoning mode)
