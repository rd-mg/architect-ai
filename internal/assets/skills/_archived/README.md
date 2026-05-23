# Archived Skills

These skills absorbed into `internal/assets/skills/adaptive-reasoning/` as inline reasoning modes. NO LONGER invoked as standalone skills or delegated as sub-agents.

## What Was Archived

| Archived Skill | Absorbed Into | Mode Name |
|---------------|---------------|-----------|
| `judgment-day/` | adaptive-reasoning v1.0 | Mode 2: `adversarial-review` |
| `autoreason-lite/` | adaptive-reasoning v1.0 | Mode 3: `bounded-synthesis` |

## Why

Previous versions required orchestrator to delegate to these as external sub-agents. Each delegation:
- Consumed ~1500 tokens for fresh sub-agent context setup
- Lost calling context (fresh sub-agent = no memory of what triggered it)
- Required skill resolution and compact rule injection
- Added latency from serialized sub-agent spawning

Absorbing reasoning procedures into single skill executing INLINE:
- Eliminates delegation overhead
- Preserves context continuity
- Reduces token cost per invocation by ~60%
- Simplifies mental model (one skill classifies and executes; no dispatching)

## Migration Notes

If encountering legacy prompts or documentation:

| Old Reference | New Equivalent |
|---------------|----------------|
| `Launch judgment-day sub-agent` | Apply adaptive-reasoning Mode 2 inline |
| `Delegate to autoreason-lite` | Apply adaptive-reasoning Mode 3 inline |
| `Routes: judgment-day, autoreason-lite, native-owner, ...` | Modes: adversarial-review, bounded-synthesis, direct-exec, native-sdd-first |
| `Judge A / Judge B procedure` | Pass A / Pass B in Mode 2 |
| `A/B/AB comparison` | Same, now in Mode 3 (unchanged semantics) |

## Do NOT

- Re-activate these skills by moving back to `internal/assets/skills/`
- Reference them in new orchestrator prompts
- Add them to active skill registry
- Treat them as callable skills

## When to Reference

Archived content kept for:
- Historical context (understanding evolution of reasoning system)
- Documentation references in migration guide
- Rollback safety (if v1.0 integration fails, temporarily restore)

## Rollback Procedure

If adaptive-reasoning v1.0 fails in production:

```bash
# 1. Restore archived skills to active
git mv internal/assets/skills/_archived/judgment-day \
       internal/assets/skills/judgment-day
git mv internal/assets/skills/_archived/autoreason-lite \
       internal/assets/skills/autoreason-lite

# 2. Revert adaptive-reasoning to classifier-only version
git checkout {previous-commit} -- internal/assets/skills/adaptive-reasoning/SKILL.md

# 3. Regenerate skill registry
architect-ai skill-registry

# 4. Commit rollback
git commit -am "[REV][adaptive-reasoning] rollback v1.0 absorption"
```

## Questions?

See v3 master plan at `plans/master-plan.md`, Phase 1, Steps 1.1-1.2.
