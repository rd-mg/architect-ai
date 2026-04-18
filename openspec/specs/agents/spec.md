# Spec: Agents

## Requirement: Honest Extensibility
- The `Adapter` interface documentation must explicitly state that adding a new agent requires modifications to core components (Factory, Registry, Catalog, UI).
- The claim "trivial without modifying component code" is deprecated and must be removed.
- A `MAINTAINERS_GUIDE_AGENT.md` (or equivalent section) must list the 10+ touchpoints required for a complete agent integration.
- EVERY agent implementation MUST provide a machine-readable summary of its adapter contract compliance (e.g., in `ADAPTER-CONTRACT.md`).

## Requirement: Gemini Capabilities
- Gemini must be categorized as a `Full` tier adapter.
- Gemini must support `SubAgentCapable` interface.
- Gemini must implement native sub-agent support targeting `~/.gemini/agents/`.
- Embedded assets for Gemini sub-agents must exist in the binary.

## Requirement: Interface Separation
- Optional capabilities (Sub-agents, Workflows) must be moved into dedicated interfaces named `SubAgentCapable` and `WorkflowCapable`.
- Component logic (e.g., SDD injection) must use type assertions against these interfaces rather than hardcoded `switch` statements on `model.AgentGeminiCLI`.

## Requirement: Adapter Hook Visibility Contract
The `Adapter` interface MUST accurately report whether it can observe token usage from the host environment.

#### Scenario: Host Wrapping (VSCode)
- GIVEN the VSCode adapter is running within a host-wrapped environment
- WHEN `SessionHookEnabled()` is called
- THEN the result MUST be `false`

#### Scenario: Direct Observation (Claude)
- GIVEN the Claude adapter is communicating directly with its provider
- WHEN `SessionHookEnabled()` is called
- THEN the result MUST be `true`

## Requirement: Nil-Safe Metering Record
The `Record(delta any)` method of an adapter MUST handle `nil` or unexpected types without panicking.

#### Scenario: Nil Input
- GIVEN any adapter
- WHEN `Record(nil)` is called
- THEN the system MUST NOT panic

#### Scenario: Malformed Type
- GIVEN any adapter
- WHEN `Record("garbage string")` is called
- THEN the system MUST NOT panic

## Requirement: Antigravity Payload Extraction
The Antigravity adapter MUST correctly extract usage tokens from Google-native response payloads.

#### Scenario: Google-Native Token Extraction
- GIVEN an Antigravity response with `usageMetadata` containing `promptTokenCount` and `candidatesTokenCount`
- WHEN `Record()` is called with the raw response
- THEN the internal metering hook MUST be updated with the correct sum.

## Verification
- `TestSupportedAgentsHaveCatalogAndRegistryParity` must pass.
- `TestGeminiAdapter_SubAgentsDir` must return the correct home-relative path.
- All `SessionHookEnabled` and `RecordResponse` safety tests MUST pass across all adapters.
