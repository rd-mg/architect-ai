---
openspec_delta:
  base_sha: "0"
  base_path: "openspec/specs/skill-registry/spec.md"
  base_captured_at: "2026-04-18T06:02:30Z"
  generator: sdd-spec
  generator_version: 1
---

# Delta for Skill Registry

## MODIFIED Requirements

### Requirement: Manifest-Driven Overlay Emission
Overlays MUST provide explicit registry entries via `manifest.json` (v19) or `manifest.yaml` (v17/v18).
Agnostic skills MUST be registered even if no specific stack version (e.g., Odoo 18.0) is detected in the project.
(Previously: Overlays MUST provide explicit registry entries via manifest.json.)

#### Scenario: Odoo DDD Skill Registration
- GIVEN an Odoo 18 overlay
- WHEN the manifest is loaded
- THEN the `patterns-ddd` skill MUST be emitted to the global registry
- AND it MUST be tagged with `Overlay` kind.

## ADDED Requirements

### Requirement: Domain-Specific Supplement Awareness
The SDD workflow MUST account for domain-specific supplements (`sdd-supplements/`) that extend standard phase behaviors for specific stacks like Odoo.
These supplements MUST explicitly reference applicable tactical skills (like `patterns-ddd`) to ensure they are loaded during design/apply phases.

#### Scenario: Supplement Cross-Reference
- GIVEN an Odoo Design Supplement (`design-odoo.md`)
- WHEN an agent executes the Design phase in an Odoo context
- THEN the supplement MUST instruct the agent to load the `patterns-ddd` skill if DDD triggers are detected.
