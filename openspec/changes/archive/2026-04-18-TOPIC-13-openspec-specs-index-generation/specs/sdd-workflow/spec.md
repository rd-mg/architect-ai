---
openspec_delta:
  base_sha: "0"
  base_path: "openspec/specs/sdd-workflow/spec.md"
  base_captured_at: "2026-04-18T06:36:56.935Z"
  generator: sdd-spec
  generator_version: 1
---

# SDD Workflow Specification

## Purpose

This specification defines the core workflows for Spec-Driven Development (SDD), ensuring consistency in how changes are explored, proposed, implemented, and archived.

## Requirements

### Requirement: INDEX Auto-Generation

After a successful change archive, the system MUST automatically regenerate the OpenSpec INDEX file. This index SHALL serve as a high-performance lookup for existing capabilities.

#### Scenario: Successful INDEX Regeneration

- GIVEN a set of domain directories in `openspec/specs/` each containing a `spec.md`
- WHEN the `sdd-archive` phase completes successfully
- THEN a new `openspec/specs/INDEX.md` MUST be created or overwritten
- AND it MUST contain a table mapping Domain, Title (first heading of spec), and Path

### Requirement: INDEX Consultation

During the exploration and proposal phases, the system MUST consult the OpenSpec INDEX if it exists. This allows agents to reuse existing capability definitions instead of duplicating them.

#### Scenario: Index Lookup during Proposal

- GIVEN an existing `openspec/specs/INDEX.md`
- WHEN the `sdd-propose` phase starts
- THEN the system MUST read `INDEX.md` to identify relevant existing domains
- AND it SHOULD use these domains to correctly categorize "Modified Capabilities"
