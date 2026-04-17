---
name: judgment-day
description: >
  DEPRECATED — absorbed into adaptive-reasoning v1.0 as Mode 2 (adversarial-review).
  This tombstone exists only to keep the skill registry consistent with the
  embedded asset set; the Go-side SkillID enum still lists judgment-day,
  so removing the file entirely produces a noisy "embedded asset not found"
  warning at injection time.
license: MIT
metadata:
  author: rd-mg
  version: "0.0.0-tombstone"
  deprecated: true
  superseded-by: adaptive-reasoning
---

# Judgment Day (Tombstone)

> This skill was **absorbed into `adaptive-reasoning` v1.0 as Mode 2** in V3.0. It no longer exists as a standalone reasoning protocol.

## Why this file exists

The `model.SkillJudgmentDay` identifier is still referenced by the Go code's default skill registry. When the injector tries to copy `skills/judgment-day/SKILL.md` to the user's agent and the file is missing, it logs:

```
skills: skipping "judgment-day" — embedded asset not found: open skills/judgment-day/SKILL.md: file does not exist
```

That warning is harmless but noisy. This tombstone file silences it by providing a valid (but empty-of-protocol) SKILL.md. The skill is marked `deprecated: true` in metadata so the resolver still ranks it below active skills.

## What happened

V2 had three separate reasoning skills:

```
adaptive-reasoning      → classifier + router
judgment-day            → adversarial two-pass review
autoreason-lite         → bounded A/B/AB synthesis
```

V3 collapsed these into ONE skill with three inline modes:

```
adaptive-reasoning v1.0 (single skill)
  ├── Mode 1: direct-exec
  ├── Mode 2: adversarial-review   ← was judgment-day
  └── Mode 3: bounded-synthesis    ← was autoreason-lite
```

No more delegation to sub-skills. The classifier and the three reasoning modes live in one file: `internal/assets/skills/adaptive-reasoning/SKILL.md`.

## How to invoke the old judgment-day behavior

The orchestrator should never match this skill — `adaptive-reasoning` is matched first by any task involving adversarial review. If the user explicitly asks for "judgment day", the orchestrator routes to `adaptive-reasoning` with an override:

```
Task matcher: adaptive-reasoning (Mode 2 — adversarial-review)
User alias: "judgment day", "dual review", "two-judge reasoning"
```

## Full archive

The original V2 `judgment-day/SKILL.md` content lives in `_archived/judgment-day/SKILL.md`. Do not ship it to agents — it conflicts with `adaptive-reasoning`.

## Removal

This tombstone can be removed once the Go-side `SkillJudgmentDay` enum is deleted and the default skill list is regenerated. Tracking:

```go
// In internal/model/skills.go — DELETE this line
SkillJudgmentDay SkillID = "judgment-day"

// In internal/installcmd/skills.go — DELETE these list members
model.SkillJudgmentDay,
```

After that, this tombstone is no longer needed and can be removed.

## See also

- `adaptive-reasoning/SKILL.md` — the active skill that absorbed judgment-day
- `_archived/README.md` — archival index
- `docs/adaptive-reasoning-v1.md` — migration reference
