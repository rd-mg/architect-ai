---
name: _archived
description: >
  Container for archived skills that were absorbed into other skills or
  deprecated. Not a functional skill — this SKILL.md exists only so the
  _archived directory has the same structure as any other skill directory
  (TestEmbeddedAssetCount walks skills/*/ and requires SKILL.md in each).
license: MIT
metadata:
  author: rd-mg
  version: "0.0.0-placeholder"
  deprecated: true
---

# Archived Skills Container

This directory holds skills that were absorbed or deprecated. It is NOT a skill itself — the orchestrator should never match, invoke, or inject this. The frontmatter `deprecated: true` marks it as filtered by the resolver.

## Contents

Each subdirectory under `_archived/` holds the full original text of a deprecated skill, preserved for reference and rollback:

- `_archived/judgment-day/` — absorbed into `adaptive-reasoning` v1.0 as Mode 2 (adversarial-review)
- `_archived/autoreason-lite/` — absorbed into `adaptive-reasoning` v1.0 as Mode 3 (bounded-synthesis)

If more skills are deprecated in future versions, they land here.

## Why this file exists

`internal/assets/assets_test.go` contains `TestEmbeddedAssetCount` which walks every directory under `skills/` and asserts each has a `SKILL.md`. Without this stub, the test fails:

```
skill directory "_archived" missing SKILL.md: open skills/_archived/SKILL.md: file does not exist
```

The stub satisfies the structural check without being a real skill.

## Rollback

To restore any archived skill, copy its directory from `_archived/` back to `skills/`:

```bash
cp -r internal/assets/skills/_archived/judgment-day internal/assets/skills/judgment-day
```

Then re-run the skill-registry generator:

```bash
architect-ai skill-registry
```

## See also

- `_archived/README.md` — human-readable archival index
- `docs/adaptive-reasoning-v1.md` — migration from judgment-day + autoreason-lite
