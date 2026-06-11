# Result Contract Protocol

Every SDD phase agent MUST emit validated JSON block as last output to orchestrator.

## JSON Schema

```json
{
  "status": "completed|failed|blocked|abandoned|infrastructure_blocked|partially_completed",
  "phase": "sdd-explore",
  "change_name": "string",
  "executive_summary": "string (min 50 chars)",
  "artifacts": ["string (path or engram key)"],
  "next_recommended": "string",
  "risks": ["string"],
  "skill_resolution": {
    "status": "paths-injected|fallback-registry|fallback-path|none",
    "skills_used": ["string"],
    "fallback_reason": null
  },
  "attempt_number": 1,
  "blocked_reason": null,
  "error_type": "domain|infrastructure|none",
  "estimated_tokens": 0,
  "audit_mode": "normal|degraded"
}
```

## Validation Rules

1. Valid JSON syntax required.
2. All keys above required (including `estimated_tokens`, `attempt_number`, `error_type`).
3. `status` field must be one of:
   `completed`, `failed`, `blocked`, `abandoned`,
   `infrastructure_blocked`, `partially_completed`
4. `error_type` field must be one of: `domain`, `infrastructure`, `none`
5. `blocked_reason` is REQUIRED (non-empty string) when `status` is `blocked` or `infrastructure_blocked`.
6. `artifacts` must be non-empty when `status` is `completed` AND `phase` is one of:
   `sdd-spec`, `sdd-design`, `sdd-apply`, `sdd-verify`, `sdd-archive`
7. `next_recommended` MUST NOT be `sdd-archive` when `status` is `failed` or `blocked`.
8. `attempt_number` MUST match `circuit_breaker.attempt_counts[phase]` from `sdd-state.yaml`.
9. Orchestrator validates via `.atl/scripts/validate-result-contract.sh`.
   On domain failure: increment `attempt_counts[phase]`, retry phase.
   On infrastructure failure (`error_type: infrastructure`): increment `infra_attempt_counts[phase]`,
   retry WITHOUT incrementing `attempt_counts[phase]`.
