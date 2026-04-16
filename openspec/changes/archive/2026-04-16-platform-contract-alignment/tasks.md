# Tasks: platform-contract-alignment

## Task Group 1: Core Interface & Agent Normalization
- [x] Refactor `internal/agents/interface.go`:
    - Define `SubAgentCapable` and `WorkflowCapable` interfaces.
    - Update `Adapter` interface comments for "Honest Extensibility."
- [x] Update `internal/agents/gemini/adapter.go`:
    - Implement `SubAgentCapable` interface.
    - Set Tier to `Full`.
- [x] Refactor `internal/components/sdd/inject.go`:
    - Use type assertions for `SubAgentCapable` to install native agents.
    - Remove hardcoded AgentID checks.
- [x] Add Gemini SDD assets:
    - Created agent files in `internal/assets/gemini/agents/`.
    - Included in `internal/assets/assets.go`.


## Task Group 2: Universal Skill Registry
- [x] Refactor `internal/cli/skill_registry.go`:
    - Implement layered collectors (`scanSystemSkills`, `scanSharedRules`).
    - Update `skillEntry` with `Kind` field.
    - Update `buildRegistryMarkdown` to render new sections for System/Shared layers.
- [x] Update `skill-registry` skill:
    - Add compact rules for the registry itself.

## Task Group 3: Conflict Detection & Namespaces
- [x] Refactor `internal/agentbuilder/registry.go`:
    - Add `ReservedSkillNamespace` logic.
    - Update `HasConflictWithBuiltin` to check full namespace.
- [x] Update `internal/catalog/skills.go`:
    - Ensure `Skills()` returns full set of built-in skills.

## Task Group 4: SDD Init Split & Language
- [x] Update `internal/cli/sdd_init.go`:
    - Implement `bootstrap.json` and `init-analysis.json` markers.
    - Update help text and success messages to use "Bootstrap" terminology.
- [x] Update `EnsureSDDReady` guard in `skill_registry.go`:
    - Check for bootstrap marker.

## Task Group 5: Release Metadata & URL Tightening
- [x] Create `internal/app/metadata.go`:
    - Define canonical owner/repo/docs constants.
- [x] Update `.goreleaser.yaml`:
    - Align homepage and repository URLs.
- [x] Update `internal/app/help.go`:
    - Update help command URLs to use docs constant.
- [x] Update `README.md`:
    - Normalize links.

## Task Group 1: SDD Initialization Check
- [x] Run `sdd-init` and confirm bootstrap state
- [x] Verify `.atl/` directory and core conventions

## Task Group 6: Verification & Tests
- [x] Add `internal/agents/parity_test.go` for catalog/registry/factory consistency.
- [x] Add unit tests for Gemini `SubAgentCapable` methods.
- [x] Add integration tests for registry layered scanning.
- [x] Run full test suite and generate `verify-report.md`.
