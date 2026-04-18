# Spec Delta: Skill Registry Contract

## Requirement: Universal Representation
- The `.atl/skill-registry.md` must explicitly represent ALL runtime-relevant skill layers.
- Silent exclusion of `sdd-*`, `_shared`, and `skill-registry` is prohibited.
- If a skill is "system-managed" and not meant for direct user invocation, it must still be listed in a dedicated "System Skills" or "Shared Rules" section.

## Requirement: Entry Metadata (Schema)
- Registry entries must include a `Kind` or `Type` field.
- Supported types: `System`, `User`, `Project`, `Overlay`, `SharedRule`, `Alias`.
- The registry markdown must use distinct sections for these types.

## Requirement: Idempotency via Markers
- The registry MUST use marker-bounded sections (e.g., `<!-- architect-ai:registry:* -->`).
- Manual content added by users outside of managed sections MUST be preserved during updates.

## Requirement: Manifest-Driven Overlay Emission
- Overlays MUST provide explicit registry entries via `manifest.json`.
- Agnostic skills MUST be registered even if no specific stack version (e.g., Odoo 18.0) is detected in the project.

## Requirement: Compact Rule Resolution
- The registry remains the primary source for compact-rule injection into sub-agent prompts.
- Resolution logic must account for `SharedRule` entries that provide conventions without being invocable skills.

## Verification
- `TestSkillRegistry_IncludesSystemManagedSkills` must pass.
- `TestSkillRegistry_IncludesSharedRuleSources` must pass.
