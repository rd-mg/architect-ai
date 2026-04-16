# Design: platform-contract-alignment

## Architectural Decisions

### 1. Optional Capability Interfaces
To normalize the `Adapter` interface without bloating it, we introduce `SubAgentCapable` and `WorkflowCapable` optional interfaces in `internal/agents/interface.go`.

```go
type SubAgentCapable interface {
    SupportsSubAgents() bool
    SubAgentsDir(homeDir string) string
    EmbeddedSubAgentsDir() string
}
```

Components like `sdd.Injector` will use type assertions to detect these capabilities, replacing hardcoded switch statements. This enables non-intrusive support for Gemini's native agents.

### 2. Layered Skill Registry
The `skill-registry` generation logic will be refactored from a single flat scan into a layered collection model:
- `collectSystemSkills()`: Scans embedded/internal SDD and Registry skills.
- `collectSharedRules()`: Scans `_shared` directories for convention rule sources.
- `collectUserSkills()`: Existing behavior for global agent skills.
- `collectProjectSkills()`: Existing behavior for project-local `SKILL.md` files.
- `collectOverlaySkills()`: Existing behavior for overlay-provided assets.

Registry entries will now include an `Origin` or `Kind` field to distinguish between these layers in the generated markdown.

### 3. Namespace Conflict Detection
Conflict detection will be moved to a centralized `ReservedSkillNamespace(projectRoot string)` function. This function will union:
- `catalog.Skills()` (all built-in skills).
- `cli.SystemSkills()` (SDD and internal tools).
- `cli.OverlayManifestSkills()` (skills declared in active overlays).

The `agentbuilder` will use this expanded set to prevent collisions that `MVPSkills()` currently ignores.

### 4. SDD Init State Dual-Marker
`sdd-init` will now manage two distinct state files in `.atl/state/`:
- `bootstrap.json`: Records CLI-level tool setup completion.
- `init-analysis.json`: Records AI-level project analysis completion (stack detection, test runners, etc.).

The `EnsureSDDReady` guard will be updated to check for `bootstrap.json` at the CLI level and warn/trigger the Init phase if `init-analysis.json` is missing.

### 5. Centralized Metadata Package
We will introduce `internal/app/metadata.go` to store canonical constants:
- `OrgName = "rd-mg"`
- `RepoURL = "https://github.com/rd-mg/architect-ai"`
- `DocsURL = "https://github.com/rd-mg/architect-ai/docs"`

All CLI help text and goreleaser templates will resolve from these constants.

## Migration Strategy
- **Backward Compatibility**: The existing `Adapter` method signatures remain unchanged. 
- **Registry Update**: Existing `.atl/skill-registry.md` files will be automatically regenerated on the next `sdd-init` or `skill-registry` run to include the new sections.
- **State Migration**: If the old `.atl` directory exists without the new JSON markers, the system will perform a one-time "upkeep" to create them based on the current disk state.
