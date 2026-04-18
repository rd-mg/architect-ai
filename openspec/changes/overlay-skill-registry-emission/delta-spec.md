# Delta Spec: Overlay Skill-Registry Emission

## Requirements
- Overlays MUST specify skills in manifest.json RegistryEntries.
- Installation MUST update .atl/skill-registry.md idempotently.
- Uninstallation MUST remove overlay sections.
- Manual user content in the registry MUST be preserved.
- Agnostic skills MUST be registered even if no stack version is detected.

## Scenarios
### Scenario 1: Fresh Odoo Install
Given a repo with no __manifest__.py
When I install odoo-development-skill
Then patterns-agnostic should appear in the registry.

### Scenario 2: Preservation
Given a registry with manual ## Notes
When I update skills
Then ## Notes should remain.
