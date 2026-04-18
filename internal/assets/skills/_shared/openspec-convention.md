# OpenSpec File Convention (shared across all SDD skills)

## Directory Structure

```
openspec/
├── config.yaml              <- Project-specific SDD config
├── specs/                   <- Source of truth (main specs)
│   └── {domain}/
│       └── spec.md
└── changes/                 <- Active changes
    ├── archive/             <- Completed changes (YYYY-MM-DD-{change-name}/)
    └── {change-name}/       <- Active change folder
        ├── state.yaml       <- DAG state (survives compaction)
        ├── exploration.md   <- (optional) from sdd-explore
        ├── proposal.md      <- from sdd-propose
        ├── specs/           <- from sdd-spec
        │   └── {domain}/
        │       └── spec.md  <- Delta spec
        ├── design.md        <- from sdd-design
        ├── tasks.md         <- from sdd-tasks
        ├── audit/           <- (optional) planning audit traces
        │   └── {trace-id}.md <- from audit-trail-capture
        ├── apply-progress.md <- from sdd-apply
        └── verify-report.md <- from sdd-verify
```

## Artifact File Paths

| Skill | Creates / Reads | Path |
|-------|----------------|------|
| orchestrator | Creates/Updates | `openspec/changes/{change-name}/state.yaml` |
| sdd-init | Creates | `openspec/config.yaml`, `openspec/specs/`, `openspec/changes/`, `openspec/changes/archive/` |
| sdd-explore | Creates (optional) | `openspec/changes/{change-name}/exploration.md` |
| sdd-propose | Creates | `openspec/changes/{change-name}/proposal.md` |
| sdd-spec | Creates | `openspec/changes/{change-name}/specs/{domain}/spec.md` |
| sdd-design | Creates | `openspec/changes/{change-name}/design.md` |
| sdd-tasks | Creates | `openspec/changes/{change-name}/tasks.md` |
| audit-trail-capture | Creates (optional) | `openspec/changes/{change-name}/audit/{trace-id}.md` |
| sdd-apply | Creates/Updates | `openspec/changes/{change-name}/apply-progress.md` |
| sdd-verify | Creates | `openspec/changes/{change-name}/verify-report.md` |
| sdd-archive | Moves | `openspec/changes/{change-name}/` → `openspec/changes/archive/YYYY-MM-DD-{change-name}/` |
| sdd-archive | Updates | `openspec/specs/{domain}/spec.md` (merges deltas into main specs) |

## Reading Artifacts

```
Proposal:   openspec/changes/{change-name}/proposal.md
Specs:      openspec/changes/{change-name}/specs/  (all domain subdirectories)
Design:     openspec/changes/{change-name}/design.md
Tasks:      openspec/changes/{change-name}/tasks.md
Audit:      openspec/changes/{change-name}/audit/{trace-id}.md
Apply:      openspec/changes/{change-name}/apply-progress.md
Verify:     openspec/changes/{change-name}/verify-report.md
Config:     openspec/config.yaml
Main specs: openspec/specs/{domain}/spec.md
```

## Writing Rules

