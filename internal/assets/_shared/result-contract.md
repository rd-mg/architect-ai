# Result Contract Protocol

Every SDD phase agent MUST emit a validated JSON block as its last output to the orchestrator.

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

1. The block must be valid JSON syntax.
2. All keys listed above are required.
3. The `status` field must be one of: `completed`, `failed`, `blocked`, or `abandoned`.
4. The orchestrator validates this JSON using `.atl/scripts/validate-result-contract.sh`. If validation fails, the attempt count is incremented and the phase is retried.
