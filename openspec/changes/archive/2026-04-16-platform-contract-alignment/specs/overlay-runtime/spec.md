# Spec Delta: Overlay Runtime Contract

## Requirement: Reserved Namespace Conflict Detection
- Conflict detection for custom agents/skills must check against the `ReservedSkillNamespace`.
- The `ReservedSkillNamespace` must include:
    - All built-in MVP and non-MVP skills.
    - All SDD skills (`sdd-*`).
    - All system skills (`skill-registry`, etc.).
    - All active overlay-provided skills and aliases.
- Conflict detection must happen at the `agentbuilder` registry layer.

## Requirement: Feedback UX
- If a naming conflict is detected, the system must report the source of the conflict (Built-in, SDD, Overlay).

## Verification
- `TestResolveSkillNameConflict_ReservedNames` must pass.
- `TestResolveSkillNameConflict_OverlayAlias` must pass.
