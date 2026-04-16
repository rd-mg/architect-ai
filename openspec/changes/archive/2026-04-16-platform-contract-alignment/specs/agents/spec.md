# Spec Delta: Agents Contract

## Requirement: Honest Extensibility
- The `Adapter` interface documentation must explicitly state that adding a new agent requires modifications to core components (Factory, Registry, Catalog, UI).
- The claim "trivial without modifying component code" is deprecated and must be removed.
- A `MAINTAINERS_GUIDE_AGENT.md` (or equivalent section) must list the 10+ touchpoints required for a complete agent integration.

## Requirement: Gemini Capabilities
- Gemini must be categorized as a `Full` tier adapter.
- Gemini must support `SubAgentCapable` interface.
- Gemini must implement native sub-agent support targeting `~/.gemini/agents/`.
- Embedded assets for Gemini sub-agents must exist in the binary.

## Requirement: Interface Separation
- Optional capabilities (Sub-agents, Workflows) must be moved into dedicated interfaces named `SubAgentCapable` and `WorkflowCapable`.
- Component logic (e.g., SDD injection) must use type assertions against these interfaces rather than hardcoded `switch` statements on `model.AgentGeminiCLI`.

## Verification
- `TestSupportedAgentsHaveCatalogAndRegistryParity` must pass.
- `TestGeminiAdapter_SubAgentsDir` must return the correct home-relative path.
