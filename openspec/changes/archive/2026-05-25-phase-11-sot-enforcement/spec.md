# Specification: Phase 11 — Source of Truth Enforcement

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/11-phase-sot-enforcement.md`
> **Change:** phase-11-sot-enforcement
> **Phase:** sdd-spec
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-spec (architect-ai)

## Requirements

### 1. Zero-Deviation Coder Protocol (sdd-apply)
- On first attempt (attempt_number=1): implement ALL code from spec+design EXACTLY as written.
- MUST NOT analyze, re-design, improve, add abstractions, change signatures, or rename variables.
- If spec contains syntax error → implement verbatim; tests catch it; attempt 2 authorizes fix.
- Modification Authorization Table:
  - First attempt, spec clear → ❌ Transcribe exactly
  - First attempt, syntax error → ❌ Transcribe exactly
  - Tests fail (attempt 2+) → ✅ Fix specific failure only
  - Design ambiguous → ❌ HALT; emit BLOCKED
  - "Better way" known → ❌ Implement as spec; note in risks
- Source of Truth Loading (Step 0): load spec, design, tasks from Engram. Verify sdd-design and sdd-tasks both `completed`.

### 2. Traceability Auto-Injection
- Every SDD artifact MUST include traceability header after title:
  ```
  > **Source of Truth:** {path_or_engram_key}
  > **Change:** {change_name}
  > **Phase:** {phase_name}
  > **Generated:** {ISO8601}
  > **Author:** {agent} (architect-ai)
  ```
- sdd-propose: Source = user request summary
- sdd-spec: Source = proposal.md
- sdd-design: Source = spec.md
- sdd-tasks: Source = design.md

### 3. Assumption Linter (assumption-linter.sh)
- Rule 1: No "ToBeDecided" or "To Be Determined" in any .md artifact.
- Rule 2: No "FixMe" or "To-Do".
- Rule 3: Warn on implicit assumption patterns ("assuming that", "probably", "should work", etc.).
- Rule 4: All artifacts MUST have Source of Truth header.
- Rule 5: spec.md MUST have BDD assertions (Given/When/Then).
- Exit 0 = PASS, Exit 1 = VIOLATIONS FOUND.

### 4. Hard-Stop Protocol
- HALT if: source_of_truth is null, required function not in design, field conflict, task dependency not met, contradictory requirements.
- Output: BLOCKED JSON with `status`, `phase`, `executive_summary`, `blocked_reason`, `clarification_needed`, `attempt_number`.
- `--force-assume` flag: for minor ambiguities ONLY. Document assumption, mark task "assumed", flag in verify.
- NOT for: missing SOT, field conflicts, broken dependencies.

### 5. Gap Detector — sdd-verify Integration
- Step 0a: Run assumption-linter.sh. Exit 1 → BLOCKED.
- Step 0b: Check all artifacts for traceability headers. Missing → BLOCKED.
- Step 0c: Run semantic audit (from Phase 02).
- Step 0d: Validate sdd-spec, sdd-design, sdd-tasks all `completed` with artifacts recorded.

### 6. Result Contract Validation
- `validate-result-contract.sh`: validates 7 required fields (`status`, `phase`, `executive_summary`, `artifacts`, `next_recommended`, `risks`, `skill_resolution`).
- Valid statuses: `completed`, `failed`, `blocked`, `abandoned`.

## Scenarios

### Scenario 1: Zero-Deviation First Attempt
**Given** sdd-design specifies `func Authenticate(ctx context.Context, token string) (*User, error)`.
**When** sdd-apply runs with attempt_number=1.
**Then** generated code has EXACTLY that signature. No renaming, no parameter reordering.

### Scenario 2: Hard Stop on Missing SOT
**Given** sdd-apply starts but Engram has no `sdd/{change}/design` entry.
**Then** agent emits BLOCKED JSON with `clarification_needed` field.
**And** exit code 1, zero `git diff`.

### Scenario 3: Assumption Linter Catches Pending Decision
**Given** spec.md contains `## Token format\nPending Decision`.
**When** `bash .atl/scripts/assumption-linter.sh {change}` runs.
**Then** exit 1, "VIOLATION Pending Decision: spec.md".

### Scenario 4: Traceability Header Present
**Given** sdd-spec generates spec.md.
**Then** spec.md line 3 contains `> **Source of Truth:**`.
**And** sdd-verify Step 0b passes.

### Scenario 5: Result Contract Validation
**Given** incomplete result `{"status":"completed","phase":"sdd-apply"}`.
**When** piped to validate-result-contract.sh.
**Then** exit 1, "INVALID: missing fields".
**Given** complete result with all 7 fields.
**Then** exit 0, "VALID".

### Scenario 6: TDD Mode Enforced
**Given** sdd-state.yaml has `tdd_mode: true`.
**When** sdd-apply starts.
**Then** commits follow RED→GREEN→TRIANGULATE→REFACTOR pattern per task.

## Verification Criteria

| Test | Input | Expected | PASS if |
|---|---|---|---|
| Zero-deviation | Design with specific signature | Exact signature in code | rg finds exact match |
| Hard stop | Missing design in Engram | BLOCKED JSON, no files modified | Zero git diff |
| Assumption linter | spec.md with Pending Decision | Exit 1 | sdd-verify blocked |
| Traceability | sdd-spec output | SOT header present | grep matches |
| Result contract | Incomplete JSON | Exit 1 with missing fields | Validator rejects |
| TDD mode | tdd_mode=true | RED→GREEN→REFACTOR commits | git log pattern matches |

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
