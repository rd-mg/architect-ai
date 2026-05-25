# Tasks: Phase 11 — Source of Truth Enforcement

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/11-phase-sot-enforcement.md`
> **Change:** phase-11-sot-enforcement
> **Phase:** sdd-tasks
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-tasks (architect-ai)

## Task Breakdown

- [x] **T1:** Patch `internal/assets/sdd/sdd-apply.md` — add Zero-Deviation Coder Protocol identity block with SOT Loading Step 0, First Attempt Constraint, Modification Authorization Table, and Implementation Workflow exactly as in SOT §11.2.
  - **depends_on:** []
  - **parallel_with:** [T3, T5]
  - **estimated_lines:** 80
  - **files:** `internal/assets/sdd/sdd-apply.md` (or `.agent/skills/sdd-apply/SKILL.md`)

- [x] **T2:** Create `internal/assets/_shared/traceability-injection.md` — traceability header template, phase-specific source rules, and sdd-verify validation script exactly as in SOT §11.3.
  - **depends_on:** []
  - **parallel_with:** [T1, T3]
  - **estimated_lines:** 40
  - **files:** `internal/assets/_shared/traceability-injection.md`

- [x] **T3:** Create `internal/assets/scripts/assumption-linter.sh` — full 5-rule linter (To-Be-Decided, Fix-Me, implicit assumptions, traceability headers, BDD assertions) exactly as in SOT §11.4.
  - **depends_on:** []
  - **parallel_with:** [T1, T2]
  - **estimated_lines:** 60
  - **files:** `internal/assets/scripts/assumption-linter.sh`

- [x] **T4:** Create `internal/assets/_shared/hard-stop-protocol.md` — halt conditions, BLOCKED JSON output format, --force-assume rules, clarification storage exactly as in SOT §11.6.
  - **depends_on:** []
  - **parallel_with:** [T1, T3]
  - **estimated_lines:** 40
  - **files:** `internal/assets/_shared/hard-stop-protocol.md`

- [x] **T5:** Patch `sdd-verify` — add Steps 0a (assumption linter), 0b (traceability check), 0d (prior phase validation) exactly as in SOT §11.5.
  - **depends_on:** [T3]
  - **parallel_with:** []
  - **estimated_lines:** 30
  - **files:** `.agent/skills/sdd-verify/SKILL.md`

- [x] **T6:** Create `internal/sdd/linter/installer.go` — `InstallScripts()` writing `assumption-linter.sh` and `validate-result-contract.sh` with exact script content from SOT §11.8.
  - **depends_on:** []
  - **parallel_with:** [T1, T3]
  - **estimated_lines:** 60
  - **files:** `internal/sdd/linter/installer.go`

- [x] **T7:** Create `internal/sdd/linter/installer_test.go` — `TestInstallScripts_CreatesFiles` and `TestInstallScripts_Idempotent` exactly as in SOT §11.8.
  - **depends_on:** [T6]
  - **parallel_with:** []
  - **estimated_lines:** 25
  - **files:** `internal/sdd/linter/installer_test.go`

- [ ] **T8:** Run all verification tests — zero-deviation first attempt, hard stop on missing SOT, assumption linter catches TBD, traceability header present, result contract validation, TDD mode enforcement.
  - **depends_on:** [T5, T7]
  - **parallel_with:** []
  - **estimated_lines:** 0

## Review Workload Forecast

```json
{
  "review_workload": {
    "estimated_lines_changed": 335,
    "budget_risk": "low",
    "chained_prs_recommended": false,
    "decision_needed_before_apply": false,
    "tasks_count": 8,
    "parallel_tasks": 5,
    "sequential_tasks": 3,
    "task_dependency_order": [
      {"task_id": "T1", "depends_on": [], "parallel_with": ["T3", "T5"]},
      {"task_id": "T2", "depends_on": [], "parallel_with": ["T1", "T3"]},
      {"task_id": "T3", "depends_on": [], "parallel_with": ["T1", "T2"]},
      {"task_id": "T4", "depends_on": [], "parallel_with": ["T1", "T3"]},
      {"task_id": "T6", "depends_on": [], "parallel_with": ["T1", "T3"]},
      {"task_id": "T5", "depends_on": ["T3"], "parallel_with": []},
      {"task_id": "T7", "depends_on": ["T6"], "parallel_with": []},
      {"task_id": "T8", "depends_on": ["T5", "T7"], "parallel_with": []}
    ]
  }
}
```

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
