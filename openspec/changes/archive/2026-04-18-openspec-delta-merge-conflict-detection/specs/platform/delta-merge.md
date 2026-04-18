---
openspec_delta:
  base_sha: "0"
  base_path: "openspec/specs/platform/delta-merge.md"
  base_captured_at: "2026-04-18T04:47:00Z"
  generator: manual
  generator_version: 1
---

# Delta Spec: Delta-Merge Conflict Detection

## Requirements

### Requirement: Base SHA Stamping
- **ID**: CONFLICT-01
- **Statement**: Every delta spec MUST include `base_sha` in its front-matter, representing the SHA-256 of the main spec at authorship time.
- **Validation**: `sdd-archive` refuses to merge if missing or mismatching.

### Requirement: Preflight Check
- **ID**: CONFLICT-02
- **Statement**: The system MUST provide a preflight tool to dry-run the conflict check before archival.
- **Validation**: `architect-ai sdd-archive-preflight` exists and reports conflicts correctly.
