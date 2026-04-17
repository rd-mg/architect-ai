# Archived Skills

These skills have been absorbed into `internal/assets/skills/adaptive-reasoning/`
as inline reasoning modes. They are NO LONGER invoked as standalone skills or
delegated to as sub-agents.

## What Was Archived

| Archived Skill | Absorbed Into | Mode Name |
|---------------|---------------|-----------|
| `judgment-day/` | adaptive-reasoning v1.0 | Mode 2: `adversarial-review` |
| `autoreason-lite/` | adaptive-reasoning v1.0 | Mode 3: `bounded-synthesis` |

## Why

Previous versions required the orchestrator to delegate to these skills as
external sub-agents. Each delegation:
- Consumed ~1500 tokens for the fresh sub-agent context setup
- Lost the calling context (fresh sub-agent = no memory of what triggered it)
- Required skill resolution and compact rule injection
- Added latency from serialized sub-agent spawning

By absorbing the reasoning procedures into a single skill that executes
INLINE in the same context, we:
- Eliminate delegation overhead
- Preserve context continuity
- Reduce token cost per reasoning invocation by ~60%
- Simplify the mental model (one skill classifies and executes; no dispatching)

## Migration Notes

If you encounter legacy prompts or documentation referencing:

| Old Reference | New Equivalent |
|---------------|----------------|
| `Launch judgment-day sub-agent` | Apply adaptive-reasoning Mode 2 inline |
| `Delegate to autoreason-lite` | Apply adaptive-reasoning Mode 3 inline |
| `Routes: judgment-day, autoreason-lite, native-owner, ...` | Modes: adversarial-review, bounded-synthesis, direct-exec, native-sdd-first |
| `Judge A / Judge B procedure` | Pass A / Pass B in Mode 2 |
| `A/B/AB comparison` | Same, but now in Mode 3 (unchanged semantics) |

## Do NOT

- Re-activate these skills by moving them back to `internal/assets/skills/`
- Reference them in new orchestrator prompts
- Add them to the active skill registry
- Treat them as callable skills

## When to Reference

The archived content is kept for:
- Historical context (understanding the evolution of the reasoning system)
- Documentation references in the migration guide
- Rollback safety (if the v1.0 integration fails, you can temporarily restore them)

## Rollback Procedure

If adaptive-reasoning v1.0 fails in production:

```bash
# 1. Restore the archived skills to active
git mv internal/assets/skills/_archived/judgment-day \
       internal/assets/skills/judgment-day
git mv internal/assets/skills/_archived/autoreason-lite \
       internal/assets/skills/autoreason-lite

# 2. Revert adaptive-reasoning to its classifier-only version
git checkout {previous-commit} -- internal/assets/skills/adaptive-reasoning/SKILL.md

# 3. Regenerate the skill registry
architect-ai skill-registry

# 4. Commit the rollback
git commit -am "[REV][adaptive-reasoning] rollback v1.0 absorption"
```

## Questions?

See the v3 master plan at `plans/master-plan.md`, Phase 1, Steps 1.1-1.2.
