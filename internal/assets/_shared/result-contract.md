# Result Contract Protocol

Every SDD phase agent MUST emit validated JSON block as last output to orchestrator.

## JSON Schema

```json
{
  "status": "completed|failed|blocked|abandoned",
  "phase": "sdd-explore",
  "change_name": "string",
  "executive_summary": "string",
  "artifacts": ["string"],
  "next_recommended": "string",
  "risks": ["string"],
  "skill_resolution": {
    "status": "paths-injected|fallback-registry|fallback-path|none",
    "skills_used": ["string"],
    "fallback_reason": null
  },
  "attempt_number": 1,
  "blocked_reason": null
}
```

## Validation Rules

1. Valid JSON syntax required.
2. All keys above required.
3. `status` field must be one of: `completed`, `failed`, `blocked`, `abandoned`.
4. Orchestrator validates via `.atl/scripts/validate-result-contract.sh`. On failure: increment attempt, retry phase.
