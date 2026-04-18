---
openspec_delta:
  base_sha: "85f85e882cd6ec37064efe8f27fd6d4ee7e8baf89d04a2e84aeee13ea5aba87d"
  base_path: "openspec/specs/agents/spec.md"
  base_captured_at: "2026-04-18T05:29:02.584Z"
  generator: sdd-spec
  generator_version: 1
---

# Delta for Agents

## ADDED Requirements

### Requirement: Adapter Hook Visibility Contract

The `Adapter` interface MUST accurately report whether it can observe token usage from the host environment.

#### Scenario: Host Wrapping (VSCode)
- GIVEN the VSCode adapter is running within a host-wrapped environment
- WHEN `SessionHookEnabled()` is called
- THEN the result MUST be `false`

#### Scenario: Direct Observation (Claude)
- GIVEN the Claude adapter is communicating directly with its provider
- WHEN `SessionHookEnabled()` is called
- THEN the result MUST be `true`

### Requirement: Nil-Safe Metering Record

The `Record(delta any)` method of an adapter MUST handle `nil` or unexpected types without panicking.

#### Scenario: Nil Input
- GIVEN any adapter
- WHEN `Record(nil)` is called
- THEN the system MUST NOT panic

#### Scenario: Malformed Type
- GIVEN any adapter
- WHEN `Record("garbage string")` is called
- THEN the system MUST NOT panic

## MODIFIED Requirements

### Requirement: Honest Extensibility

- The `Adapter` interface documentation must explicitly state that adding a new agent requires modifications to core components (Factory, Registry, Catalog, UI).
- The claim "trivial without modifying component code" is deprecated and must be removed.
- A `MAINTAINERS_GUIDE_AGENT.md` (or equivalent section) must list the 10+ touchpoints required for a complete agent integration.
- EVERY agent implementation MUST provide a machine-readable summary of its adapter contract compliance (e.g., in `ADAPTER-CONTRACT.md`).

(Previously: Honest Extensibility requirements without explicit machine-readable summary mandate)

#### Scenario: Integration Documentation
- GIVEN a new developer adding an agent
- WHEN they read `internal/agents/interface.go`
- THEN they MUST find a warning about required touchpoints in Factory/Registry.

#### Scenario: Contract Verification
- GIVEN the `internal/agents/vscode` adapter
- WHEN auditing for compliance
- THEN it MUST have a corresponding entry in `ADAPTER-CONTRACT.md` describing its hook capabilities.

### Requirement: Antigravity Payload Extraction

The Antigravity adapter MUST correctly extract usage tokens from Google-native response payloads.

(Previously: Antigravity was not explicitly mentioned in the agent spec)

#### Scenario: Google-Native Token Extraction
- GIVEN an Antigravity response with `usageMetadata` containing `promptTokenCount` and `candidatesTokenCount`
- WHEN `Record()` is called with the raw response
- THEN the internal metering hook MUST be updated with the correct sum.
