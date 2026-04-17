# Phase Protocol: sdd-verify

## Dependencies
- **Reads**: proposal, spec, design, tasks, apply-progress
- **Writes**: `verify-report` artifact

## Cognitive Posture
+++Adversarial — Find defects. Assume nothing is correct until proven.

## Model
sonnet — systematic validation

## Sub-Agent Launch Template

```
+++Adversarial
Try to break the artifact under review. Find the failure modes the author
missed. Assume nothing is correct until proven. Construct:
- Counterexamples that violate stated invariants
- Edge cases the happy path ignores
- Hostile inputs that exploit assumptions
- Race conditions in concurrent execution
- Upgrade paths that corrupt existing data

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

{if odoo overlay active, add:}
## Odoo Phase Context (auto-resolved)
{content of sdd-supplements/verify-odoo.md from active overlay}

## Phase: sdd-verify

Task: Validate the implementation of "{change-name}" against proposal, spec,
and design. Determine if the change meets acceptance criteria.

## Validation Procedure
1. Re-read proposal Success Criteria — each MUST be verifiable
2. Re-read spec capabilities — each MUST have been implemented
3. Re-read design — check implementation matches
4. Check tasks.md — all tasks marked [x] or blocked with reason
5. Read apply-progress — verify no open blockers
6. Run test suite if available — report pass/fail per test
7. Apply adversarial-review (from adaptive-reasoning Mode 2):
   - Pass A: happy-path correctness
   - Pass B: failure-mode lens
8. Classify findings: CRITICAL / WARNING (real) / WARNING (theoretical) / SUGGESTION
9. Output verdict: APPROVED / CONDITIONALLY APPROVED / NEEDS CHANGES / UNRESOLVED

## Deterministic Checks (in addition to adversarial review)
- [ ] All success criteria from proposal are verifiable
- [ ] All spec capabilities have implementation
- [ ] All tasks marked [x] or with documented block reason
- [ ] No TODO / FIXME / XXX in changed code
- [ ] Tests exist for each capability
- [ ] Test runner passes (if available)

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/verify-report",
  topic_key: "sdd/{change-name}/verify-report",
  type: "verification-report",
  project: "{project}",
  content: "{your verify-report markdown with verdict}"
)

## Size Budget: 700 words max

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Verdict field is AUTHORITATIVE for orchestrator decision:
  - `APPROVED` → next recommended `sdd-archive`
  - `CONDITIONALLY APPROVED` → present to user for manual decision
  - `NEEDS CHANGES` → back to `sdd-apply` with new batch
  - `UNRESOLVED` → escalate to user
- Update state: `applying` → `verified` (or `verify-failed`)

## Failure Handling

- Never treat `APPROVED` as merge permission without human sign-off
- If adversarial review reveals CRITICAL finding → override to `NEEDS CHANGES`
- If deterministic checks fail → return `blocked` with specific failure list
- If test runner unavailable → flag as RISK in report, do NOT claim `APPROVED`
