# Design: Overlay Skill-Registry Emission

## Architecture
- **Push Model**: InstallOverlay populates RegistryEntries in the manifest.
- **Marker Injection**: WriteLocalSkillRegistry uses filemerge.InjectMarkdownSection.

## Components
- **overlay.go**: RegistryEntry struct, populateRegistryEntries().
- **skill_registry.go**: Marker-aware WriteLocalSkillRegistry.