- Changes are Atomic: one `openspec/changes/{name}` directory per feature/bugfix
- Audit traces are optional and policy-gated; do not create `audit/` for routine planning work
- Always create the change directory before writing artifacts
- If a file already exists, READ it first and UPDATE it (don't overwrite blindly)
- If the change directory already exists with artifacts, the change is being CONTINUED
- Use `openspec/config.yaml` `rules` section for project-specific constraints per phase

## Config File Reference

```yaml
# openspec/config.yaml
schema: spec-driven

context: |
  Tech stack: {detected}
  Architecture: {detected}
  Testing: {detected}
  Style: {detected}

rules:
  proposal:
    - Include rollback plan for risky changes
  specs:
    - Use Given/When/Then for scenarios
    - Use RFC 2119 keywords (MUST, SHALL, SHOULD, MAY)
  design:
    - Include sequence diagrams for complex flows
    - Document architecture decisions with rationale
  tasks:
    - Group by phase, use hierarchical numbering
    - Keep tasks completable in one session
  apply:
    - Follow existing code patterns
    tdd: true           # Set to true to enable RED-GREEN-REFACTOR
    test_command: ""
  verify:
    test_command: ""
    build_command: ""
    coverage_threshold: 0
  archive:
    - Warn before merging destructive deltas
```

## Archive Structure

When archiving, the change folder moves to:
```
openspec/changes/archive/YYYY-MM-DD-{change-name}/
```

Use today's date in ISO format. The archive is an AUDIT TRAIL — never delete or modify archived changes.

---

## `state.yaml` Schema (V1)

This file is the authoritative phase-level state for an active change. The
orchestrator reads it on session resume. `sdd-archive` reads it for the
archive report. **Agents must validate with** `architect-ai sdd-status
{change-name}` **after every write.**

### Canonical example

```yaml
schema_version: 1
change_name: add-user-export
created_at: 2026-04-17T14:22:00Z
updated_at: 2026-04-17T16:45:00Z
artifact_store: openspec

phases:
  sdd-explore:  { status: skipped }
  sdd-propose:  { status: completed, completed_at: 2026-04-17T14:40:00Z,
                  artifact: proposal.md, model: opus }
  sdd-spec:     { status: completed, completed_at: 2026-04-17T15:10:00Z,
                  artifacts: [specs/sale/spec.md] }
  sdd-design:   { status: in_progress, started_at: 2026-04-17T16:30:00Z,
                  model: sonnet }
  sdd-tasks:    { status: pending,    depends_on: [sdd-design] }
  sdd-apply:    { status: pending,    depends_on: [sdd-tasks] }
  sdd-verify:   { status: pending,    depends_on: [sdd-apply] }
  sdd-archive:  { status: pending,    depends_on: [sdd-verify] }

metering:           # optional, appended by sdd-archive
  total_tokens: 47120
  sessions: 3
  estimated_cost_usd: 0.28
```

### Required fields (top-level)

| Field | Type | Rule |
|-------|------|------|
| `schema_version` | int | Must equal `1`. |
| `change_name` | string | Kebab-case. Must match parent folder. |
| `created_at` | RFC 3339 UTC | Set once at change creation. |
| `updated_at` | RFC 3339 UTC | Updated on every write. `>= created_at`. |
| `artifact_store` | enum | `engram | openspec | hybrid | none`. |
| `phases` | map | Keys in the phase enum below. |

### Phase enum

`sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`,
`sdd-apply`, `sdd-verify`, `sdd-archive`.

### Status enum (per phase)

`pending`, `in_progress`, `completed`, `skipped`, `failed`.

- If `completed` → `completed_at` required.
- If `in_progress` → `started_at` required.
- If `failed` → `error` (free-form string) required.

### Atomicity

Writes MUST be atomic. Write to `state.yaml.tmp` in the same directory,
`fsync`, then `rename` to `state.yaml`. This prevents a crashed agent
from leaving a truncated file that then fails validation permanently.

### Hybrid mode authority

In `artifact_store: hybrid`, `state.yaml` on disk is the **authoritative**
record of phase status. Engram is advisory. If the two disagree, file wins.

### Validation Invariants (I1..I12)

| ID | Name | Constraint |
|----|------|------------|
| I1 | Version | `schema_version == 1` |
| I2 | Change Name | `change_name` must match parent folder name |
| I3 | Time Order | `updated_at >= created_at` |
| I4 | Store | `artifact_store` in `{engram, openspec, hybrid, none}` |
| I5 | Phase Name | Keys in `phases` must be valid SDD phase names |
| I6 | Status Enum | Status in `{pending, in_progress, completed, skipped, failed}` |
| I7 | Completeness | If `status == completed`, `completed_at` MUST be present |
| I8 | Liveness | If `status == in_progress`, `started_at` MUST be present |
| I9 | Dependency | `depends_on` phase names must exist in the `phases` map |
| I10| DAG | No cycles in the phase graph |
| I11| Atomicity | Write to `.tmp` then rename |
| I12| Authority | In hybrid mode, `state.yaml` file wins over Engram metadata |

### Validation

Before trusting the file, call `architect-ai sdd-status {change-name}`.
Non-zero exit means the file is invalid — refuse to proceed and follow
`docs/openspec-state-recovery.md`.

