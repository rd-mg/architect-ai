---
name: judgment-day
description: >
  DEPRECATED — absorbed into adaptive-reasoning v1.0 as Mode 2 (adversarial-review).
  Tombstone exists to keep skill registry consistent with embedded asset set;
  Go-side SkillID enum still lists judgment-day, so removing file produces
  "embedded asset not found" warning at injection time.
license: MIT
metadata:
  author: rd-mg
  version: "0.0.0-tombstone"
  deprecated: true
  superseded-by: adaptive-reasoning
---

# Judgment Day (Tombstone)

> Absorbed into `adaptive-reasoning` v1.0 as Mode 2 in V3.0. No longer a standalone reasoning protocol.

## Why this file exists

`model.SkillJudgmentDay` identifier still referenced by Go default skill registry. Missing file logs:
```
skills: skipping "judgment-day" — embedded asset not found: open skills/judgment-day/SKILL.md: file does not exist
```
Harmless but noisy. This tombstone silences it. Skill marked `deprecated: true` so resolver ranks it below active skills.

## What happened

V2 had three reasoning skills:

```
adaptive-reasoning      → classifier + router
judgment-day            → adversarial two-pass review
autoreason-lite         → bounded A/B/AB synthesis
```

V3 collapsed into ONE skill with three inline modes:

```
adaptive-reasoning v1.0 (single skill)
  ├── Mode 1: direct-exec
  ├── Mode 2: adversarial-review   ← was judgment-day
  └── Mode 3: bounded-synthesis    ← was autoreason-lite
```

No more delegation to sub-skills. All in `internal/assets/skills/adaptive-reasoning/SKILL.md`.

## How to invoke old judgment-day behavior

Orchestrator should never match this skill — `adaptive-reasoning` matched first for adversarial review. If user explicitly asks "judgment day", orchestrator routes to `adaptive-reasoning` with override:

```
Task matcher: adaptive-reasoning (Mode 2 — adversarial-review)
User alias: "judgment day", "dual review", "two-judge reasoning"
```

## Full archive

Original V2 content in `_archived/judgment-day/SKILL.md`. Do NOT ship to agents — conflicts with `adaptive-reasoning`.

## Removal

Remove this tombstone once Go-side `SkillJudgmentDay` enum deleted and default skill list regenerated:

```go
// In internal/model/skills.go — DELETE this line
SkillJudgmentDay SkillID = "judgment-day"

// In internal/installcmd/skills.go — DELETE these list members
model.SkillJudgmentDay,
```

## See also

- `adaptive-reasoning/SKILL.md` — active skill absorbing judgment-day
- `_archived/README.md` — archival index
- `docs/adaptive-reasoning-v1.md` — migration reference
