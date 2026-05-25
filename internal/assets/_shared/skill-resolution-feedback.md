## Skill Resolution Feedback Protocol [All orchestrators — post-delegation]

After every sub-agent returns a Result Contract, the orchestrator MUST check
the `skill_resolution` field and act accordingly.

### skill_resolution Status Values

| Status | Meaning | Orchestrator Action |
|---|---|---|
| `paths-injected` | Skills correctly applied | Proceed normally |
| `fallback-registry` | Agent fell back to generic behavior — skill wasn't applied | Re-read skill-manifest.yaml, rebuild injection, retry phase ONCE |
| `fallback-path` | Skills loaded but from wrong path version | Re-resolve skill path, rebuild injection, retry phase ONCE |
| `none` | Agent ran without any skill injection — serious gap | Log to Engram, escalate if D5>=2, retry with explicit skill list |

### Auto-Correction Flow

```
result = sub_agent_result_contract

IF result.skill_resolution.status == "paths-injected":
  → PROCEED: skills were correctly applied
  → No action needed

ELIF result.skill_resolution.status in ["fallback-registry", "fallback-path", "none"]:
  → LOG: mem_save("skill-resolution-failure/{phase}/{timestamp}", {result, context})
  → REBUILD: re-read .atl/skill-manifest.yaml
  → REBUILD: regenerate .atl/_generated/foundation.md
  → RETRY: re-run phase once with explicit skill injection
    - Include full foundation block
    - Include all relevant tier-2 skills
    - Add explicit header: "## Skill Override — previous run missed skill injection"

  IF retry also returns non-"paths-injected":
    → If D5 >= 2: STATUS: BLOCKED, escalate to human
    → If D5 < 2:  WARN and proceed (log the gap)
```

### Skill Resolution in Result Contract

Every agent MUST include this field in their Result Contract:

```json
{
  "skill_resolution": {
    "status": "paths-injected",
    "skills_used": ["foundation", "go-testing"],
    "tier_2_activated": ["go-testing"],
    "foundation_hash": "a1b2c3d4",
    "fallback_reason": null
  }
}
```

If the agent ran without the expected skills (detected by checking if foundation block
was in the prompt):
```json
{
  "skill_resolution": {
    "status": "fallback-registry",
    "skills_used": [],
    "tier_2_activated": [],
    "foundation_hash": null,
    "fallback_reason": "Foundation block not found in context at execution time"
  }
}
```
